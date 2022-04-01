package jobprocessor

import (
	"context"
	"testing"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	chandownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/channel/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	ticketdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/ticket/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
	userdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/user/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/KompiTech/itsm-reporting-service/internal/repository/memory"
	"github.com/KompiTech/itsm-reporting-service/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_processor_ProcessNewJob(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	lastJob := job.Job{}
	err := lastJob.SetUUID("d6aa467b-d07d-41e0-9182-aeedb1b02398")
	require.NoError(t, err)

	lastJob2 := job.Job{}
	err = lastJob2.SetUUID("c0582f65-4c7d-469f-a3a4-42360f287074")
	require.NoError(t, err)

	t.Run("when the processor is busy", func(t *testing.T) {
		jobsRepo := new(mocks.JobRepositoryMock)
		jobsRepo.On("GetLastJob").Return(lastJob, nil).Once()
		jobsRepo.On("GetJob", lastJob.UUID()).Return(lastJob, nil)
		jobsRepo.On("UpdateJob", mock.AnythingOfType("job.Job")).Return(lastJob.UUID(), nil)

		channelDownloader := new(mocks.ChannelDownloaderMock)
		channelDownloader.On("DownloadChannelList").Return(nil).Once()

		userDownloader := new(mocks.UserDownloaderMock)
		userDownloader.On("DownloadUsers").Return(nil).Once()

		ticketDownloader := new(mocks.TicketDownloaderMock)
		ticketDownloader.On("DownloadTickets").Return(nil).Once()
		ticketDownloader.Wg.Add(1)

		jp := NewJobProcessor(logger, jobsRepo, channelDownloader, userDownloader, ticketDownloader)
		jp.WaitForJobs()

		err = jp.ProcessNewJob(lastJob.UUID())
		assert.NoError(t, err, "unexpected error", err)

		err = jp.ProcessNewJob(lastJob2.UUID()) // this call should return error
		assert.Error(t, err, "expecting error but none returned")
		var domainErr *domain.Error
		assert.ErrorAs(t, err, &domainErr)
		assert.EqualError(t, err, "job is being processed, try it later")

		ticketDownloader.Wg.Wait() // wait for job processor to finish

		jobsRepo.AssertExpectations(t)
		channelDownloader.AssertExpectations(t)
		userDownloader.AssertExpectations(t)
		ticketDownloader.AssertExpectations(t)
	})

	t.Run("when the processor is ready", func(t *testing.T) {
		jobsRepo := new(mocks.JobRepositoryMock)
		jobsRepo.On("GetLastJob").Return(lastJob, nil).Once()
		jobsRepo.On("GetLastJob").Return(lastJob2, nil).Once()
		jobsRepo.On("GetJob", lastJob.UUID()).Return(lastJob, nil).Times(4)
		jobsRepo.On("GetJob", lastJob2.UUID()).Return(lastJob2, nil)
		jobsRepo.On("UpdateJob", mock.AnythingOfType("job.Job")).Return(lastJob.UUID(), nil)

		channelDownloader := new(mocks.ChannelDownloaderMock)
		channelDownloader.On("DownloadChannelList").Return(nil).Twice()

		userDownloader := new(mocks.UserDownloaderMock)
		userDownloader.On("DownloadUsers").Return(nil).Twice()

		ticketDownloader := new(mocks.TicketDownloaderMock)
		ticketDownloader.On("DownloadTickets").Return(nil).Twice()
		ticketDownloader.Wg.Add(2)

		jp := NewJobProcessor(logger, jobsRepo, channelDownloader, userDownloader, ticketDownloader)
		jp.WaitForJobs()

		err = jp.ProcessNewJob(lastJob.UUID())
		assert.NoError(t, err, "unexpected error", err)

		time.Sleep(100 * time.Millisecond) // wait for processor to get ready - TODO: do it better later?
		err = jp.ProcessNewJob(lastJob2.UUID())
		assert.NoError(t, err, "unexpected error", err)

		ticketDownloader.Wg.Wait() // wait for job processor to finish

		jobsRepo.AssertExpectations(t)
		channelDownloader.AssertExpectations(t)
		userDownloader.AssertExpectations(t)
		ticketDownloader.AssertExpectations(t)
	})
}

