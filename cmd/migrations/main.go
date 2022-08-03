package main

import (
	"nats-server/internal/platform/database"
	"nats-server/internal/schema"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	db, err := database.Open()
	if err != nil {
		logger.Panic("Opening DB: ", zap.Error(err))
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		logger.Panic("applying migrations", zap.Error(err))
	}

	logger.Info("migrations complete")
}
