package handlers

import (
	"nats-server/cmd/sub/internal/subscription"

	"github.com/go-chi/chi/v5"
)

// API knows how to initialize chi.Mux for the server
func API(o *subscription.OrderSubscription) *chi.Mux {
	routes := chi.NewRouter()

	orderAPI := OrderAPI{
		OrderSubscription: o,
	}
	routes.Get("/order/{id}", orderAPI.RetriveOrder)

	return routes
}
