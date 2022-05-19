package ticketdownloader

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
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
		d.logger.Infow("Downloading tickets from the channel", "channel", channel.Name)

		var ticketList ticket.List
		ticketsCount := 0

		ticketList, err = d.client.GetIncidents(ctx, channel)
		if err != nil {
			return err
		}

		err = d.resolveAssignee(ctx, channel.ChannelID, ticketList)
		if err != nil {
			return err
		}

		if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
			return err
		}

		ticketsCount += len(ticketList)

		ticketList, err = d.client.GetRequests(ctx, channel)
		if err != nil {
			return err
		}

		err = d.resolveAssignee(ctx, channel.ChannelID, ticketList)
		if err != nil {
			return err
		}

		if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
			return err
		}

		ticketsCount += len(ticketList)

		d.logger.Infow("Tickets from the channel successfully downloaded", "channel", channel.Name, "tickets found", ticketsCount)
	}

	return nil
}

func (d *ticketDownloader) resolveAssignee(ctx context.Context, channelID string, ticketList ticket.List) error {
	for i, tckt := range ticketList {
		if tckt.UserID != "" {
			user, err := d.userRepository.GetUserInChannel(ctx, channelID, tckt.UserID)
			if err != nil {
				return domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not get user '%s' from repository", tckt.UserID)
			}

			tckt.UserName = user.Name
			tckt.UserEmail = user.Email
			tckt.UserOrgName = user.OrgName

			ticketList[i] = tckt
		}
	}

	return nil
}

func (d *ticketDownloader) Reset(ctx context.Context) error {
	return d.ticketRepository.Truncate(ctx)
}

func (d *ticketDownloader) Close() error {
	return d.client.Close()
}
