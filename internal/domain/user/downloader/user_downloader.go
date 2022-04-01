package userdownloader

import (
	"context"
	"fmt"

	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// UserDownloader downloads list of users from the ITSM service
type UserDownloader interface {
	// DownloadUsers downloads and stores list of users from the ITSM service
	DownloadUsers(ctx context.Context) error

	// Reset removes all items from downloader repository
	Reset(ctx context.Context) error

	// Close closes client connections
	Close() error
}

func NewUserDownloader(channelRepository repository.ChannelRepository, userRepository repository.UserRepository, client UserClient) UserDownloader {
	return &userDownloader{
		client:            client,
		channelRepository: channelRepository,
		userRepository:    userRepository,
	}
}

type userDownloader struct {
	client            UserClient
	channelRepository repository.ChannelRepository
	userRepository    repository.UserRepository
}

func (d *userDownloader) DownloadUsers(ctx context.Context) error {
	channels, err := d.channelRepository.GetChannelList(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		fmt.Println("DownloadUsers - channel:", channel)

		userList, err := d.client.GetEngineers(ctx, channel)
		if err != nil {
			return err
		}

		if err := d.userRepository.AddUserList(ctx, userList); err != nil {
			return err
		}
	}

	return nil
}

func (d *userDownloader) Reset(ctx context.Context) error {
	return d.userRepository.Truncate(ctx)
}

func (d *userDownloader) Close() error {
	return d.client.Close()
}
