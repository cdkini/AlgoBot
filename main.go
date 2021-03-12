package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cdkini/recurse-mock-interview-bot/src/bot"
)

// It's alive! The application starts here.
func main() {
	http.HandleFunc("/", bot.Nope)
	http.HandleFunc("/webhooks", bot.Handle)
	http.HandleFunc("/cron", bot.Cron)
	http.HandleFunc("/config/", bot.Config)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
