package main

import (
	"nats-server/cmd/sub/internal/handlers"
	"nats-server/cmd/sub/internal/subscription"
	"nats-server/internal/platform/cache"
	"nats-server/internal/platform/database"
	"net/http"

	"github.com/nats-io/stan.go"
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
	// Setup Order
	order := subscription.OrderSubscription{
		Cache:  cache.NewCache(),
		Logger: logger,
		DB:     db,
	}

	// =======================================================
	// Stan connection
	sc, err := stan.Connect("nats-streaming", "sub", stan.NatsURL(":14222"))
	if err != nil {
		logger.Panic("Failed to connect nats-streaming: ", zap.Error(err))
	}
	defer sc.Close()

	_, err = sc.Subscribe("orders", func(msg *stan.Msg) {
		if err := order.AddOrUpdate(msg.Data); err != nil {
			logger.Info("Upserting error",
				zap.Error(err),
			)
		}
	})
	if err != nil {
		logger.Panic("Could not subscribe to the orders subject", zap.Error(err))
	}

	// =======================================================
	// Start API service
	api := http.Server{
		Addr:    ":3020",
		Handler: handlers.API(&order),
	}
	api.ListenAndServe()
}
