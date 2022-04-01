package ticketdownloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/client"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
)

// TicketClient gets ticket list (incidents and requests from external service
type TicketClient interface {
	// GetIncidents gets ticket list with incidents from external service
	GetIncidents(ctx context.Context, channel channel.Channel, user user.User) (ticket.List, error)

	// GetRequests gets ticket list with requests from external service
	GetRequests(ctx context.Context, channel channel.Channel, user user.User) (ticket.List, error)

	// Close closes client connections
	Close() error
}

func NewTicketClient(incidentClient, requestClient client.Client) TicketClient {
	return &ticketClient{
		incidentClient: incidentClient,
		requestClient:  requestClient,
	}
}

type ticketClient struct {
	incidentClient client.Client
	requestClient  client.Client
}

func (c ticketClient) GetIncidents(ctx context.Context, channel channel.Channel, user user.User) (ticket.List, error) {
	var ticketList ticket.List
	var bookmark string

	for {
		payload := c.preparePayload(user, bookmark)
		body := strings.NewReader(payload)
		resp, err := c.incidentClient.Query(ctx, channel.ChannelID, body)
		if err != nil {
			return ticketList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not retrieve info about incidents")
		}

		ticketList, bookmark, err = c.processResponse(resp, user, channel)
		if err != nil {
			return ticketList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not decode incident service Ok response")
		}

		fmt.Println(">> GetIncidents for ", user.Name, len(ticketList), bookmark)
		if len(ticketList) < 10 {
			break
		}
	}
	return ticketList, nil
}

func (c ticketClient) GetRequests(ctx context.Context, channel channel.Channel, user user.User) (ticket.List, error) {
	var ticketList ticket.List
	var bookmark string

	for {
		payload := c.preparePayload(user, bookmark)
		body := strings.NewReader(payload)
		resp, err := c.requestClient.Query(ctx, channel.ChannelID, body)
		if err != nil {
			return ticketList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not retrieve info about requests")
		}

		ticketList, bookmark, err = c.processResponse(resp, user, channel)
		if err != nil {
			return ticketList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not decode request service Ok response")
		}

		if len(ticketList) < 10 {
			break
		}

	}
	return ticketList, nil
}

func (c *ticketClient) Close() error {
	if err := c.incidentClient.Close(); err != nil {
		return err
	}
	if err := c.requestClient.Close(); err != nil {
		return err
	}
	return nil
}

func (c ticketClient) preparePayload(user user.User, bookmark string) string {
	// state_id: 4 = Resolved, 5 = Closed, 6 = Cancelled
	return `{"selector":{"$and":[{"state_id":{"$ne":4}},{"state_id":{"$ne":5}},{"state_id":{"$ne":6}}],"assigned_to":"` + user.UserID + `"},"fields":["uuid","number","short_description"],"bookmark":"` + bookmark + `"}`
}

func (c ticketClient) processResponse(resp *http.Response, user user.User, channel channel.Channel) (ticketList ticket.List, bookmark string, err error) {
	type OKPayload struct {
		Bookmark string `json:"bookmark"`
		Result   []struct {
			TicketType       string `json:"docType"`
			Number           string `json:"number"`
			ShortDescription string `json:"short_description"`
		} `json:"result"`
	}
	var okPayload OKPayload

	bookmark = okPayload.Bookmark

	defer func() { _ = resp.Body.Close() }()
	if err = json.NewDecoder(resp.Body).Decode(&okPayload); err != nil {
		return ticketList, bookmark, err
	}

	for _, v := range okPayload.Result {
		ticketList = append(ticketList, ticket.Ticket{
			UserEmail:   user.Email,
			ChannelName: channel.Name,
			TicketType:  v.TicketType,
			TicketData: ticket.Data{
				Number:           v.Number,
				ShortDescription: v.ShortDescription,
			},
		},
		)
	}

	return ticketList, bookmark, nil
}
