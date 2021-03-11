package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// It's alive! The application starts here.
func main() {
	http.HandleFunc("/", nope)
	http.HandleFunc("/webhooks", handle)
	http.HandleFunc("/cron", cron)
	http.HandleFunc("/config", config)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
