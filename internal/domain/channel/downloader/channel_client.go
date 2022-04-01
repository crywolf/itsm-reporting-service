package channeldownloader

import (
	"context"
	"encoding/json"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/client"
)

// ChannelClient gets channel list from external service
type ChannelClient interface {
	// GetChannels gets channel list from external service
	GetChannels(ctx context.Context) (channel.List, error)

	// Close closes client connections
	Close() error
}

func NewChannelClient(client client.Client) ChannelClient {
	return &channelClient{
		Client: client,
	}
}

type channelClient struct {
	client.Client
}

func (c channelClient) GetChannels(ctx context.Context) (channel.List, error) {
	var channelList channel.List

	resp, err := c.Get(ctx, "")
	if err != nil {
		return channelList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not retrieve info about channels")
	}

	type OKPayload struct {
		Result []struct {
			ID   string `json:"space"`
			Name string `json:"name"`
		} `json:"spaces"`
	}
	var payload OKPayload

	defer func() { _ = resp.Body.Close() }()
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return channelList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not decode channel service Ok response")
	}

	for _, v := range payload.Result {
		channelList = append(channelList, channel.Channel{
			ChannelID: v.ID,
			Name:      v.Name,
		},
		)
	}

	return channelList, nil
}
