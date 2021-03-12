package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var botMessages = InitMessenger("src/bot/messages.json")

const botEmailAddress = "mockinterview-bot@recurse.zulipchat.com"
const zulipAPIURL = "https://recurse.zulipchat.com/api/v1/messages"

func sanityCheck(ctx context.Context, client *firestore.Client, w http.ResponseWriter, r *http.Request) (incomingJSON, error) {
	var userReq incomingJSON
	// Look at the incoming webhook and slurp up the JSON
	// Error if the JSON from Zulip istelf is bad
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.NotFound(w, r)
		return userReq, err
	}

	// validate our zulip-bot token
	// this was manually put into the database before deployment
	doc, err := client.Collection("botauth").Doc("token").Get(ctx)
	if err != nil {
		log.Println("Something weird happened trying to read the auth token from the database")
		return userReq, err
	}
	token := doc.Data()
	if userReq.Token != token["value"] {
		http.NotFound(w, r)
		return userReq, errors.New("unauthorized interaction attempt")
	}
	return userReq, err
}

func dispatch(ctx context.Context, client *firestore.Client, cmd string, cmdArgs []string, userID string, userEmail string, userName string) (string, error) {
	var response string
	var err error
	var recurser = map[string]interface{}{
		"id":                 "string",
		"name":               "string",
		"email":              "string",
		"isSkippingTomorrow": false,
		"schedule": map[string]interface{}{
			"monday":    false,
			"tuesday":   false,
			"wednesday": false,
			"thursday":  false,
			"friday":    false,
			"saturday":  false,
			"sunday":    false,
		},
	}

	// get the users "document" (database entry) out of firestore
	// we temporarily keep it in 'doc'
	doc, err := client.Collection("recursers").Doc(userID).Get(ctx)
	// this says "if there's an error, and if that error was not document-not-found"
	if err != nil && grpc.Code(err) != codes.NotFound {
		response = botMessages.ReadError
		return response, err
	}
	// if there's a db entry, that means they were already subscribed to pairing bot
	// if there's not, they were not subscribed
	isSubscribed := doc.Exists()

	// if the user is in the database, get their current state into this map
	// also assign their zulip name to the name field, just in case it changed
	// also assign their email, for the same reason
	if isSubscribed {
		recurser = doc.Data()
		recurser["name"] = userName
		recurser["email"] = userEmail
	}
	// here's the actual actions. command input from
	// the user has already been sanitized, so we can
	// trust that cmd and cmdArgs only have valid stuff in them
	switch cmd {
	case "schedule":
		if isSubscribed == false {
			response = botMessages.NotSubscribed
			break
		}
		// create a new blank schedule
		var newSchedule = map[string]interface{}{
			"monday":    false,
			"tuesday":   false,
			"wednesday": false,
			"thursday":  false,
			"friday":    false,
			"saturday":  false,
			"sunday":    false,
		}
		// populate it with the new days they want to pair on
		for _, day := range cmdArgs {
			newSchedule[day] = true
		}
		// put it in the database
		recurser["schedule"] = newSchedule
		_, err = client.Collection("recursers").Doc(userID).Set(ctx, recurser, firestore.MergeAll)
		if err != nil {
			response = botMessages.WriteError
			break
		}
		response = "Awesome, your new schedule's been set! You can check it with `status`."

	case "subscribe":
		if isSubscribed {
			response = "You're already subscribed! Use `schedule` to set your schedule."
			break
		}

		// recurser isn't really a type, because we're using maps
		// and not struct. but we're using it *as* a type,
		// and this is the closest thing to a definition that occurs
		recurser = map[string]interface{}{
			"id":                 userID,
			"name":               userName,
			"email":              userEmail,
			"isSkippingTomorrow": false,
			"schedule": map[string]interface{}{
				"monday":    true,
				"tuesday":   true,
				"wednesday": true,
				"thursday":  true,
				"friday":    true,
				"saturday":  false,
				"sunday":    false,
			},
		}
		_, err = client.Collection("recursers").Doc(userID).Set(ctx, recurser)
		if err != nil {
			response = botMessages.WriteError
			break
		}
		response = botMessages.Subscribe

	case "unsubscribe":
		if isSubscribed == false {
			response = botMessages.NotSubscribed
			break
		}
		_, err = client.Collection("recursers").Doc(userID).Delete(ctx)
		if err != nil {
			response = botMessages.WriteError
			break
		}
		response = botMessages.Unsubscribe

	case "skip":
		if isSubscribed == false {
			response = botMessages.NotSubscribed
			break
		}
		recurser["isSkippingTomorrow"] = true
		_, err = client.Collection("recursers").Doc(userID).Set(ctx, recurser, firestore.MergeAll)
		if err != nil {
			response = botMessages.WriteError
			break
		}
		response = `Tomorrow: cancelled. I feel you. **I will not match you** for pairing tomorrow <3`

	case "unskip":
		if isSubscribed == false {
			response = botMessages.NotSubscribed
			break
		}
		recurser["isSkippingTomorrow"] = false
		_, err = client.Collection("recursers").Doc(userID).Set(ctx, recurser, firestore.MergeAll)
		if err != nil {
			response = botMessages.WriteError
			break
		}
		response = "Tomorrow: uncancelled! Heckin *yes*! **I will match you** for pairing tomorrow :)"

	case "status":
		if isSubscribed == false {
			response = botMessages.NotSubscribed
			break
		}
		// this particular days list is for sorting and printing the
		// schedule correctly, since it's stored in a map in all lowercase
		var daysList = []string{
			"Monday",
			"Tuesday",
			"Wednesday",
			"Thursday",
			"Friday",
			"Saturday",
			"Sunday"}

		// get their current name
		whoami := recurser["name"]

		// get skip status and prepare to write a sentence with it
		var skipStr string
		if recurser["isSkippingTomorrow"].(bool) {
			skipStr = " "
		} else {
			skipStr = " not "
		}

		// make a sorted list of their schedule
		var schedule []string
		for _, day := range daysList {
			// this line is a little wild, sorry. it looks so weird because we
			// have to do type assertion on both interface types
			if recurser["schedule"].(map[string]interface{})[strings.ToLower(day)].(bool) {
				schedule = append(schedule, day)
			}
		}
		// make a lil nice-lookin schedule string
		var scheduleStr string
		for i := range schedule[:len(schedule)-1] {
			scheduleStr += schedule[i] + "s, "
		}
		if len(schedule) > 1 {
			scheduleStr += "and " + schedule[len(schedule)-1] + "s"
		} else if len(schedule) == 1 {
			scheduleStr += schedule[0] + "s"
		}

		response = fmt.Sprintf("* You're %v\n* You're scheduled for pairing on **%v**\n* **You're%vset to skip** pairing tomorrow", whoami, scheduleStr, skipStr)

	case "help":
		response = botMessages.Help
	default:
		// this won't execute because all input has been sanitized
		// by parseCmd() and all cases are handled explicitly above
	}
	return response, err
}

