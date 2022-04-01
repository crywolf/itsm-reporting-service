package channeldownloader

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// ChannelDownloader downloads list of channels from the external service
type ChannelDownloader interface {
	// DownloadChannelList downloads and stores list of channels from the external service
	DownloadChannelList(ctx context.Context) error

	// Close closes client connections
	Close() error
}

func NewChannelDownloader(channelRepository repository.ChannelRepository, client ChannelClient) ChannelDownloader {
	return &channelDownloader{
		client:            client,
		channelRepository: channelRepository,
	}
}

type channelDownloader struct {
	client            ChannelClient
	channelRepository repository.ChannelRepository
}

func (d *channelDownloader) DownloadChannelList(ctx context.Context) error {
	channelList, err := d.client.GetChannels(ctx)
	if err != nil {
		return err
	}

	if err := d.channelRepository.StoreChannelList(ctx, channelList); err != nil {
		return err
	}

	return nil
}

func (d *channelDownloader) Close() error {
	return d.client.Close()
}
