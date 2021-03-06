package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"

	"github.com/fatih/structs"
	"github.com/gorilla/mux"
)

type Recurser struct {
	Id                 string     `structs:"id" firestore:"id"`
	Name               string     `structs:"name" firestore:"name"`
	Email              string     `structs:"email" firestore:"email"`
	IsSkippingTomorrow bool       `structs:"isSkippingTomorrow" firestore:"isSkippingTomorrow"`
	IsPairingTomorrow  bool       `structs:"isPairingTomorrow" firestore:"isPairingTomorrow"`
	Config             UserConfig `structs:"config" firestore:"config"`
}

func newRecurser(id string, name string, email string) Recurser {
	return Recurser{
		Id:                 id,
		Name:               name,
		Email:              email,
		IsSkippingTomorrow: false,
		IsPairingTomorrow:  false,
		Config:             defaultUserConfig(),
	}
}

func (r Recurser) stringifyUserConfig() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("You are %s!\n\n", r.Name))
	b.WriteString(fmt.Sprintf("Your experience level is %s.\n", r.Config.Experience))
	b.WriteString(fmt.Sprintf("You are working through the %s pset.\n", r.Config.ProblemSet))

	if len(r.Config.Topics) == 0 {
		b.WriteString("You have not selected specific topics to work on.\n")
	} else {
		b.WriteString(fmt.Sprintf("You are focusing on these topics: %s\n", r.Config.Topics))
	}

	if len(r.Config.SoloDays) == 0 {
		b.WriteString("You are not scheduled for solo sessions.\n")
	} else {
		b.WriteString(fmt.Sprintf("You have solo sessions scheduled for these days: %s\n", r.Config.SoloDays))
		b.WriteString(fmt.Sprintf("You will receive questions of this difficulty: %s\n", r.Config.SoloDifficulty))
		if r.IsSkippingTomorrow {
			b.WriteString("You are set to skip tomorrow's solo session.\n")
		}
	}

	if !r.IsPairingTomorrow {
		b.WriteString("You are not in the queue for a pairing session.\n")
	} else {
		b.WriteString("You are in the queue for a pairing session.\n")
		b.WriteString(fmt.Sprintf("You will receive questions of this difficulty: %s\n", r.Config.PairingDifficulty))
		b.WriteString(fmt.Sprintf("Your preferred environment is %s.\n", r.Config.Environment))
		if r.Config.ManualQuestion {
			b.WriteString("You will be choosing your own questions for your pairing sessions.\n")
		} else {
			b.WriteString("You will receive random questions for your pairing sessions.\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}

func (r Recurser) isConfigured() bool {
	return r.hasSoloConfig() && r.hasPairingConfig()
}

func (r Recurser) hasSoloConfig() bool {
	return len(r.Config.ProblemSet) > 0 &&
		len(r.Config.SoloDays) > 0 &&
		len(r.Config.SoloDifficulty) > 0
}

func (r Recurser) hasPairingConfig() bool {
	return len(r.Config.Environment) > 0 &&
		len(r.Config.Experience) > 0 &&
		len(r.Config.ProblemSet) > 0 &&
		len(r.Config.SoloDifficulty) > 0 &&
		len(r.Config.PairingDifficulty) > 0
}

type UserConfig struct {
	Comments          string   `structs:"comments" firestore:"comments"`
	Environment       string   `structs:"environment" firestore:"environment"`
	Experience        string   `structs:"experience" firestore:"experience"`
	ProblemSet        string   `structs:"problemSet" firestore:"problemSet"`
	Topics            []string `structs:"topics" firestore:"topics"`
	SoloDays          []string `structs:"soloDays" firestore:"soloDays"`
	SoloDifficulty    []string `structs:"soloDifficulty" firestore:"soloDifficulty"`
	PairingDifficulty []string `structs:"pairingDifficulty" firestore:"pairingDifficulty"`
	ManualQuestion    bool     `structs:"manualQuestion" firestore:"manualQuestion"`
}

func defaultUserConfig() UserConfig {
	return UserConfig{
		Comments:          "N/A",
		Environment:       "leetcode",
		Experience:        "medium",
		ProblemSet:        "topInterview",
		Topics:            []string{},
		SoloDays:          []string{"mon", "tue", "wed", "thu", "fri"},
		SoloDifficulty:    []string{"easy", "medium"},
		PairingDifficulty: []string{"easy", "medium"},
		ManualQuestion:    false,
	}
}

func Config(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "static/templates/config.html")
	case "POST":
		vars := mux.Vars(r)
		handlePOST(w, r, vars["id"])
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

	// var recurser Recurser
	ctx := context.Background()
	// ctx := r.Context()
	var err error
	client, err = firestore.NewClient(ctx, "algobot-308118")
	defer client.Close()
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

	// Retrieve current config / user profile and update
	doc := client.Collection("recursers").Doc(id)
	doc.Update(ctx, []firestore.Update{{Path: "config", Value: structs.Map(config)}})
}
