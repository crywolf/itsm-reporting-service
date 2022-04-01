package ticketdownloader

import (
	"context"
	"fmt"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
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

func NewTicketDownloader(channelRepository repository.ChannelRepository, userRepository repository.UserRepository, ticketRepository repository.TicketRepository, client TicketClient) TicketDownloader {
	return &ticketDownloader{
		client:            client,
		channelRepository: channelRepository,
		userRepository:    userRepository,
		ticketRepository:  ticketRepository,
	}
}

type ticketDownloader struct {
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

		fmt.Println("\n>>> len(userList):", len(userList))

		var ticketList ticket.List

		for _, user := range userList {
			ticketList, err = d.client.GetIncidents(ctx, channel, user)
			if err != nil {
				return err
			}

			fmt.Println("---> returned INCIDENTS for user:", user.Email, user.Name, len(ticketList))

			if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
				return err
			}

			ticketList, err = d.client.GetRequests(ctx, channel, user)
			if err != nil {
				return err
			}

			fmt.Println("---> returned REQUESTS for user:", user.Email, user.Name, len(ticketList))

			if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
				return err
			}
		}

		fmt.Printf("===> Tickets from the channel %s downloaded\n", channel.Name)
	}

	return nil
}

func (d *ticketDownloader) Reset(ctx context.Context) error {
	return d.ticketRepository.Truncate(ctx)
}

func (d *ticketDownloader) Close() error {
	return d.client.Close()
}
