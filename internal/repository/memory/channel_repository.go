package memory

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

type channelRepositoryMemory struct {
	channelList channel.List
}

// NewChannelRepositoryMemory returns new initialized repository
func NewChannelRepositoryMemory() repository.ChannelRepository {
	return &channelRepositoryMemory{}
}

func (r *channelRepositoryMemory) StoreChannelList(_ context.Context, channelList channel.List) error {
	r.channelList = channelList
	return nil
}

func (r channelRepositoryMemory) GetChannelList(_ context.Context) (channel.List, error) {
	return r.channelList, nil
}
