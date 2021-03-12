package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type UserConfig struct {
	comments          string
	environment       string
	experience        string
	questionList      string
	topics            []string
	soloDays          []string
	soloDifficulty    []string
	pairingDifficulty []string
	manualQuestion    bool
}

func Config(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "static/templates/config.html")
	case "POST":
		id := strings.TrimPrefix(r.URL.Path, "/config/")
		handlePOST(w, r, id)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func handlePOST(w http.ResponseWriter, r *http.Request, id string) {
	// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "mock-interview-bot-307121")
	if err != nil {
		log.Panic(err)
	}

	// Store results of POST in struct
	config := UserConfig{
		r.PostFormValue("comments"),
		r.PostFormValue("environment"),
		r.PostFormValue("experience"),
		r.PostFormValue("questionList"),
		r.PostForm["topics"],
		r.PostForm["soloDays"],
		r.PostForm["soloDifficulty"],
		r.PostForm["pairingDifficulty"],
		r.PostFormValue("manualQuestion") == "manualQuestion",
	}

	// Retrieve current config / user profile
	doc, err := client.Collection("recursers").Doc(id).Get(ctx)
	if err != nil && grpc.Code(err) != codes.NotFound {
		log.Panic(err)
	}

	// Update to new config / user profile
	recurser := doc.Data()
	recurser["config"] = config
	_, err = client.Collection("recursers").Doc(id).Set(ctx, recurser)
}
