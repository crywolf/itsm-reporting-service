package userdownloader

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"go.uber.org/zap"
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

func NewUserDownloader(
	logger *zap.SugaredLogger, channelRepository repository.ChannelRepository,
	userRepository repository.UserRepository, client UserClient,
) UserDownloader {
	return &userDownloader{
		logger:            logger,
		client:            client,
		channelRepository: channelRepository,
		userRepository:    userRepository,
	}
}

type userDownloader struct {
	logger            *zap.SugaredLogger
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
		d.logger.Infow("Downloading users from the channel", "channel", channel.Name)

		userList, err := d.client.GetEngineers(ctx, channel)
		if err != nil {
			return err
		}

		if err := d.userRepository.AddUserList(ctx, userList); err != nil {
			return err
		}

		d.logger.Infow("Users from the channel successfully downloaded", "channel", channel.Name, "users found", len(userList))
	}

	return nil
}

func (d *userDownloader) Reset(ctx context.Context) error {
	return d.userRepository.Truncate(ctx)
}

func (d *userDownloader) Close() error {
	return d.client.Close()
}
