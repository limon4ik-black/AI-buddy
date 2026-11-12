package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
)

type Client struct {
	URL, Token string
	HTTP       *http.Client
}

func (c *Client) Answer(ctx context.Context, message string, history []chat.Turn) (string, error) {
	body, err := json.Marshal(struct {
		Message string      `json:"message"`
		History []chat.Turn `json:"history"`
	}{message, history})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.URL, "/")+"/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", errors.New("AI request failed")
	}
	var result struct {
		Answer *string `json:"answer"`
	}
	dec := json.NewDecoder(io.LimitReader(res.Body, 4<<20))
	if err = dec.Decode(&result); err != nil {
		return "", err
	}
	if err = dec.Decode(new(any)); err != io.EOF {
		return "", errors.New("invalid AI response")
	}
	if result.Answer == nil || strings.TrimSpace(*result.Answer) == "" {
		return "", errors.New("missing AI answer")
	}
	return *result.Answer, nil
}
