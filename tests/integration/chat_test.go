package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	chatsvc "github.com/limon4ik-black/AI-buddy/chat-service/chat"
	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/idgen"
)

func TestChatFlowAndIsolation(t *testing.T) {
	h := setup(t)
	base := "/api/v1/chats"
	req(t, h.client, h.gateway, "GET", base, h.token, nil, 200)
	req(t, h.client, h.gateway, "POST", base, h.token, obj("title", ""), 400)
	chat := unpack[chatsvc.Chat](t, req(t, h.client, h.gateway, "POST", base, h.token, obj("title", "Первые шаги"), 201))
	path := base + "/" + chat.ID
	req(t, h.client, h.gateway, "GET", path, h.token, nil, 200)
	req(t, h.client, h.gateway, "PATCH", path, h.token, obj("title", "Новый заголовок"), 200)
	req(t, h.client, h.gateway, "PATCH", path, h.token, obj("title", ""), 400)
	other := h.keys.Sign("other", "other@example.test")
	for _, method := range []string{"GET", "DELETE", "PATCH"} {
		var body any
		if method == "PATCH" {
			body = obj("title", "Кража")
		}
		req(t, h.client, h.gateway, method, path, other, body, 404)
	}
	for _, method := range []string{"GET", "POST"} {
		req(t, h.client, h.gateway, method, path+"/messages", other, obj("content", "secret", "clientMessageId", idgen.New()), 404)
	}
	req(t, h.client, h.gateway, "POST", path+"/messages", h.token, obj("content", "", "clientMessageId", idgen.New()), 400)
	content := obj("content", "Как получить доступ к Stash?", "clientMessageId", idgen.New())
	msg := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages", h.token, content, 201))
	duplicate := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages", h.token, content, 201))
	if msg.ID != duplicate.ID {
		t.Fatal("Duplicate message")
	}
	content["content"] = "different"
	req(t, h.client, h.gateway, "POST", path+"/messages", h.token, content, 409)
	answerPath := path + "/messages/" + msg.ID + "/answer"
	req(t, h.client, h.gateway, "POST", answerPath, other, nil, 404)
	answer := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", answerPath, h.token, nil, 201))
	if answer.SuggestedAction == nil || answer.Mode != "mock" || len(answer.Sources) == 0 {
		t.Fatal(answer)
	}
	same := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", answerPath, h.token, nil, 200))
	if same.ID != answer.ID {
		t.Fatal("Duplicate generation")
	}
	req(t, h.client, h.gateway, "POST", path+"/messages/nope/answer", h.token, nil, 404)
	page := unpack[struct{ Items []chatsvc.Message }](t, req(t, h.client, h.gateway, "GET", path+"/messages", h.token, nil, 200))
	if len(page.Items) != 2 {
		t.Fatal(page)
	}
	req(t, h.client, h.gateway, "GET", path+"/messages?after=-1", h.token, nil, 400)
	req(t, h.client, h.gateway, "GET", path+"/messages?after=bad", h.token, nil, 400)
	// Pagination remains stable when additional messages are appended.
	for n := 0; n < 100; n++ {
		_, e := h.chat.DB.Exec(context.Background(), "INSERT INTO messages(chat_id,role,content) VALUES($1,'user','page test')", chat.ID)
		if e != nil {
			t.Fatal(e)
		}
	}
	page2 := unpack[struct {
		Items      []chatsvc.Message
		NextCursor *int64
	}](t, req(t, h.client, h.gateway, "GET", path+"/messages", h.token, nil, 200))
	if len(page2.Items) != 100 || page2.NextCursor == nil {
		t.Fatal("Pagination failed")
	}
	req(t, h.client, h.gateway, "GET", fmt.Sprintf("%s/messages?after=%d", path, *page2.NextCursor), h.token, nil, 200)
	req(t, h.client, h.gateway, "DELETE", path, h.token, nil, 204)
	req(t, h.client, h.gateway, "DELETE", path, h.token, nil, 404)
}

func TestGenerationFailuresAndRetry(t *testing.T) {
	h := setup(t)
	chat := unpack[chatsvc.Chat](t, req(t, h.client, h.gateway, "POST", "/api/v1/chats", h.token, obj("title", "Failure"), 201))
	path := "/api/v1/chats/" + chat.ID
	msg := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages", h.token, obj("content", "Какие задачи остались?", "clientMessageId", idgen.New()), 201))
	answerPath := path + "/messages/" + msg.ID + "/answer"
	h.chat.NoteURL = "http://127.0.0.1:1"
	req(t, h.client, h.gateway, "POST", answerPath, h.token, nil, 502)
	h.chat.NoteURL = h.infoURL
	h.chat.RAGURL = "http://127.0.0.1:1"
	req(t, h.client, h.gateway, "POST", answerPath, h.token, nil, 502)
	for _, output := range []contracts.GenerateOutput{{Answer: "", Mode: "rag"}, {Answer: "Secret", Mode: "rag", Sources: []contracts.Source{{DocumentID: "analytics-private"}}}, {Answer: "a", Mode: "rag", SuggestedAction: &contracts.Action{Type: "DELETE_ALL", Label: "bad", System: "stash"}}, {Answer: "a", Mode: "rag", SuggestedAction: &contracts.Action{Type: "REQUEST_ACCESS", Label: "bad", System: "secret-system"}}} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { httpapi.JSON(w, 200, output) }))
		h.chat.RAGURL = s.URL
		req(t, h.client, h.gateway, "POST", answerPath, h.token, nil, 502)
		s.Close()
	}
	h.chat.RAGURL = h.ragURL
	req(t, h.client, h.gateway, "POST", answerPath, h.token, nil, 201)
}
