package subscription_test

import (
	"bytes"
	"encoding/json"
	"nats-server/cmd/sub/internal/subscription"
	"nats-server/internal/config"
	"nats-server/internal/order"
	"nats-server/internal/platform/database"
	"nats-server/internal/schema"
	"nats-server/internal/subscription/seed"
	"os/exec"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

func TestOrders(t *testing.T) {
	tests := OrderTests{
		Config: config.GetConfig(true),
	}

	tests.startNatsContainer(t)
	tests.startDb(t)
	tests.initializeOrderSubscription(t)
	t.Cleanup(func() {
		tests.teardown(t)
	})

	t.Log("RUN ORDER TESTS")
	t.Run("Publish", tests.Publish)
	t.Run("Check items in Cache", tests.CheckPublished)
	t.Run("Check items in DB", tests.CheckDB)
}

// OrderTests holds methods for each product subtest
// These type allows passing dependencies for tests
type OrderTests struct {
	NatsContainerID string
	DbContainerID   string
	Config          *config.Config
	DB              *sqlx.DB
	StanConnection  stan.Conn
	Subscription    *subscription.OrderSubscription
	PublishedItems  [][]byte
	PublishedMap    map[string][]byte
}

// startNatsContainer knows how to initilize new docker container with nats
func (o *OrderTests) startNatsContainer(t *testing.T) {
	t.Helper()

	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"-p", o.Config.Nats.Port+":4222",
		"nats-streaming:0.24.6",
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	t.Log("Starting container")
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container: %v", err)
	}
	t.Log("Container started")

	o.NatsContainerID = out.String()[:12]
}

// initializeOrderSubscription knows how to initilize subscription to stan connection
func (o *OrderTests) initializeOrderSubscription(t *testing.T) {
	logger, _ := zap.NewProduction()
	sub, sc := subscription.NewOrderSubscription(logger, o.DB, o.Config.Nats)
	o.Subscription = sub
	o.StanConnection = sc
}

// startDb knows how to initilize new docker container with nats
func (o *OrderTests) startDb(t *testing.T) {
	t.Helper()

	cmd := exec.Command(
		"docker", "run",
		"-d",
		"-p", o.Config.DB.Port+":5432",
		"-e", "POSTGRES_PASSWORD="+o.Config.DB.Password,
		"-e", "POSTGRESQL_USER="+o.Config.DB.User,
		"postgres:14.1-alpine",
	)
	var out bytes.Buffer
	cmd.Stdout = &out

	t.Log("Starting DB container")
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start db container: %v.", err)
	}
	t.Log("DB container success")

	o.DbContainerID = out.String()[:12]

	db, err := database.Open(o.Config.DB)
	if err != nil {
		t.Fatalf("Could not open DB: %v", err)
	}
	o.DB = db

	t.Log("waiting for db ready")

	var pingError error
	// wait for db to be ready
	for attempts := 1; attempts < 20; attempts++ {
		pingError = db.Ping()

		if pingError == nil {
			break
		}

		time.Sleep(time.Second * time.Duration(attempts))
	}

	if pingError != nil {
		t.Fatalf("db is not ready: %v", pingError)
	}

	if err := schema.Migrate(db); err != nil {
		t.Fatalf("could not migrate, %v", err)
	}

	t.Log("Db ready")
}

// deleteContainer knows how to shutdown and delete container
func (o *OrderTests) deleteContainer(t *testing.T, containers ...string) {
	t.Helper()

	for _, container := range containers {
		cmd := exec.Command("docker", "stop", container)

		t.Log("Stoping container")
		if err := cmd.Run(); err != nil {
			t.Fatalf("could not stop container: %v, id: %s", err, container)
		}
		t.Log("Container stopped")

		cmd = exec.Command("docker", "rm", container)
		t.Log("Deleting container")
		if err := cmd.Run(); err != nil {
			t.Fatalf("could not delete container: %v, id: %s", err, container)
		}
		t.Log("Container deleted")
	}
}

// teardown knows how to cleanup tests
func (o *OrderTests) teardown(t *testing.T) {
	t.Helper()

	o.deleteContainer(t, o.DbContainerID, o.NatsContainerID)
	o.DB.Close()
	o.StanConnection.Close()
}

