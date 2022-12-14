package userdownloader

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/client"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
)

// UserClient gets list of users from external service
type UserClient interface {
	// GetUsers gets users from external service
	GetUsers(ctx context.Context, channel channel.Channel) (user.List, error)

	// Close closes client connections
	Close() error
}

func NewUserClient(client client.Client) UserClient {
	return &userClient{
		Client: client,
	}
}

type userClient struct {
	client.Client
}

func (c userClient) GetUsers(ctx context.Context, channel channel.Channel) (user.List, error) {
	var userList user.List
	var bookmark string

	for {
		payload := `{"fields":["uuid","full_name","email","type","org_display_name"],"bookmark":"` + bookmark + `"}`
		body := strings.NewReader(payload)
		resp, err := c.Query(ctx, channel.ChannelID, body)
		if err != nil {
			return userList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not retrieve info about users")
		}

		type OKPayload struct {
			Bookmark string `json:"bookmark"`
			Result   []struct {
				ID      string `json:"uuid"`
				Name    string `json:"full_name"`
				Email   string `json:"email"`
				Type    string `json:"type"`
				OrgName string `json:"org_display_name"`
			} `json:"result"`
		}
		var okPayload OKPayload

		if err = json.NewDecoder(resp.Body).Decode(&okPayload); err != nil {
			return userList, domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not decode user service Ok response")
		}

		_ = resp.Body.Close()

		bookmark = okPayload.Bookmark

		for _, v := range okPayload.Result {
			if v.Email == "" { // this is in case of some data inconsistency in ITSM, we need to skip invalid users
				continue
			}

			userList = append(userList, user.User{
				ChannelID: channel.ChannelID,
				UserID:    v.ID,
				Email:     v.Email,
				Name:      v.Name,
				Type:      v.Type,
				OrgName:   v.OrgName,
			},
			)
		}

		if len(okPayload.Result) < 10 {
			break
		}
	}

	return userList, nil
}
