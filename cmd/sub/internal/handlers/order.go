package handlers

import (
	"nats-server/cmd/sub/internal/subscription"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderAPI struct {
	OrderSubscription *subscription.OrderSubscription
}

func (o *OrderAPI) RetriveOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := o.OrderSubscription.Get(id)
	if err != nil {
		w.Write([]byte("Nothing found"))
		return
	}

	w.Write(*res)
}
