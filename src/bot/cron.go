package bot

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

// Cron makes matches for pairing, and messages those people to notify them of their match
// it runs once per day at 8am (it's triggered with app engine's Cron service)
func Cron(w http.ResponseWriter, r *http.Request) {
	// Check that the request is originating from within app engine
	// https://cloud.google.com/appengine/docs/flexible/go/scheduling-jobs-with-cron-yaml#validating_cron_requests
	if r.Header.Get("X-Appengine-Cron") != "true" {
		http.NotFound(w, r)
		return
	}

	// setting up database connection
	ctx := context.Background()
	var err error
	client, err = firestore.NewClient(ctx, "algobot-308118")
	defer client.Close()

	if err != nil {
		log.Panic(err)
	}

	switch hour := time.Now().Hour(); hour {
	case 9:
		MessagePairs(client, ctx)
	case 11:
		MessageSolo(client, ctx)
	case 13:
		PostDaily(client, ctx)
	default:
		log.Fatal("Something is up with cron; this shouldn't be running! Check out your YAML")
	}
}
