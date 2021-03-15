package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cdkini/recurse-mock-interview-bot/src/bot"
	"github.com/gorilla/mux"
)

// It's alive! The application starts here.
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/webhooks", bot.Webhook)
	r.HandleFunc("/cron", bot.Cron)
	r.HandleFunc("/config/{id}", bot.Config)

	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
