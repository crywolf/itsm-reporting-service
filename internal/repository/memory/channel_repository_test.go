package memory

import (
	"context"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelRepositoryMemory_StoringAndGettingChannelList(t *testing.T) {
	ctx := context.Background()
	repo := NewChannelRepositoryMemory()

	ch1 := channel.Channel{
		ChannelID: "c5bea8d9-1d90-4d90-a445-e6ce74dff4cc",
		Name:      "First channel",
	}

	ch2 := channel.Channel{
		ChannelID: "8b6353c3-46ca-485d-87c3-66bc36c70d88",
		Name:      "Second channel",
	}

	list := channel.List{ch1, ch2}

	err := repo.StoreChannelList(ctx, list)
	require.NoError(t, err)

	retChannelList, err := repo.GetChannelList(ctx)
	require.NoError(t, err)

	assert.Len(t, retChannelList, 2)
	assert.Equal(t, ch1, retChannelList[0])
	assert.Equal(t, ch2, retChannelList[1])
}
