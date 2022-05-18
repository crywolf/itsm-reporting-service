package jobprocessor

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	chandownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/channel/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/excel"
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

		excelGen := new(mocks.ExcelGeneratorMock)
		excelGen.On("GenerateExcelFilesForFieldEngineers").Return(nil).Once()
		excelGen.On("GenerateExcelFilesForServiceDesk").Return(nil).Once()

		emailSender := new(mocks.EmailSenderMock)
		emailSender.On("SendEmailsForFieldEngineers").Return(nil).Once()
		emailSender.Wg.Add(1)

		emailSender.On("SendEmailsForServiceDesk").Return(nil).Once()
		emailSender.Wg.Add(1)

		jp := NewJobProcessor(logger, jobsRepo, channelDownloader, userDownloader, ticketDownloader, excelGen, emailSender)
		jp.WaitForJobs()

		err = jp.ProcessNewJob(lastJob.UUID())
		assert.NoError(t, err, "unexpected error", err)

		err = jp.ProcessNewJob(lastJob2.UUID()) // this call should return error
		assert.Error(t, err, "expecting error but none returned")
		var domainErr *domain.Error
		assert.ErrorAs(t, err, &domainErr)
		assert.EqualError(t, err, "job is being processed, try it later")

		emailSender.Wg.Wait() // wait for job processor to finish

		jobsRepo.AssertExpectations(t)
		channelDownloader.AssertExpectations(t)
		userDownloader.AssertExpectations(t)
		ticketDownloader.AssertExpectations(t)
		excelGen.AssertExpectations(t)
		emailSender.AssertExpectations(t)
	})

	t.Run("when the processor is ready", func(t *testing.T) {
		jobsRepo := new(mocks.JobRepositoryMock)
		jobsRepo.On("GetLastJob").Return(lastJob, nil).Once()
		jobsRepo.On("GetLastJob").Return(lastJob2, nil).Once()
		jobsRepo.On("GetJob", lastJob.UUID()).Return(lastJob, nil).Times(6)
		jobsRepo.On("GetJob", lastJob2.UUID()).Return(lastJob2, nil)
		jobsRepo.On("UpdateJob", mock.AnythingOfType("job.Job")).Return(lastJob.UUID(), nil).Times(12)
		jobsRepo.On("UpdateJob", mock.AnythingOfType("job.Job")).Return(lastJob2.UUID(), nil)

		channelDownloader := new(mocks.ChannelDownloaderMock)
		channelDownloader.On("DownloadChannelList").Return(nil).Twice()

		userDownloader := new(mocks.UserDownloaderMock)
		userDownloader.On("DownloadUsers").Return(nil).Twice()

		ticketDownloader := new(mocks.TicketDownloaderMock)
		ticketDownloader.On("DownloadTickets").Return(nil).Twice()

		excelGen := new(mocks.ExcelGeneratorMock)
		excelGen.On("GenerateExcelFilesForFieldEngineers").Return(nil).Twice()
		excelGen.On("GenerateExcelFilesForServiceDesk").Return(nil).Twice()

		emailSender := new(mocks.EmailSenderMock)
		emailSender.On("SendEmailsForFieldEngineers").Return(nil).Twice()
		emailSender.Wg.Add(2)

		emailSender.On("SendEmailsForServiceDesk").Return(nil).Twice()
		emailSender.Wg.Add(2)

		jp := NewJobProcessor(
			logger,
			jobsRepo,
			channelDownloader,
			userDownloader,
			ticketDownloader,
			excelGen,
			emailSender,
		)
		jp.WaitForJobs()

		err = jp.ProcessNewJob(lastJob.UUID())
		assert.NoError(t, err, "unexpected error", err)

		time.Sleep(200 * time.Millisecond) // wait for processor to get ready - TODO: do it better later?
		err = jp.ProcessNewJob(lastJob2.UUID())
		assert.NoError(t, err, "unexpected error", err)

		emailSender.Wg.Wait() // wait for job processor to finish

		jobsRepo.AssertExpectations(t)
		channelDownloader.AssertExpectations(t)
		userDownloader.AssertExpectations(t)
		ticketDownloader.AssertExpectations(t)
		excelGen.AssertExpectations(t)
		emailSender.AssertExpectations(t)
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
		Name:      "Alfons",
		Type:      "engineer",
		OrgName:   "KompiTech",
	}
	u2 := user.User{
		ChannelID: ch1.ChannelID,
		UserID:    "b599fdbe-09df-47f9-9b08-c08caccab3b1",
		Email:     email2,
		Name:      "Bill",
		Type:      "",
		OrgName:   "ABC",
	}
	userListChan1 := user.List{u1, u2}

	u3 := user.User{
		ChannelID: ch2.ChannelID,
		UserID:    "b599fdbe-09df-47f9-9b08-c08caccab3b1",
		Email:     email2,
		Name:      "Bill 2",
		Type:      "xxx",
		OrgName:   "ABC",
	}
	userListChan2 := user.List{u3}

	inc1 := ticket.Ticket{
		UserID:      u1.UserID,
		UserOrgName: u1.OrgName,
		ChannelID:   ch1.ChannelID,
		ChannelName: ch1.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC1111",
			ShortDescription: "Incident 1111",
			StateID:          0,
			Location:         "Cz Praha 10 Czechia  Jerevanská, 1158",
		},
	}
	inc2 := ticket.Ticket{
		UserID:      u1.UserID,
		ChannelID:   ch1.ChannelID,
		ChannelName: ch1.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC2222",
			ShortDescription: "Incident 2222",
			StateID:          2,
			Location:         "Bel Bürglen Switzerland  Chaussée de Liège",
		},
	}
	inc3 := ticket.Ticket{
		UserID:      u1.UserID,
		ChannelID:   ch1.ChannelID,
		ChannelName: ch1.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC3333",
			ShortDescription: "Incident 3333",
			StateID:          3,
			Location:         "Nid Nuland Netherlands  Nulandsestraat, 2",
		},
	}

	inc4 := ticket.Ticket{
		UserID:      u3.UserID,
		ChannelID:   ch2.ChannelID,
		ChannelName: ch2.Name,
		TicketType:  "INCIDENT",
		TicketData: ticket.Data{
			Number:           "INC4444",
			ShortDescription: "Incident 4444",
			StateID:          0,
			Location:         "Sp Teruel Spain  Poligono 25, 66",
		},
	}

	req1 := ticket.Ticket{
		UserID:      u1.UserID,
		ChannelID:   ch1.ChannelID,
		ChannelName: ch1.Name,
		TicketType:  "K_REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ1111",
			ShortDescription: "Request 1111",
			StateID:          3,
			Location:         "Au Schaldorf Austria  Neubaugasse, 13",
		},
	}

	req2 := ticket.Ticket{
		UserID:      u2.UserID,
		ChannelID:   ch1.ChannelID,
		ChannelName: ch1.Name,
		TicketType:  "K_REQUEST",
		TicketData: ticket.Data{
			Number:           "REQ2222",
			ShortDescription: "Request 2222",
			StateID:          2,
			Location:         "Cor Corbie France  Route de Bray",
		},
	}

	incListCh1 := ticket.List{inc1, inc2, inc3}
	reqListCh1 := ticket.List{req1, req2}
	incListCh2 := ticket.List{inc4}

	channelClient := new(mocks.ChannelClientMock)
	channelClient.On("GetChannels").Return(channelList, nil).Once()

	userClient := new(mocks.UserClientMock)
	userClient.On("GetEngineers", ch1).Return(userListChan1, nil).Once()
	userClient.On("GetEngineers", ch2).Return(userListChan2, nil).Once()

	ticketClient := new(mocks.TicketClientMock)
	ticketClient.On("GetIncidents", ch1).Return(incListCh1, nil).Once()
	ticketClient.On("GetIncidents", ch2).Return(incListCh2, nil).Once()

	ticketClient.On("GetRequests", ch1).Return(reqListCh1, nil).Once()
	ticketClient.On("GetRequests", ch2).Return(ticket.List{}, nil).Once()
	ticketClient.Wg.Add(4)

	channelRepository := memory.NewChannelRepositoryMemory()
	channelDownloader := chandownloader.NewChannelDownloader(channelRepository, channelClient)
	userRepository := memory.NewUserRepositoryMemory()
	userDownloader := userdownloader.NewUserDownloader(logger, channelRepository, userRepository, userClient)
	ticketRepository := memory.NewTicketRepositoryMemory()
	ticketDownloader := ticketdownloader.NewTicketDownloader(logger, channelRepository, userRepository, ticketRepository, ticketClient)

	sdAgentEmails := []string{"firstSDAgent@email.test", "secondSDAgent@email.test", "thirdSDAgent@email.test"}
	excelGen := excel.NewExcelGenerator(logger, ticketRepository, sdAgentEmails)

	emailSender := new(mocks.EmailSenderMock)
	emailSender.On("SendEmailsForFieldEngineers").Return(nil)
	emailSender.Wg.Add(1)

	emailSender.On("SendEmailsForServiceDesk").Return(nil)
	emailSender.Wg.Add(1)

	jp := NewJobProcessor(
		logger,
		jobsRepo,
		channelDownloader,
		userDownloader,
		ticketDownloader,
		excelGen,
		emailSender,
	)
	jp.WaitForJobs()

	err = jp.ProcessNewJob(lastJob.UUID())
	assert.NoError(t, err, "unexpected error", err)

	ticketClient.Wg.Wait() // wait for job processor to finish

	emailSender.Wg.Wait() // wait for job processor to finish

	ctx := context.Background()

	expectedTicketListU1, err := ticketRepository.GetTicketsByEmailAddress(ctx, email1)
	require.NoError(t, err)

	assert.Len(t, expectedTicketListU1, 4, "len of tickets for email 1 = tickets for user 1")
	for _, tckt := range expectedTicketListU1 {
		assert.Equal(t, tckt.UserEmail, email1)
		assert.Equal(t, tckt.UserID, u1.UserID)
		assert.Equal(t, tckt.UserName, u1.Name)
		assert.Equal(t, tckt.UserOrgName, u1.OrgName)
	}

	expectedTicketListU2, err := ticketRepository.GetTicketsByEmailAddress(ctx, email2)
	require.NoError(t, err)

	assert.Len(t, expectedTicketListU2, 2, "len of tickets for email 2 = tickets for user2 + user3")

	assert.Equal(t, email2, expectedTicketListU2[0].UserEmail)
	assert.Equal(t, u3.Name, expectedTicketListU2[0].UserName)
	assert.Equal(t, u3.OrgName, expectedTicketListU2[0].UserOrgName)

	assert.Equal(t, email2, expectedTicketListU2[1].UserEmail)
	assert.Equal(t, u2.Name, expectedTicketListU2[1].UserName)
	assert.Equal(t, u2.OrgName, expectedTicketListU2[1].UserOrgName)

	// check generated Excel files
	// 1) files for field engineers
	filesFE, err := os.ReadDir(excelGen.FEDirPath())
	require.NoError(t, err)

	assert.Len(t, filesFE, 2, "Excel files count for FE == email addresses count")
	assert.Equal(t, filesFE[0].Name(), email1+".xlsx")
	assert.Equal(t, filesFE[1].Name(), email2+".xlsx")

	// 2) files for service desk
	filesSD, err := os.ReadDir(excelGen.SDDirPath())
	require.NoError(t, err)

	assert.Len(t, filesSD, 3, "Excel files count for SD == email addresses count")
	assert.Equal(t, filesSD[0].Name(), sdAgentEmails[0]+".xlsx")
	assert.Equal(t, filesSD[1].Name(), sdAgentEmails[1]+".xlsx")
	assert.Equal(t, filesSD[2].Name(), sdAgentEmails[2]+".xlsx")

	jobsRepo.AssertExpectations(t)
	channelClient.AssertExpectations(t)
	userClient.AssertExpectations(t)
	ticketClient.AssertExpectations(t)
}
