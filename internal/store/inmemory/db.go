package inmemory

import (
	"context"
	"fmt"
	"midProject/internal/models"
	"midProject/internal/store"
	"sync"
)

type DB struct {
	data map[int]*models.User

	mu *sync.RWMutex
}

func NewDB() store.Store {
	return &DB{
		data: make(map[int]*models.User),
		mu:   new(sync.RWMutex),
	}
}

func (db *DB) Create(ctx context.Context, user *models.User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data[user.ID] = user
	return nil
}

func (db *DB) All(ctx context.Context) ([]*models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	neons := make([]*models.User, 0, len(db.data))
	for _, neon := range db.data {
		neons = append(neons, neon)
	}

	return neons, nil
}

func (db *DB) ByID(ctx context.Context, id int) (*models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, ok := db.data[id]
	if !ok {
		return nil, fmt.Errorf("no user with id %d", id)
	}

	return user, nil
}

func (db *DB) Update(ctx context.Context, user *models.User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data[user.ID] = user
	return nil
}

func (db *DB) Delete(ctx context.Context, id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.data, id)
	return nil
}
