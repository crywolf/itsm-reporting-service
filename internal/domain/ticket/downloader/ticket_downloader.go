package ticketdownloader

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"go.uber.org/zap"
)

// TicketDownloader downloads list of users from the ITSM service
type TicketDownloader interface {
	// DownloadTickets downloads and stores list of tickets from the ITSM service
	DownloadTickets(ctx context.Context) error

	// Reset removes all items from downloader repository
	Reset(ctx context.Context) error

	// Close closes client connections
	Close() error
}

func NewTicketDownloader(
	logger *zap.SugaredLogger,
	channelRepository repository.ChannelRepository,
	userRepository repository.UserRepository,
	ticketRepository repository.TicketRepository,
	client TicketClient,
) TicketDownloader {
	return &ticketDownloader{
		logger:            logger,
		client:            client,
		channelRepository: channelRepository,
		userRepository:    userRepository,
		ticketRepository:  ticketRepository,
	}
}

type ticketDownloader struct {
	logger            *zap.SugaredLogger
	client            TicketClient
	channelRepository repository.ChannelRepository
	userRepository    repository.UserRepository
	ticketRepository  repository.TicketRepository
}

func (d *ticketDownloader) DownloadTickets(ctx context.Context) error {
	channels, err := d.channelRepository.GetChannelList(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		userList, err := d.userRepository.GetUsersByChannel(ctx, channel.ChannelID)
		if err != nil {
			return err
		}

		d.logger.Infow("Downloading tickets from the channel", "channel", channel.Name, "users found", len(userList))

		var ticketList ticket.List

		for _, user := range userList {
			ticketList, err = d.client.GetIncidents(ctx, channel, user)
			if err != nil {
				return err
			}

			if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
				return err
			}

			ticketList, err = d.client.GetRequests(ctx, channel, user)
			if err != nil {
				return err
			}

			if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
				return err
			}
		}

		d.logger.Infof("Tickets from the '%s' channel succesfully downloaded", channel.Name)
	}

	return nil
}

func (d *ticketDownloader) Reset(ctx context.Context) error {
	return d.ticketRepository.Truncate(ctx)
}

func (d *ticketDownloader) Close() error {
	return d.client.Close()
}
