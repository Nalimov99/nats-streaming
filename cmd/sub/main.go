package main

import (
	"nats-server/cmd/sub/internal/handlers"
	"nats-server/cmd/sub/internal/subscription"
	"nats-server/internal/platform/database"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	// =======================================================
	// Setup logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// =======================================================
	// Open DB
	db, err := database.Open()
	if err != nil {
		logger.Panic("DB openning", zap.Error(err))
	}

	// =======================================================
	// Setup orderSubscription
	orderSubscription := subscription.NewOrderSubscription(logger, db)

	// =======================================================
	// Start API service
	api := http.Server{
		Addr:    ":3020",
		Handler: handlers.API(orderSubscription),
	}
	api.ListenAndServe()
}
