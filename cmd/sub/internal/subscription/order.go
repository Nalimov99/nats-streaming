package subscription

import (
	"errors"
	"nats-server/internal/order"
	"nats-server/internal/platform/cache"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var (
	ErrInvalidItem = errors.New("Item is not valid")
	ErrInvalidID   = errors.New("Item ID is not valid")
)

type OrderSubscription struct {
	Cache  *cache.Cache
	Logger *zap.Logger
	DB     *sqlx.DB
}

// isValid knows how to unmarshal bytes to the Order struct.
// It will check validity of struct, if struct is invalid it will error
func (o *OrderSubscription) isValid(msg []byte) (*order.Order, error) {
	var item order.Order
	if err := json.Unmarshal(msg, &item); err != nil {
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

	if _, err = o.Get(item.OrderUID); err != nil {
		return o.add(item.OrderUID, msg)
	}

	o.update(item)
	return nil
}

// add knows how to set item in the Cache
func (o *OrderSubscription) add(key string, data []byte) error {
	o.Cache.Set(key, data)
	return order.Add(o.DB, key, data)
}

// update knows how to update item in the Cache
func (o *OrderSubscription) update(order *order.Order) {

}

// Get knows how to retrieve item from Cache by key. It will error if the
// specified key does not reference the existing item
func (o *OrderSubscription) Get(key string) (*[]byte, error) {
	return o.Cache.Get(key)
}