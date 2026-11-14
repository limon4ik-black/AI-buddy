package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	chatsvc "github.com/limon4ik-black/AI-buddy/chat-service/chat"
	info "github.com/limon4ik-black/AI-buddy/employee-service/employee"
	"github.com/limon4ik-black/AI-buddy/shared/platform/idgen"
)

// Small pools expose deadlocks caused by querying outside a transaction while holding its connection.
func TestConcurrentIdempotency(t *testing.T) {
	h := setup(t)
	req(t, h.client, h.gateway, "GET", "/api/v1/profile", h.token, nil, 200)
	post := func(path string, body any, extra map[string]string) ([]byte, int, error) {
		b, _ := json.Marshal(body)
		r, _ := http.NewRequest("POST", h.gateway+path, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+h.token)
		for k, v := range extra {
			r.Header.Set(k, v)
		}
		res, e := h.client.Do(r)
		if e != nil {
			return nil, 0, e
		}
		defer res.Body.Close()
		data, e := io.ReadAll(res.Body)
		return data, res.StatusCode, e
	}
	var wg sync.WaitGroup
	results := make(chan string, 4)
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, status, e := post("/api/v1/access-requests", obj("system", "stash", "confirmed", true), map[string]string{"Idempotency-Key": "concurrent-access-key"})
			if e != nil || status != 200 && status != 201 {
				results <- fmt.Sprintf("failure: %d %v", status, e)
				return
			}
			var v info.AccessRequest
			_ = json.Unmarshal(data, &v)
			results <- v.ID
		}()
	}
	wg.Wait()
	close(results)
	id := ""
	for result := range results {
		if strings.HasPrefix(result, "failure") {
			t.Fatal(result)
		}
		if id != "" && id != result {
			t.Fatal("Duplicate access request")
		}
		id = result
	}
	chat := unpack[chatsvc.Chat](t, req(t, h.client, h.gateway, "POST", "/api/v1/chats", h.token, obj("title", "Concurrent"), 201))
	path := "/api/v1/chats/" + chat.ID
	msg := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages", h.token, obj("content", "Привет", "clientMessageId", idgen.New()), 201))
	results = make(chan string, 4)
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, status, e := post(path+"/messages/"+msg.ID+"/answer", nil, nil)
			if e != nil || status != 200 && status != 201 {
				results <- fmt.Sprintf("failure: %d %v", status, e)
				return
			}
			var v chatsvc.Message
			_ = json.Unmarshal(data, &v)
			results <- v.ID
		}()
	}
	wg.Wait()
	close(results)
	id = ""
	for result := range results {
		if strings.HasPrefix(result, "failure") {
			t.Fatal(result)
		}
		if id != "" && id != result {
			t.Fatal("Duplicate answer")
		}
		id = result
	}
}
