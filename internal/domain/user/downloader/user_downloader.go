package userdownloader

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// UserDownloader downloads list of users from the ITSM service
type UserDownloader interface {
	// DownloadUsers downloads and stores list of users from the ITSM service
	DownloadUsers(ctx context.Context) error

	// Close closes client connections
	Close() error
}

func NewUserDownloader(channelRepository repository.ChannelRepository, userRepository repository.UserRepository) UserDownloader {
	return &userDownloader{
		client:            http.DefaultClient,
		channelRepository: channelRepository,
		userRepository:    userRepository,
	}
}

type userDownloader struct {
	client            *http.Client
	channelRepository repository.ChannelRepository
	userRepository    repository.UserRepository
}

func (d *userDownloader) Close() error {
	d.client.CloseIdleConnections()
	return nil
}

func (d *userDownloader) DownloadUsers(ctx context.Context) error {
	channels, err := d.channelRepository.GetChannelList(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		fmt.Println(channel)

		// TODO stahnout usery pro kazdy kanal a ulozit
		//d.client.Do()
	}

	userList := user.List{
		user.User{
			ChannelID: "c5bea8d9-1d90-4d90-a445-e6ce74dff4cc",
			UserID:    "c8d1b9fb-35f1-46cb-aa37-a16b96937734",
			Email:     "first@user.com",
		},
		user.User{
			ChannelID: "c5bea8d9-1d90-4d90-a445-e6ce74dff4cc",
			UserID:    "b599fdbe-09df-47f9-9b08-c08caccab3b1",
			Email:     "second@user.com",
		},
		user.User{
			ChannelID: "8b6353c3-46ca-485d-87c3-66bc36c70d88",
			UserID:    "bb3f1241-6f52-4227-92fc-949385895cd5",
			Email:     "third@user.com",
		},
	}

	if err := d.userRepository.AddUserList(ctx, userList); err != nil {
		return err
	}

	return nil
}
