package order

import (
	"github.com/jmoiron/sqlx"
)

// Add insert order into database
func Add(db *sqlx.DB, key string, data []byte) error {
	q := `
		INSERT INTO orders
		(order_uid, data)
		VALUES
		($1, $2);
	`

	_, err := db.Exec(q, key, data)

	return err
}

// List returns all known Orders
func List(db *sqlx.DB) (map[string]Order, error) {
	q := `SELECT order_uid, data FROM orders;`
	orders := make(map[string]Order)
	if err := db.Select(orders, q); err != nil {
		return nil, err
	}

	return orders, nil
}