func (o *OrderTests) Publish(t *testing.T) {
	t.Helper()

	sc, err := stan.Connect(o.Config.Nats.ClusterID, "pub1", stan.NatsURL(o.Config.Nats.Port))
	if err != nil {
		t.Fatalf("could not connect, %v", err)
	}
	defer sc.Close()

	t.Log("Seeding")
	items, err := seed.Seed(o.Config.Nats)
	if err != nil {
		t.Fatalf("could not seed: %v", err)
	}
	t.Log("Done seeding")

	// Failure
	sc.Publish("orders", []byte(`
		Hello world
	`))

	sc.Publish("orders", []byte(`
		{
			"order_uid": "2d6c8abc-5a79-4220-8ffc-c5c24fd9630723323",
			"track_number": "WBILMTESTTRACK",
			"entry": "WBIL",
			"delivery": {
				"name": "Test Testov",
				"phone": "+9720000000",
				"zip": "2639809",
				"city": "Kiryat Mozkin",
				"address": "Ploshad Mira 15",
				"region": "Kraiot",
				"email": "test@gmail.com"
			},
			"payment": {
				"transaction": "b563feb7b2b84b6test",
				"request_id": "",
				"currency": "USD",
				"provider": "wbpay",
				"amount": 1817,
				"payment_dt": 1637907727,
				"bank": "alpha",
				"delivery_cost": 1500,
				"goods_total": 317,
				"custom_fee": 0
			},
			"items": [
				{
				"chrt_id": 9934930,
				"track_number": "WBILMTESTTRACK",
				"price": 453,
				"rid": "ab4219087a764ae0btest",
				"name": "Mascaras",
				"sale": 30,
				"size": "0",
				"total_price": 317,
				"nm_id": 2389212,
				"brand": "Vivienne Sabo",
				"status": 202
				}
			],
			"locale": "en",
			"internal_signature": "",
			"customer_id": "test",
			"delivery_service": "meest",
			"shardkey": "9",
			"sm_id": 99,
			"date_created": "2021-11-26T06:22:19Z",
			"oof_shard": "1"
		}
	`))

	sc.Publish("orders", []byte(`
		{
			"order_uid": "2d6c8abc-5a79-4220-8ffc-c5c24fd96307",
			"delivery": {
				"name": "Test Testov",
				"phone": "+9720000000",
				"zip": "2639809",
				"city": "Kiryat Mozkin",
				"address": "Ploshad Mira 15",
				"region": "Kraiot",
				"email": "test@gmail.com"
			},
			"payment": {
				"transaction": "b563feb7b2b84b6test",
				"request_id": "",
				"currency": "USD",
				"provider": "wbpay",
				"amount": 1817,
				"payment_dt": 1637907727,
				"bank": "alpha",
				"delivery_cost": 1500,
				"goods_total": 317,
				"custom_fee": 0
			},
			"items": [
				{
				"chrt_id": 9934930,
				"track_number": "WBILMTESTTRACK",
				"price": 453,
				"rid": "ab4219087a764ae0btest",
				"name": "Mascaras",
				"sale": 30,
				"size": "0",
				"total_price": 317,
				"nm_id": 2389212,
				"brand": "Vivienne Sabo",
				"status": 202
				}
			],
			"locale": "en",
			"internal_signature": "",
			"customer_id": "test",
			"delivery_service": "meest",

			"sm_id": 99,
			"date_created": "2021-11-26T06:22:19Z",
			"oof_shard": "1"
		}
	`))

	o.PublishedItems = items
	// Should wait until stan delivers all data
	time.Sleep(2 * time.Second)
}

func (o *OrderTests) CheckPublished(t *testing.T) {
	items := o.Subscription.Cache.GetItems()
	keyLenght := 0
	for range items {
		keyLenght++
	}

	if keyLenght != 3 {
		t.Fatalf("expected length should be 3, but was: %d", keyLenght)
	}

	publishedMap := make(map[string][]byte)

	orderUID := struct {
		ID string `json:"order_uid"`
	}{}
	for _, item := range o.PublishedItems {
		if err := json.Unmarshal(item, &orderUID); err != nil {
			t.Fatalf("could not get order_uid: %v", err)
		}

		publishedMap[orderUID.ID] = item
	}

	for key := range items {
		if bytes.Compare(items[key], publishedMap[key]) != 0 {
			t.Fatalf("Order in cache and published order are not equal, id: %s", key)
		}
	}

	o.PublishedMap = publishedMap
}

func (o *OrderTests) CheckDB(t *testing.T) {
	items, err := order.List(o.DB)
	if err != nil {
		t.Fatalf("could not get list, %v", err)
	}

	for key := range items {
		var want order.Order
		var got order.Order

		if err := json.Unmarshal(items[key], &got); err != nil {
			t.Fatal("could not unmarshal")
		}

		if err := json.Unmarshal(o.PublishedMap[key], &want); err != nil {
			t.Fatal("could not unmarshal")
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatal(diff)
		}
	}
}
