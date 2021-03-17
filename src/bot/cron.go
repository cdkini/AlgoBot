package bot

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/fatih/structs"
	"google.golang.org/api/iterator"
)

// Cron makes matches for pairing, and messages those people to notify them of their match
// it runs once per day at 8am (it's triggered with app engine's Cron service)
func Cron(w http.ResponseWriter, r *http.Request) {
	// Check that the request is originating from within app engine
	// https://cloud.google.com/appengine/docs/flexible/go/scheduling-jobs-with-cron-yaml#validating_cron_requests
	// if r.Header.Get("X-Appengine-Cron") != "true" {
	// 	http.NotFound(w, r)
	// 	return
	// }

	// setting up database connection
	ctx := context.Background()
	var err error
	client, err = firestore.NewClient(ctx, "mock-interview-bot-307121")
	defer client.Close()

	if err != nil {
		log.Panic(err)
	}

	switch hour := time.Now().Hour(); hour {
	case 8:
		MessageSolo(client, ctx)
	case 10:
		MessagePairs(client, ctx)
	case 12:
		PostDaily(client, ctx)
	default:
		// MessageSolo(client, ctx)
		MessagePairs(client, ctx)
		// log.Fatal("Something is up with cron; this shouldn't be running! Check out your YAML")
	}
}

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

func selectQuestion(recurser Recurser, client *firestore.Client, ctx context.Context) map[string]interface{} {
	config := recurser.Config

	if config.ManualQuestion {
		return nil
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	difficulty := config.SoloDifficulty[r.Intn(len(config.SoloDifficulty))]
	query := client.Collection("questions").Where("difficulty", "==", difficulty)

	if config.Topics != nil && len(config.Topics) != 0 {
		topic := config.Topics[r.Intn(len(config.Topics))]
		query = query.Where("tags", "array-contains", topic)
	}

	if config.ProblemSet != "random" {
		query = query.Where("psets", "array-contains", config.ProblemSet)
	}

	iter := query.Documents(ctx)
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

	if len(documents) == 0 {
		return nil
	}

	selection := documents[r.Intn(len(documents))]

	return selection.Data()
}

func iterToRecurserList(iter *firestore.DocumentIterator) []Recurser {
	var recursersList []Recurser

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Panic(err)
		}

		var recurser Recurser
		if err = doc.DataTo(&recurser); err != nil {
			log.Fatal(err)
		}
		recursersList = append(recursersList, recurser)
	}

	return recursersList
}

func MessagePairs(client *firestore.Client, ctx context.Context) {
	iter := client.Collection("recursers").Where("isPairingTomorrow", "==", true).Documents(ctx)
	recursersList := iterToRecurserList(iter)

	// if for some reason there's no matches today, we're done
	if len(recursersList) == 0 {
		log.Println("No one was signed up to pair today -- so there were no matches")
		return
	}

	// shuffle our recursers. This will not error if the list is empty
	shuffle(recursersList)

	// optimalPath := determineBestPath(recursersList)
	// recursersList = optimalPath.order

	// message the peeps!
	doc, err := client.Collection("auth").Doc("api").Get(ctx)
	if err != nil {
		log.Panic(err)
	}

	apikey := doc.Data()["key"]
	botPassword := apikey.(string)
	zulipClient := &http.Client{}

	// if there's an odd number today, message the last person in the list
	// and tell them they don't get a match today, then knock them off the list
	if len(recursersList)%2 != 0 {
		recurser := recursersList[len(recursersList)-1]
		recursersList = recursersList[:len(recursersList)-1]
		log.Println("Someone was the odd-one-out today")
		messageRequest := url.Values{}
		messageRequest.Add("type", "private")
		messageRequest.Add("to", recurser.Email)
		messageRequest.Add("content", botMessages.OddOneOut)
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
		}
		log.Println(string(respBodyText))
	}

	// Send out messages notifying pairs that they've been matched
	for i := 0; i < len(recursersList); i += 2 {
		messageRequest := url.Values{}
		messageRequest.Add("type", "private")
		messageRequest.Add("to", recursersList[i].Email+", "+recursersList[i+1].Email)
		messageRequest.Add("content", botMessages.Matched)
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
		}
		log.Println(string(respBodyText))
		log.Println("A match went out")
	}

	// Send private messages to each individual about the question they should prepare for their partner
	for i := range recursersList {
		interviewer := recursersList[i]
		var interviewee Recurser

		// Interviews go both ways so we need to ensure interviewers become interviewees and vice versa
		if i%2 == 0 {
			interviewee = recursersList[i+1]
		} else {
			interviewee = recursersList[i-1]
		}

		messageRequest := url.Values{}
		messageRequest.Add("type", "private")
		messageRequest.Add("to", interviewer.Email)

		question := selectQuestion(interviewee, client, ctx)
		msg := fmtInterviewerMessage(question, interviewee)
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
			"interviewer": interviewer.Id,
			"question":    question["id"],
			"timeStamp":   time.Now(),
		}

		doc := client.Collection("pairingSessions").Doc(interviewee.Id)
		_, err = doc.Update(ctx, []firestore.Update{{Path: "sessions", Value: firestore.ArrayUnion(session)}})
		if err != nil {
			log.Println(err)
		} else {
			log.Println("A session was recorded")
		}

		// Upon having an interview, kick out of queue
		// We require manual sign-ups to prevent people from forgetting and ruining someone else's prep
		doc = client.Collection("recursers").Doc(interviewee.Id)
		doc.Update(ctx, []firestore.Update{{Path: "isPairingTomorrow", Value: false}})
	}
}

