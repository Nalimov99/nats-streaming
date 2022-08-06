package subscription

import (
	"errors"
	"nats-server/internal/order"
	"nats-server/internal/platform/cache"

	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

var validate *validator.Validate

var (
	ErrInvalidItem = errors.New("Item is not valid")
	ErrInvalidID   = errors.New("Item ID is not valid")
)

func init() {
	validate = validator.New()
}

type OrderSubscription struct {
	Cache  *cache.Cache
	Logger *zap.Logger
	DB     *sqlx.DB
}

// NewOrderSubcription know how to initilize OrderSubscription
func NewOrderSubscription(logger *zap.Logger, db *sqlx.DB) (*OrderSubscription, stan.Conn) {
	items := make(map[string][]byte)
	if dbItems, err := order.List(db); err == nil {
		items = dbItems
	}

	orderSubscription := OrderSubscription{
		Cache:  cache.NewCache(items),
		Logger: logger,
		DB:     db,
	}

	sc, err := stan.Connect("nats-streaming", "sub", stan.NatsURL(":14222"))
	if err != nil {
		logger.Panic("Failed to connect nats-streaming: ", zap.Error(err))
	}

	_, err = sc.Subscribe("orders", func(msg *stan.Msg) {
		if err := orderSubscription.AddOrUpdate(msg.Data); err != nil {
			logger.Info("Upserting error",
				zap.Error(err),
			)
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		logger.Panic("Could not subscribe to the orders subject", zap.Error(err))
	}

	return &orderSubscription, sc
}

// isValid knows how to unmarshal bytes to the Order struct.
// It will check validity of struct, if struct is invalid it will error
func (o *OrderSubscription) isValid(msg []byte) (*order.Order, error) {
	var item order.Order
	if err := json.Unmarshal(msg, &item); err != nil {
		return nil, ErrInvalidItem
	}

	if err := validate.Struct(item); err != nil {
		return nil, ErrInvalidItem
	}

	if _, err := uuid.Parse(item.OrderUID); err != nil {
		return nil, ErrInvalidID
	}

	return &item, nil
}

// AddOrUpdate knows how to store Order item in the Cache.
// It will parse bytes to the Order struct.
// It will ignore message if value is invalid
func (o *OrderSubscription) AddOrUpdate(msg []byte) error {
	item, err := o.isValid(msg)
	if err != nil {
		o.Logger.Error("Struct is invalid",
			zap.Error(err),
			zap.String("message", string(msg)),
		)
		return nil
	}

	if _, err = o.Cache.Get(item.OrderUID); err != nil {
		return o.add(item.OrderUID, msg)
	}

	return o.update(item.OrderUID, msg)
}

// add knows how to set item in the Cache
func (o *OrderSubscription) add(key string, data []byte) error {
	o.Cache.Set(key, data)
	return order.Add(o.DB, key, data)
}

// update knows how to update item in the Cache
func (o *OrderSubscription) update(key string, data []byte) error {
	o.Cache.Set(key, data)
	return order.Update(o.DB, key, data)
}

// Get knows how to retrieve item from Cache by key. It will error if the
// specified key does not reference the existing item
func (o *OrderSubscription) Get(key string) (*[]byte, error) {
	return o.Cache.Get(key)
}
