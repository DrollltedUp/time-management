package auth

import (
	"context"
	"strings"
	"sync"
	"time"
)

type MemotyStore struct {
	ctx    *sync.RWMutex
	data   map[string]User
	nextID int
}

func NewMemoryStore() *MemotyStore {
	return &MemotyStore{
		data: make(map[string]User),
	}
}

func (m *MemotyStore) Create(_ context.Context, u User) (User, error) {
	m.ctx.Lock()
	defer m.ctx.Unlock()

	key := strings.ToLower(u.Email)
	if _, ok := m.data[key]; ok {
		return User{}, ErrEmailTaken
	}

	// Задать остальные данные

	m.nextID++
	u.ID = int64(m.nextID)
	u.Created_at = time.Now()

	m.data[key] = u
	return u, nil
}