func Handle(w http.ResponseWriter, r *http.Request) {
	responder := json.NewEncoder(w)
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "mock-interview-bot-307121")
	if err != nil {
		log.Panic(err)
	}
	// sanity check the incoming request
	userReq, err := sanityCheck(ctx, client, w, r)
	if err != nil {
		log.Println(err)
		return
	}

	if userReq.Trigger != "private_message" {
		err = responder.Encode(botResponse{"Hi! I'm Pairing Bot (she/her)!\n\nSend me a PM that says `subscribe` to get started :smiley:\n\n:pear::robot:\n:octopus::octopus:"})
		if err != nil {
			log.Println(err)
		}
		return
	}
	// if there aren't two 'recipients' (one sender and one receiver),
	// then don't respond. this stops pairing bot from responding in the group
	// chat she starts when she matches people
	if len(userReq.Message.DisplayRecipient.([]interface{})) != 2 {
		err = responder.Encode(botNoResponse{true})
		if err != nil {
			log.Println(err)
		}
		return
	}
	// you *should* be able to throw any freakin string at this thing and get back a valid command for dispatch()
	// if there are no commad arguments, cmdArgs will be nil
	cmd, cmdArgs, err := parseCmd(userReq.Data)
	if err != nil {
		log.Println(err)
	}
	// the tofu and potatoes right here y'all
	response, err := dispatch(ctx, client, cmd, cmdArgs, strconv.Itoa(userReq.Message.SenderID), userReq.Message.SenderEmail, userReq.Message.SenderFullName)
	if err != nil {
		log.Println(err)
	}
	err = responder.Encode(botResponse{response})
	if err != nil {
		log.Println(err)
	}
}

