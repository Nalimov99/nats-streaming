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
func List(db *sqlx.DB) (map[string][]byte, error) {
	q := `SELECT order_uid, data FROM orders;`
	orders := []OrderDBRow{}
	if err := db.Select(&orders, q); err != nil {
		return nil, err
	}

	items := make(map[string][]byte)
	for _, item := range orders {
		items[item.Order_uid] = item.Data
	}

	return items, nil
}
