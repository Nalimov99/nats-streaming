package handlers

import (
	"nats-server/cmd/sub/internal/subscription"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type OrderAPI struct {
	OrderSubscription *subscription.OrderSubscription
	DB                *sqlx.DB
}

func (o *OrderAPI) RetriveOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := o.OrderSubscription.Get(id)
	if err != nil {
		w.Write([]byte("Nothing found"))
		return
	}

	t, _ := template.ParseFiles("templates/order/item.html")
	t.Execute(w, string(*res))
}

func (o *OrderAPI) OrderList(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, o.OrderSubscription.Cache.GetItems())
}
