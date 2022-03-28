package memory

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// NewUserRepositoryMemory returns new initialized user repository that keeps data in memory
func NewUserRepositoryMemory() repository.UserRepository {
	return &userRepositoryMemory{}
}

type userRepositoryMemory struct {
	users []user.User
}

func (r *userRepositoryMemory) AddUserList(_ context.Context, userList user.List) error {
	for _, u := range userList {
		r.users = append(r.users, u)
	}

	return nil
}

func (r userRepositoryMemory) GetUsersByChannel(_ context.Context, channelID string) (user.List, error) {
	var list user.List

	for _, u := range r.users {
		if u.ChannelID == channelID {
			list = append(list, u)
		}
	}

	return list, nil
}