func Test_processor_DataProcessing(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	lastJob := job.Job{}
	err := lastJob.SetUUID("d6aa467b-d07d-41e0-9182-aeedb1b02398")
	require.NoError(t, err)

	jobsRepo := new(mocks.JobRepositoryMock)
	jobsRepo.On("GetLastJob").Return(lastJob, nil).Once()
	jobsRepo.On("GetJob", lastJob.UUID()).Return(lastJob, nil)
	jobsRepo.On("UpdateJob", mock.AnythingOfType("job.Job")).Return(lastJob.UUID(), nil)

	ch1 := channel.Channel{
		ChannelID: "c5bea8d9-1d90-4d90-a445-e6ce74dff4cc",
		Name:      "First channel",
	}
	ch2 := channel.Channel{
		ChannelID: "8b6353c3-46ca-485d-87c3-66bc36c70d88",
		Name:      "Second channel",
	}
	channelList := channel.List{
		ch1,
		ch2,
	}

	email1 := "first@user.com"
	email2 := "second@user.com"

	u1 := user.User{
		ChannelID: ch1.ChannelID,
		UserID:    "c8d1b9fb-35f1-46cb-aa37-a16b96937734",
		Email:     email1,
	}
	u2 := user.User{
		ChannelID: ch1.ChannelID,
		UserID:    "b599fdbe-09df-47f9-9b08-c08caccab3b1",
		Email:     email2,
	}
	userListChan1 := user.List{u1, u2}

	u3 := user.User{
		ChannelID: ch2.ChannelID,
		UserID:    "b599fdbe-09df-47f9-9b08-c08caccab3b1",
		Email:     email2,
	}
	userListChan2 := user.List{u3}

	inc1 := ticket.Ticket{
		UserEmail:   u1.Email,
		ChannelName: ch1.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC1111",
			ShortDescription: "Incident 1111",
		},
	}
	inc2 := ticket.Ticket{
		UserEmail:   u1.Email,
		ChannelName: ch1.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC2222",
			ShortDescription: "Incident 2222",
		},
	}
	inc3 := ticket.Ticket{
		UserEmail:   u1.Email,
		ChannelName: ch1.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC3333",
			ShortDescription: "Incident 3333",
		},
	}
	incidentListU1Ch1 := ticket.List{inc1, inc2, inc3}

	inc4 := ticket.Ticket{
		UserEmail:   u3.Email,
		ChannelName: ch2.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC4444",
			ShortDescription: "Incident 4444",
		},
	}
	incidentListU2Ch3 := ticket.List{inc4}

	req1 := ticket.Ticket{
		UserEmail:   u1.Email,
		ChannelName: ch1.Name,
		TicketType:  "K_REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ1111",
			ShortDescription: "Request 1111",
		},
	}
	requestListU1Ch1 := ticket.List{req1}

	req2 := ticket.Ticket{
		UserEmail:   u2.Email,
		ChannelName: ch1.Name,
		TicketType:  "K_REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ2222",
			ShortDescription: "Request 2222",
		},
	}
	requestListU2Ch1 := ticket.List{req2}

	channelClient := new(mocks.ChannelClientMock)
	channelClient.On("GetChannels").Return(channelList, nil).Once()

	userClient := new(mocks.UserClientMock)
	userClient.On("GetEngineers", ch1).Return(userListChan1, nil).Once()
	userClient.On("GetEngineers", ch2).Return(userListChan2, nil).Once()

	ticketClient := new(mocks.TicketClientMock)
	ticketClient.On("GetIncidents", ch1, u1).Return(incidentListU1Ch1, nil).Once()
	ticketClient.On("GetIncidents", ch1, u2).Return(ticket.List{}, nil).Once()
	ticketClient.On("GetIncidents", ch2, u3).Return(incidentListU2Ch3, nil).Once()

	ticketClient.On("GetRequests", ch1, u1).Return(requestListU1Ch1, nil).Once()
	ticketClient.On("GetRequests", ch1, u2).Return(requestListU2Ch1, nil).Once()
	ticketClient.On("GetRequests", ch2, u3).Return(ticket.List{}, nil).Once()
	ticketClient.Wg.Add(6)

	channelRepository := memory.NewChannelRepositoryMemory()
	channelDownloader := chandownloader.NewChannelDownloader(channelRepository, channelClient)
	userRepository := memory.NewUserRepositoryMemory()
	userDownloader := userdownloader.NewUserDownloader(channelRepository, userRepository, userClient)
	ticketRepository := memory.NewTicketRepositoryMemory()
	ticketDownloader := ticketdownloader.NewTicketDownloader(channelRepository, userRepository, ticketRepository, ticketClient)

	jp := NewJobProcessor(logger, jobsRepo, channelDownloader, userDownloader, ticketDownloader)
	jp.WaitForJobs()

	err = jp.ProcessNewJob(lastJob.UUID())
	assert.NoError(t, err, "unexpected error", err)

	ticketClient.Wg.Wait() // wait for job processor to finish

	ctx := context.Background()

	expectedUserListCh1, err := userRepository.GetUsersByChannel(ctx, ch1.ChannelID)
	require.NoError(t, err)

	expectedUserListCh2, err := userRepository.GetUsersByChannel(ctx, ch2.ChannelID)
	require.NoError(t, err)

	//	fmt.Println(">>>>>>>>> TEST user list:", expectedUserListCh1, expectedUserListCh2)

	assert.Len(t, expectedUserListCh1, len(userListChan1), "len of users in channel 1")
	assert.Len(t, expectedUserListCh2, len(userListChan2), "len of users in channel 2")

	// TODO GetTicketsByEmail - bud pro jistotu setridit a vracet jako prvni incidenty nebo radeji volat 2x, jednou pro Inc a pak pro Req
	expectedTicketListU1, err := ticketRepository.GetTicketsByEmail(ctx, email1)
	require.NoError(t, err)

	assert.Len(t, expectedTicketListU1, 4, "len of tickets for email 1 = user 1")
	for _, tckt := range expectedTicketListU1 {
		assert.Equal(t, tckt.UserEmail, email1)
	}

	expectedTicketListU2, err := ticketRepository.GetTicketsByEmail(ctx, email2)
	require.NoError(t, err)

	assert.Len(t, expectedTicketListU2, 2, "len of tickets for email 2 = user2 + user3")
	for _, tckt := range expectedTicketListU2 {
		assert.Equal(t, tckt.UserEmail, email2)
	}

	//	fmt.Println(">>>>>>>>> TEST ticket list:", expectedTicketListU1, expectedTicketListU2)

	jobsRepo.AssertExpectations(t)
	channelClient.AssertExpectations(t)
	userClient.AssertExpectations(t)
	ticketClient.AssertExpectations(t)
}
