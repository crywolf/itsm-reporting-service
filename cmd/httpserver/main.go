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
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job/processor"
	jobsvc "github.com/KompiTech/itsm-reporting-service/internal/domain/job/service"
	ticketdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/ticket/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
	userdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/user/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest"
	"github.com/KompiTech/itsm-reporting-service/internal/repository/memory"
	"github.com/spf13/viper"
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

	loadEnvConfiguration()

	clock := realClock{}
	jobRepository := memory.NewJobRepositoryMemory(clock)
	jobService := jobsvc.NewJobService(jobRepository)

	tokenSvcClient := client.NewTokenSvcClient()

	channelRepository := memory.NewChannelRepositoryMemory()
	channelClient := chandownloader.NewChannelClient(client.NewHTTPClient(viper.GetString("ChannelEndpointURI"), logger, tokenSvcClient))
	channelDownloader := chandownloader.NewChannelDownloader(channelRepository, channelClient)

	userRepository := memory.NewUserRepositoryMemory()
	userClient := userdownloader.NewUserClient(client.NewHTTPClient(viper.GetString("UserEndpointURI"), logger, tokenSvcClient))
	userDownloader := userdownloader.NewUserDownloader(channelRepository, userRepository, userClient)

	ticketRepository := memory.NewTicketRepositoryMemory()
	ticketClient := ticketdownloader.NewTicketClient(
		client.NewHTTPClient(viper.GetString("IncidentEndpointURI"), logger, tokenSvcClient),
		client.NewHTTPClient(viper.GetString("RequestEndpointURI"), logger, tokenSvcClient),
	)
	ticketDownloader := ticketdownloader.NewTicketDownloader(channelRepository, userRepository, ticketRepository, ticketClient)

	jobProcessor := jobprocessor.NewJobProcessor(logger, jobRepository, channelDownloader, userDownloader, ticketDownloader)

	// HTTP server
	server := rest.NewServer(rest.Config{
		Addr:                    viper.GetString("HTTPBindAddress"),
		URISchema:               "http://",
		Logger:                  logger,
		JobsService:             jobService,
		JobsProcessor:           jobProcessor,
		ExternalLocationAddress: viper.GetString("ExternalLocationAddress"),
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
		timeout := viper.GetInt("HTTPShutdownTimeoutInSeconds")
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
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