func Config(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "static/templates/config.html")
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}
func parseCmd(cmdStr string) (string, []string, error) {
	var err error
	var cmdList = []string{
		"subscribe",
		"unsubscribe",
		"help",
		"schedule",
		"skip",
		"unskip",
		"status"}

	var daysList = []string{
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
		"sunday"}

	// convert the string to a slice
	// after this, we have a value "cmd" of type []string
	// where cmd[0] is the command and cmd[1:] are any arguments
	space := regexp.MustCompile(`\s+`)
	cmdStr = space.ReplaceAllString(cmdStr, ` `)
	cmdStr = strings.TrimSpace(cmdStr)
	cmdStr = strings.ToLower(cmdStr)
	cmd := strings.Split(cmdStr, ` `)

	// Big validation logic -- hellooo darkness my old frieeend
	switch {
	// if there's nothing in the command string srray
	case len(cmd) == 0:
		err = errors.New("the user-issued command was blank")
		return "help", nil, err

	// if there's a valid command and if there's no arguments
	case contains(cmdList, cmd[0]) && len(cmd) == 1:
		if cmd[0] == "schedule" || cmd[0] == "skip" || cmd[0] == "unskip" {
			err = errors.New("the user issued a command without args, but it required args")
			return "help", nil, err
		}
		return cmd[0], nil, err

	// if there's a valid command and there's some arguments
	case contains(cmdList, cmd[0]) && len(cmd) > 1:
		switch {
		case cmd[0] == "subscribe" || cmd[0] == "unsubscribe" || cmd[0] == "help" || cmd[0] == "status":
			err = errors.New("the user issued a command with args, but it disallowed args")
			return "help", nil, err
		case cmd[0] == "skip" && len(cmd) != 2 && cmd[1] != "tomorrow":
			err = errors.New("the user issued SKIP with malformed arguments")
			return "help", nil, err
		case cmd[0] == "unskip" && len(cmd) != 2 && cmd[1] != "tomorrow":
			err = errors.New("the user issued UNSKIP with malformed arguments")
			return "help", nil, err
		case cmd[0] == "schedule":
			for _, v := range cmd[1:] {
				if contains(daysList, v) == false {
					err = errors.New("the user issued SCHEDULE with malformed arguments")
					return "help", nil, err
				}
			}
			fallthrough
		default:
			return cmd[0], cmd[1:], err
		}

	// if there's not a valid command
	default:
		err = errors.New("the user-issued command wasn't valid")
		return "help", nil, err
	}
}

func Nope(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

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
	client, err := firestore.NewClient(ctx, "mock-interview-bot-307121")
	if err != nil {
		log.Panic(err)
	}
	var recursersList []map[string]interface{}
	var skippersList []map[string]interface{}
	// this gets the time from system time, which is UTC
	// on app engine (and most other places). This works
	// fine for us in NYC, but might not if pairing bot
	// were ever running in another time zone
	today := strings.ToLower(time.Now().Weekday().String())

	// ok this is how we have to get all the recursers. it's weird.
	// this query returns an iterator, and then we have to use firestore
	// magic to iterate across the results of the query and store them
	// into our 'recursersList' variable which is a slice of map[string]interface{}
	iter := client.Collection("recursers").Where("isSkippingTomorrow", "==", false).Where("schedule."+today, "==", true).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		recursersList = append(recursersList, doc.Data())
	}

	// get everyone who was set to skip today and set them back to isSkippingTomorrow = false
	iter = client.Collection("recursers").Where("isSkippingTomorrow", "==", true).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		skippersList = append(skippersList, doc.Data())
	}
	for i := range skippersList {
		skippersList[i]["isSkippingTomorrow"] = false
		_, err = client.Collection("recursers").Doc(skippersList[i]["id"].(string)).Set(ctx, skippersList[i], firestore.MergeAll)
		if err != nil {
			log.Println(err)
		}
	}

	// shuffle our recursers. This will not error if the list is empty
	recursersList = shuffle(recursersList)

	// if for some reason there's no matches today, we're done
	if len(recursersList) == 0 {
		log.Println("No one was signed up to pair today -- so there were no matches")
		return
	}

	// message the peeps!
	doc, err := client.Collection("apiauth").Doc("key").Get(ctx)
	if err != nil {
		log.Panic(err)
	}
	apikey := doc.Data()
	botPassword := apikey["value"].(string)
	zulipClient := &http.Client{}

	// if there's an odd number today, message the last person in the list
	// and tell them they don't get a match today, then knock them off the list
	if len(recursersList)%2 != 0 {
		recurser := recursersList[len(recursersList)-1]
		recursersList = recursersList[:len(recursersList)-1]
		log.Println("Someone was the odd-one-out today")
		messageRequest := url.Values{}
		messageRequest.Add("type", "private")
		messageRequest.Add("to", recurser["email"].(string))
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

	for i := 0; i < len(recursersList); i += 2 {
		messageRequest := url.Values{}
		messageRequest.Add("type", "private")
		messageRequest.Add("to", recursersList[i]["email"].(string)+", "+recursersList[i+1]["email"].(string))
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
}
