package memory

import (
	"context"
	"sync"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// NewUserRepositoryMemory returns new initialized user repository that keeps data in memory
func NewUserRepositoryMemory() repository.UserRepository {
	return &userRepositoryMemory{}
}

type userRepositoryMemory struct {
	users []user.User
	mu    sync.Mutex
}

func (r *userRepositoryMemory) AddUserList(_ context.Context, userList user.List) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range userList {
		r.users = append(r.users, u)
	}

	return nil
}

func (r *userRepositoryMemory) GetUserInChannel(_ context.Context, channelID, userID string) (user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.ChannelID == channelID && u.UserID == userID {
			return u, nil
		}
	}

	return user.User{}, repository.ErrNotFound
}

func (r *userRepositoryMemory) Truncate(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users = nil

	return nil
}
