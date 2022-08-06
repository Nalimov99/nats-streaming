package handlers

import (
	"nats-server/cmd/sub/internal/subscription"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// API knows how to initialize chi.Mux for the server
func API(o *subscription.OrderSubscription, db *sqlx.DB) *chi.Mux {
	routes := chi.NewRouter()

	orderAPI := OrderAPI{
		OrderSubscription: o,
		DB:                db,
	}
	routes.Get("/order/{id}", orderAPI.RetriveOrder)
	routes.Get("/", orderAPI.List)

	return routes
}
