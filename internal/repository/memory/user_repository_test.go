package memory

import (
	"context"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryMemory_AddingAndGettingUsers(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepositoryMemory()

	channelID := "c5bea8d9-1d90-4d90-a445-e6ce74dff4cc"
	channel2ID := "8b6353c3-46ca-485d-87c3-66bc36c70d88"

	u1 := user.User{
		ChannelID: channelID,
		UserID:    "c8d1b9fb-35f1-46cb-aa37-a16b96937734",
		Email:     "first@user.com",
	}
	u2 := user.User{
		ChannelID: channelID,
		UserID:    "b599fdbe-09df-47f9-9b08-c08caccab3b1",
		Email:     "second@user.com",
	}
	u3 := user.User{
		ChannelID: channel2ID,
		UserID:    "bb3f1241-6f52-4227-92fc-949385895cd5",
		Email:     "third@user.com",
	}

	list := user.List{u1, u2, u3}

	err := repo.AddUserList(ctx, list)
	require.NoError(t, err)

	notFoundUser, err := repo.GetUserInChannel(ctx, channelID, "nonexistentID")
	require.Equal(t, notFoundUser, user.User{})
	require.ErrorIs(t, err, repository.ErrNotFound)

	notFoundUser, err = repo.GetUserInChannel(ctx, channel2ID, u2.UserID)
	require.Equal(t, notFoundUser, user.User{})
	require.ErrorIs(t, err, repository.ErrNotFound)

	notFoundUser, err = repo.GetUserInChannel(ctx, channelID, u3.UserID)
	require.Equal(t, notFoundUser, user.User{})
	require.ErrorIs(t, err, repository.ErrNotFound)

	retUser, err := repo.GetUserInChannel(ctx, channelID, u2.UserID)
	require.NoError(t, err)

	assert.Equal(t, u2, retUser)
}
