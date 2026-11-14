package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// SendWebSocket uses exactly the same persistence, authorization and generation
// path as REST, so both transports share sources, actions and reply records.
func (c *Service) SendWebSocket(r *http.Request, chatID uuid.UUID, content string) (string, error) {
	request := r.Clone(r.Context())
	request.SetPathValue("id", chatID.String())
	body, _ := json.Marshal(map[string]string{"content": content, "clientMessageId": uuid.NewString()})
	request.Body = io.NopCloser(bytes.NewReader(body))
	request.Header = request.Header.Clone()
	request.Header.Set("Content-Type", "application/json")
	message, err := c.saveMessage(request)
	if err != nil {
		return "", err
	}
	request.SetPathValue("messageId", message.ID)
	answer, _, err := c.generate(request)
	return answer.Content, err
}
