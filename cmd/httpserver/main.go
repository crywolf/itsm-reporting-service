package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chandownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/channel/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/client"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/email"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/excel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job/processor"
	jobsvc "github.com/KompiTech/itsm-reporting-service/internal/domain/job/service"
	ticketdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/ticket/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
	userdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/user/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest"
	"github.com/KompiTech/itsm-reporting-service/internal/repository/memory"
	"github.com/KompiTech/itsm-reporting-service/internal/repository/sql"
	"go.uber.org/zap"
)

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (c realClock) NowFormatted() types.DateTime {
	return types.DateTime(c.Now().Format(time.RFC3339))
}

func main() {
	l, _ := zap.NewProduction()
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)

	logger := l.Sugar()

	logger.Info("App starting...")

	config, err := loadEnvConfig()
	if err != nil {
		logger.Fatalw("Error loading configuration", "error", err)
	}

	clock := realClock{}
	db, err := sql.OpenDB(config.DBConnectionString)
	if err != nil {
		logger.Fatalw("Error opening database", "error", err)
	}

	jobRepository, err := sql.NewJobRepositorySQL(clock, db, nil)
	if err != nil {
		logger.Fatalw("Error creating jobRepositorySQL", "error", err)
	}
	jobService := jobsvc.NewJobService(jobRepository)

	tokenSvcClient := client.NewTokenSvcClient(client.Config{
		AssertionToken:         config.AssertionToken,
		AssertionTokenEndpoint: config.AssertionTokenEndpoint,
		AssertionTokenOrg:      config.AssertionTokenOrg,
	})

	channelRepository := memory.NewChannelRepositoryMemory()
	channelClient := chandownloader.NewChannelClient(client.NewHTTPClient(config.ChannelEndpointPath, logger, tokenSvcClient))
	channelDownloader := chandownloader.NewChannelDownloader(channelRepository, channelClient)

	userRepository := memory.NewUserRepositoryMemory()
	userClient := userdownloader.NewUserClient(client.NewHTTPClient(config.UserEndpointPath, logger, tokenSvcClient))
	userDownloader := userdownloader.NewUserDownloader(channelRepository, userRepository, userClient)

	ticketRepository := memory.NewTicketRepositoryMemory()
	ticketClient := ticketdownloader.NewTicketClient(
		client.NewHTTPClient(config.IncidentEndpointPath, logger, tokenSvcClient),
		client.NewHTTPClient(config.RequestEndpointPath, logger, tokenSvcClient),
	)
	ticketDownloader := ticketdownloader.NewTicketDownloader(logger, channelRepository, userRepository, ticketRepository, ticketClient)

	excelGen := excel.NewExcelGenerator(logger, ticketRepository)

	emailSender := email.NewEmailSender(
		logger,
		config.PostmarkServerURL,
		config.PostmarkServerToken,
		config.PostmarkMessageStream,
		config.FromEmailAddress,
		excelGen.DirName(),
		ticketRepository,
	)

	jobProcessor := jobprocessor.NewJobProcessor(
		logger,
		jobRepository,
		channelDownloader,
		userDownloader,
		ticketDownloader,
		excelGen,
		emailSender,
	)

	// HTTP server
	server := rest.NewServer(rest.Config{
		Addr:                    config.HTTPBindAddress,
		URISchema:               "http://",
		Logger:                  logger,
		JobsService:             jobService,
		JobsProcessor:           jobProcessor,
		ExternalLocationAddress: config.HTTPExternalLocationAddress,
	})

	srv := &http.Server{
		Addr:    server.Addr,
		Handler: server,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		// Trap sigterm or interrupt and gracefully shutdown the server
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		sig := <-sigint
		logger.Infof("Got signal: %s", sig)
		// We received a signal, shut down.

		// Gracefully shutdown the server, waiting max 'timeout' seconds for current operations to complete
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.HTTPShutdownTimeoutInSeconds)*time.Second)
		defer cancel()

		logger.Info("Shutting down HTTP server...")
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			logger.Errorw("HTTP server Shutdown", "error", err)
		}
		logger.Info("HTTP server shutdown finished successfully")

		// Close connection to external channel service
		logger.Info("Closing ChannelDownloader client")
		if err := channelDownloader.Close(); err != nil {
			logger.Error("error closing ChannelDownloader client", zap.Error(err))
		}

		// Close connection to external user service
		logger.Info("Closing UserDownloader client")
		if err := userDownloader.Close(); err != nil {
			logger.Error("error closing UserDownloader client", zap.Error(err))
		}

		// Close connection to external ticket service
		logger.Info("Closing TicketDownloader client")
		if err := ticketDownloader.Close(); err != nil {
			logger.Error("error closing TicketDownloader client", zap.Error(err))
		}

		close(idleConnsClosed)
	}()

	// Start the server
	logger.Infof("Starting HTTP server at %s", server.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		logger.Fatalw("HTTP server ListenAndServe", "error", err)
	}

	// Block until a signal is received and graceful shutdown completed.
	<-idleConnsClosed

	logger.Info("Exiting")
	_ = logger.Sync()
}
