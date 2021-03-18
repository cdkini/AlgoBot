package bot

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/fatih/structs"
)

func MessageSolo(client *firestore.Client, ctx context.Context) {
	today := strings.ToLower(time.Now().Weekday().String())[:3]

	iter := client.Collection("recursers").
		Where("isSkippingTomorrow", "==", false).
		Where("config.soloDays", "array-contains", today).
		Documents(ctx)

	recursersList := iterToRecurserList(iter)

	// if for some reason there's no matches today, we're done
	if len(recursersList) == 0 {
		log.Println("No one was signed up for solo sessions today")
		return
	}

	// message the peeps!
	doc, err := client.Collection("auth").Doc("api").Get(ctx)
	if err != nil {
		log.Panic(err)
	}

	apikey := doc.Data()["key"]
	botPassword := apikey.(string)
	zulipClient := &http.Client{}

	for i := range recursersList {
		interviewee := recursersList[i]

		messageRequest := url.Values{}
		messageRequest.Add("type", "private")
		messageRequest.Add("to", interviewee.Email)

		question := selectQuestion(interviewee, client, ctx)
		msg := fmtSoloMessage(question)
		messageRequest.Add("content", msg)

		req, err := http.NewRequest("POST", zulipAPIURL, strings.NewReader(messageRequest.Encode()))
		req.SetBasicAuth(botEmailAddress, botPassword)
		req.Header.Set("content-type", "application/x-www-form-urlencoded")

		resp, err := zulipClient.Do(req)
		if err != nil {
			log.Panic(err)
		}
		defer resp.Body.Close()

		respBodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(string(respBodyText))
			log.Println("A question went out")
		}

		session := map[string]interface{}{
			"question":  question["id"],
			"timeStamp": time.Now(),
		}

		doc := client.Collection("soloSessions").Doc(interviewee.Id)
		_, err = doc.Update(ctx, []firestore.Update{{Path: "sessions", Value: firestore.ArrayUnion(session)}})
		if err != nil {
			log.Println(err)
		} else {
			log.Println("A session was recorded")
		}
	}

	// get everyone who was set to skip today and set them back to isSkippingTomorrow = false
	iter = client.Collection("recursers").Where("isSkippingTomorrow", "==", true).Documents(ctx)

	skippersList := iterToRecurserList(iter)
	for i := range skippersList {
		skippersList[i].IsSkippingTomorrow = false
		_, err := client.Collection("recursers").Doc(skippersList[i].Id).Set(ctx, structs.Map(skippersList[i]), firestore.MergeAll)
		if err != nil {
			log.Println(err)
		}
	}
}

func fmtSoloMessage(question map[string]interface{}) string {
	var builder strings.Builder
	builder.WriteString("Hey there! I've got your next question prepared and ready to go!\n")
	builder.WriteString("The question was randomly selected based on your config and question history; use `config` to make modifications.\n\n")
	builder.WriteString(fmt.Sprintf("[Today's Question](%s)\n\n", question["url"]))
	builder.WriteString("Want even more practice? Feel free to `schedule` a mock interview or work on the daily question in #**Daily LeetCode** :)")
	return builder.String()
}
