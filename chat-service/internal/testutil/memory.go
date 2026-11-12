// Package testutil supplies a deterministic repository for service and transport tests.
package testutil

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
)

type Memory struct {
	mu        sync.Mutex
	owners    map[uuid.UUID]uuid.UUID
	summaries map[uuid.UUID]chat.Summary
	messages  map[uuid.UUID][]chat.Message
	next      int64
	Failure   error
}

func New() *Memory {
	return &Memory{owners: make(map[uuid.UUID]uuid.UUID), summaries: make(map[uuid.UUID]chat.Summary), messages: make(map[uuid.UUID][]chat.Message)}
}
func (m *Memory) Create(ctx context.Context, u uuid.UUID) (chat.Created, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Failure != nil {
		return chat.Created{}, m.Failure
	}
	id := uuid.New()
	now := time.Now()
	m.owners[id] = u
	m.summaries[id] = chat.Summary{ID: id, Title: "Новый чат", UpdatedAt: now}
	m.messages[id] = []chat.Message{}
	return chat.Created{ID: id, Title: "Новый чат", CreatedAt: now}, nil
}
func (m *Memory) List(ctx context.Context, u uuid.UUID) ([]chat.Summary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Failure != nil {
		return nil, m.Failure
	}
	out := []chat.Summary{}
	for id, owner := range m.owners {
		if owner == u {
			out = append(out, m.summaries[id])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (m *Memory) History(ctx context.Context, u, id uuid.UUID) ([]chat.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Failure != nil {
		return nil, m.Failure
	}
	if m.owners[id] != u {
		return nil, chat.ErrNotFound
	}
	return append([]chat.Message{}, m.messages[id]...), nil
}
func (m *Memory) Delete(ctx context.Context, u, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Failure != nil {
		return m.Failure
	}
	if m.owners[id] != u {
		return chat.ErrNotFound
	}
	delete(m.owners, id)
	delete(m.messages, id)
	delete(m.summaries, id)
	return nil
}
func (m *Memory) Append(ctx context.Context, u, id uuid.UUID, role, content string) (chat.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Failure != nil {
		return chat.Message{}, m.Failure
	}
	if m.owners[id] != u {
		return chat.Message{}, chat.ErrNotFound
	}
	m.next++
	msg := chat.Message{ID: m.next, Role: role, Content: content, CreatedAt: time.Now()}
	m.messages[id] = append(m.messages[id], msg)
	summary := m.summaries[id]
	summary.UpdatedAt = time.Now()
	m.summaries[id] = summary
	return msg, nil
}

type AIFunc func(context.Context, string, []chat.Turn) (string, error)

func (f AIFunc) Answer(ctx context.Context, s string, h []chat.Turn) (string, error) {
	return f(ctx, s, h)
}
