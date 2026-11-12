package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
)

type Handler struct {
	Service *chat.Service
	Timeout time.Duration
	Origins []string
	active  sync.WaitGroup
}
type event struct {
	Type    string `json:"type"`
	ChatID  string `json:"chat_id"`
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	Message string `json:"message,omitempty"`
}

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func publicError(err error) (int, string) {
	switch {
	case errors.Is(err, chat.ErrNotFound):
		return 404, "Чат не найден"
	case errors.Is(err, chat.ErrMessage):
		return 400, "Сообщение должно содержать от 1 до 8000 символов"
	case errors.Is(err, chat.ErrAI):
		return 502, "Не удалось получить ответ AI"
	default:
		return 503, "Сервис временно недоступен"
	}
}
func fail(w http.ResponseWriter, err error) {
	status, message := publicError(err)
	respond(w, status, map[string]string{"message": message})
}
func parseID(raw string) (uuid.UUID, error) {
	// Require canonical UUIDs rather than accepting other UUID encodings.
	if len(raw) != 36 {
		return uuid.Nil, errors.New("invalid UUID")
	}
	return uuid.Parse(raw)
}
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /chats", h.create)
	mux.HandleFunc("GET /chats", h.list)
	mux.HandleFunc("GET /chats/{id}/messages", h.history)
	mux.HandleFunc("DELETE /chats/{id}", h.delete)
	mux.HandleFunc("GET /ws", h.ws)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := parseID(r.Header.Get("X-User-ID"))
		if err != nil {
			respond(w, 401, map[string]string{"message": "Отсутствует или некорректен X-User-ID"})
			return
		}
		mux.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

type userKey struct{}

func user(r *http.Request) uuid.UUID { return r.Context().Value(userKey{}).(uuid.UUID) }
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
	defer cancel()
	c, err := h.Service.Create(ctx, user(r))
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 201, c)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
	defer cancel()
	c, err := h.Service.List(ctx, user(r))
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, c)
}
func chatID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		respond(w, 400, map[string]string{"message": "Некорректный UUID чата"})
		return uuid.Nil, false
	}
	return id, true
}
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	id, ok := chatID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
	defer cancel()
	messages, err := h.Service.History(ctx, user(r), id)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, messages)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := chatID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
	defer cancel()
	if err := h.Service.Delete(ctx, user(r), id); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) Wait() { h.active.Wait() }
func (h *Handler) ws(w http.ResponseWriter, r *http.Request) {
	h.active.Add(1)
	defer h.active.Done()
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.Origins})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(64 << 10)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	type input struct {
		kind websocket.MessageType
		data []byte
	}
	incoming := make(chan input)
	// Keep reading while AI is running so a peer disconnect cancels its HTTP request.
	// No generation runs in this goroutine; processing below remains sequential.
	go func() {
		defer cancel()
		defer close(incoming)
		for {
			kind, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			select {
			case incoming <- input{kind, data}:
			case <-ctx.Done():
				return
			}
		}
	}()
	write := func(e event) error {
		writeCtx, done := context.WithTimeout(ctx, 10*time.Second)
		defer done()
		data, _ := json.Marshal(e)
		return conn.Write(writeCtx, websocket.MessageText, data)
	}
	for in := range incoming {
		var body struct {
			ChatID  string `json:"chat_id"`
			Content string `json:"content"`
		}
		dec := json.NewDecoder(strings.NewReader(string(in.data)))
		dec.DisallowUnknownFields()
		if in.kind != websocket.MessageText || dec.Decode(&body) != nil || !json.Valid(in.data) {
			if write(event{Type: "error", Message: "Некорректное сообщение JSON"}) != nil {
				return
			}
			continue
		}
		id, err := parseID(body.ChatID)
		if err != nil {
			if write(event{Type: "error", ChatID: body.ChatID, Message: "Некорректный UUID чата"}) != nil {
				return
			}
			continue
		}
		opCtx, done := context.WithTimeout(ctx, h.Timeout)
		answer, err := h.Service.Send(opCtx, user(r), id, body.Content)
		done()
		if err != nil {
			_, message := publicError(err)
			if write(event{Type: "error", ChatID: body.ChatID, Message: message}) != nil {
				return
			}
			continue
		}
		if write(event{Type: "message", ChatID: body.ChatID, Role: "assistant", Content: answer}) != nil {
			return
		}
	}
}
