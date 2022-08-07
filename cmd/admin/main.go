package main

import (
	"flag"
	"fmt"
	"nats-server/internal/config"
	"nats-server/internal/platform/database"
	"nats-server/internal/schema"
	"nats-server/internal/subscription/seed"
)

func main() {
	cfg := config.GetConfig(false)

	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		if err := Migrate(cfg.DB); err != nil {
			fmt.Print("applying migrations: ", err)
			break
		}
		fmt.Print("migrations complete")
	case "seed":
		if _, err := seed.Seed(cfg.Nats); err != nil {
			fmt.Println("seeding: ", err)
			break
		}

		fmt.Println("seed complete")
	default:
		fmt.Println("No args passed")
	}
}

func Migrate(cfg config.DbConfig) error {
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		return err
	}

	return nil
}
