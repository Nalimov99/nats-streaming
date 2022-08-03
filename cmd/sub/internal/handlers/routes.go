package handlers

import (
	"nats-server/cmd/sub/internal/subscription"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func API(o *subscription.OrderSubscription) *chi.Mux {
	routes := chi.NewRouter()

	routes.Get("/order/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		res, _ := o.Get(id)

		w.Write(*res)
	})

	return routes
}