func fmtInterviewerMessage(question map[string]interface{}, interviewee Recurser) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Here's what you need to know as the interviewer when you pair with %s:\n\n", interviewee.Name))
	builder.WriteString(fmt.Sprintf("[Your question to prepare](%s)\n", question["url"]))
	builder.WriteString("Try to learn multiple solutions, starting from brute force and ending with the optimal algorithm.\n\n")
	builder.WriteString(fmt.Sprintf("Please conduct the interview on %s.\n\n", interviewee.Config.Environment))
	builder.WriteString(fmt.Sprintf("Here are some additional notes from your interviewee: %s\n\n", interviewee.Config.Comments))
	builder.WriteString("Whether you're a pro at interviews or are just getting started, please read over [the guidelines](http://github.com/cdkini) before your session! Thanks :)")
	return builder.String()
}

type Path struct {
	order      []Recurser
	validPairs int
}

func determineBestPath(recursers []Recurser) Path {
	bestPath := new(Path)
	stack := make([]Path, 0)
	for _, recurser := range recursers {
		stack = append(stack, Path{[]Recurser{recurser}, 0})
	}

	wg := new(sync.WaitGroup)
	bestPossibleScore := len(recursers) / 2

	for i := range stack {
		wg.Add(1)
		go func(path Path) {
			getNext(path, recursers, map[string]bool{path.order[0].Id: true}, bestPath, bestPossibleScore, wg)
			defer wg.Done()
		}(stack[i])
	}
	wg.Wait()

	return *bestPath
}

func getNext(path Path, recursers []Recurser, seen map[string]bool, bestPath *Path, bestPossibleScore int, wg *sync.WaitGroup) {
	if bestPath.validPairs == bestPossibleScore {
		return
	}

	if len(path.order)%2 == 0 {
		if isValidSoFar(path.order) {
			path.validPairs++
		}
	}

	if len(path.order) == len(recursers) && path.validPairs > bestPath.validPairs {
		*bestPath = path
		return
	}

	for _, recurser := range recursers {
		if _, ok := seen[recurser.Id]; ok {
			continue
		}

		pathCopy := path
		pathCopy.order = make([]Recurser, len(path.order))
		copy(pathCopy.order, path.order)
		pathCopy.order = append(pathCopy.order, recurser)

		seenCopy := make(map[string]bool, 0)
		for k, v := range seen {
			seenCopy[k] = v
		}
		seenCopy[recurser.Id] = true

		wg.Add(1)
		go func() {
			getNext(pathCopy, recursers, seenCopy, bestPath, bestPossibleScore, wg)
			defer wg.Done()
		}()

	}
}

func isValidSoFar(path []Recurser) bool {
	recurserOne := path[len(path)-2]
	recurserTwo := path[len(path)-1]

	difficulties := map[string]int{
		"easy":   0,
		"medium": 1,
		"hard":   2,
	}

	return min(recurserOne.Config.PairingDifficulty, difficulties) <= difficulties[recurserTwo.Config.Experience] &&
		min(recurserTwo.Config.PairingDifficulty, difficulties) <= difficulties[recurserOne.Config.Experience]
}

func PostDaily(client *firestore.Client, ctx context.Context) {
	var difficulty string

	switch today := strings.ToLower(time.Now().Weekday().String())[:3]; today {
	case "mon":
		difficulty = "easy"
	case "tue":
		difficulty = "easy"
	case "wed":
		difficulty = "medium"
	case "thu":
		difficulty = "medium"
	case "fri":
		difficulty = "hard"
	default:
		log.Fatal("Something is up with cron; this shouldn't be running! Check out your YAML")
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

	selection := documents[rand.Intn(len(documents))]
	fmt.Println(selection)
}
