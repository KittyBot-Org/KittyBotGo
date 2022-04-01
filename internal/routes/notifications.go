package routes

import (
	"context"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend"
	"net/http"

	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
)

func NotificationsHandler(b *backend.Backend) http.HandlerFunc {
	sub, _ := b.PubSubClient.CreateSubscription(context.TODO(), "notifications", pubsub.SubscriptionConfig{
		Topic: b.PubSubClient.Topic("notifications"),
	})

	sub.Receive(context.TODO(), func(ctx context.Context, msg *pubsub.Message) {

	})
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		_, ok := params["service"]
		if !ok {
			http.NotFound(w, r)
			return
		}

	}
}
