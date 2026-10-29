package user

import (
	"context"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	ctx    *sync.RWMutex
	data   map[string]User
	nextID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]User),
	}
}

func (m *MemoryStore) Create(_ context.Context, u User) (User, error) {
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

func (m *MemoryStore) GetByMail(_ context.Context, email string) (User, error) {
	m.ctx.RLock()
	defer m.ctx.RUnlock()

	user, ok := m.data[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}
