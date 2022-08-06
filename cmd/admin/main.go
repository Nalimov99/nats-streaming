package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"nats-server/internal/platform/database"
	"nats-server/internal/schema"
	"os"

	"github.com/nats-io/stan.go"
)

const seedDir = "cmd/admin/seed/"

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		if err := migrate(); err != nil {
			fmt.Print("applying migrations: ", err)
			break
		}
		fmt.Print("migrations complete")
	case "seed":
		if err := seed(); err != nil {
			fmt.Println("seeding: ", err)
			break
		}

		fmt.Println("seed complete")
	default:
		fmt.Println("No args passed")
	}
}

func migrate() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		return err
	}

	return nil
}

func seed() error {
	sc, err := stan.Connect("nats-streaming", "pub", stan.NatsURL(":14222"))
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(seedDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(seedDir + file.Name())
		if err != nil {
			return err
		}

		if err := sc.Publish("orders", data); err != nil {
			return err
		}
	}

	return nil
}
