package ticketdownloader

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// TicketDownloader downloads list of users from the ITSM service
type TicketDownloader interface {
	// DownloadTickets downloads and stores list of tickets from the ITSM service
	DownloadTickets(ctx context.Context) error

	// Close closes client connections
	Close() error
}

func NewTicketDownloader(channelRepository repository.ChannelRepository, userRepository repository.UserRepository, ticketRepository repository.TicketRepository) TicketDownloader {
	return &ticketDownloader{
		client:            http.DefaultClient,
		channelRepository: channelRepository,
		userRepository:    userRepository,
		ticketRepository:  ticketRepository,
	}
}

type ticketDownloader struct {
	client            *http.Client
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

		fmt.Println(userList)
		// TODO stahnout tickety pro kazdy kanal a uzivatele a ulozit
		//d.client.Do()

		var ticketList ticket.List

		for i, user := range userList {
			ticketList = append(ticketList, ticket.Ticket{
				UserEmail:   user.Email,
				ChannelName: channel.Name,
				TicketType:  "incident",
				TicketData: ticket.Data{
					Number:           "INC123456",
					ShortDescription: fmt.Sprintf("Inc %d", i),
				},
			})
		}

		if err := d.ticketRepository.AddTicketList(ctx, ticketList); err != nil {
			return err
		}

		fmt.Printf("===> Tickets from the channel %s downloaded %v\n", channel.Name, ticketList)
	}

	return nil
}

func (d *ticketDownloader) Close() error {
	//TODO implement me
	panic("implement me")
}
