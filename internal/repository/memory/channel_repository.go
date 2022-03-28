package memory

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// NewChannelRepositoryMemory returns new initialized channel repository that keeps data in memory
func NewChannelRepositoryMemory() repository.ChannelRepository {
	return &channelRepositoryMemory{}
}

type channelRepositoryMemory struct {
	channelList channel.List
}

func (r *channelRepositoryMemory) StoreChannelList(_ context.Context, channelList channel.List) error {
	r.channelList = channelList
	return nil
}

func (r channelRepositoryMemory) GetChannelList(_ context.Context) (channel.List, error) {
	return r.channelList, nil
}
