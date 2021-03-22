package bot

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func PostDaily(client *firestore.Client, ctx context.Context) {
	t := time.Now()
	today := fmt.Sprintf("%v-%v-%v", t.Month(), t.Day(), t.Year())

	question := generateDailyQuestion(today, ctx)

	session := map[string]interface{}{
		"question":  question["id"],
		"timeStamp": time.Now(),
	}

	_, err := client.Collection("dailyQuestions").Doc(today).Create(ctx, session)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("A daily question was recorded")
	}

	doc, err := client.Collection("auth").Doc("api").Get(ctx)
	if err != nil {
		log.Panic(err)
	}

	apikey := doc.Data()["key"]

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("**AlgoBot Daily Question (%s):**\n\n", today))
	builder.WriteString(fmt.Sprintf("[%v. %s](%s) [%s]\n\n", question["id"], question["name"], question["url"], strings.Title(question["difficulty"].(string))))
	builder.WriteString("Feel free to post your answers below (but take care to add spoilers!).\n")
	builder.WriteString(fmt.Sprintf("Problems get more difficult as the week progresses. Check out [the schedule](%s#daily-questions)!\n\n", githubURL))

	builder.WriteString("Send me a DM to create a study schedule and practice mock interviews.")

	msg := builder.String()
	stream := "Daily LeetCode"
	topic := "AlgoBot Daily Question"

	// Directly run curl as noted in the Zulip API docs
	cmd := exec.Command(
		"curl", "-X", "POST", zulipAPIURL,
		"-u", fmt.Sprintf("%s:%s", botEmailAddress, apikey),
		"--data-urlencode", "type=stream",
		"--data-urlencode", fmt.Sprintf("to=%s", stream),
		"--data-urlencode", fmt.Sprintf("subject=%s", topic),
		"--data-urlencode", fmt.Sprintf("content=%s", msg),
	)
	fmt.Println(cmd)

	err = cmd.Run()

	if err != nil {
		log.Println(err)
	} else {
		log.Println("A daily question was sent out")
	}
}

func generateDailyQuestion(today string, ctx context.Context) map[string]interface{} {
	var difficulty string
	switch day := strings.ToLower(time.Now().Weekday().String())[:3]; day {
	case "mon":
		difficulty = "easy"
	case "tue":
		difficulty = "easy"
	case "wed":
		difficulty = "medium"
	case "thu":
		difficulty = "medium"
	case "fri":
		difficulty = "medium"
	case "sat":
		difficulty = "hard"
	case "sun":
		difficulty = "hard"
	}

	iter := client.Collection("questions").Where("difficulty", "==", difficulty).Documents(ctx)
	var documents []*firestore.DocumentSnapshot
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		documents = append(documents, doc)
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	selection := documents[r.Intn(len(documents))]
	question := selection.Data()

	return question
}
