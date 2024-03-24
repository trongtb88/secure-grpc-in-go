package stores

import (
	"github.com/trongtb88/secure-grpc-in-go/stores/models"
	"sync"
)

// UserStore is an interface to store users
type UserStore interface {
	// Save saves a user to the store
	Save(user *models.User) error
	// Find finds a user by username
	Find(username string) (*models.User, error)
}

// InMemoryUserStore stores users in memory
type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*models.User
}

// NewInMemoryUserStore returns a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*models.User),
	}
}

// Save saves a user to the store
func (store *InMemoryUserStore) Save(user *models.User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.users[user.Username] != nil {
		return ErrAlreadyExists
	}

	store.users[user.Username] = user.Clone()
	return nil
}

// Find finds a user by username
func (store *InMemoryUserStore) Find(username string) (*models.User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.users[username]
	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}
