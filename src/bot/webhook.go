package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/fatih/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const botEmailAddress = "algo-bot@recurse.zulipchat.com"
const zulipAPIURL = "https://recurse.zulipchat.com/api/v1/messages"
const gcloudServerURL = "https://algobot-308118.ue.r.appspot.com"

var botMessages = InitMessenger("src/bot/messages.json")
var client *firestore.Client

// This is a struct that gets only what
// we need from the incoming JSON payload
type incomingJSON struct {
	Data    string `json:"data"`
	Token   string `json:"token"`
	Trigger string `json:"trigger"`
	Message struct {
		SenderID         int         `json:"sender_id"`
		DisplayRecipient interface{} `json:"display_recipient"`
		SenderEmail      string      `json:"sender_email"`
		SenderFullName   string      `json:"sender_full_name"`
	} `json:"message"`
}

// Zulip has to get JSON back from the bot,
// this does that. An empty message field stops
// zulip from throwing an error at the user that
// messaged the bot, but doesn't send a response
type botResponse struct {
	Message string `json:"content"`
}

type botNoResponse struct {
	Message bool `json:"response_not_required"`
}

// TODO: Docstring here!
func Webhook(w http.ResponseWriter, r *http.Request) {
	responder := json.NewEncoder(w)
	ctx := context.Background()

	var err error
	client, err = firestore.NewClient(ctx, "algobot-308118")
	defer client.Close()

	if err != nil {
		log.Panic(err)
	}

	// sanity check the incoming request
	userReq, err := sanityCheck(ctx, w, r)
	if err != nil {
		log.Println(err)
		return
	}

	if userReq.Trigger != "private_message" {
		err = responder.Encode(botResponse{
			`Hi! I'm AlgoBot!\n\nSend me a PM to get started :octopus::octopus::octopus:\n
            Why don't you also check out [the daily question](https://recurse.zulipchat.com/#narrow/stream/256561-Daily-LeetCode/topic/AlgoBot.20Daily.20Question)?`})
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
	response, err := dispatch(ctx, cmd, cmdArgs, strconv.Itoa(userReq.Message.SenderID), userReq.Message.SenderEmail, userReq.Message.SenderFullName)
	if err != nil {
		log.Println(err)
	}

	err = responder.Encode(botResponse{response})
	if err != nil {
		log.Println(err)
	}
}

// sanityCheck simply validates Zulip JSON originating from incoming webhooks
func sanityCheck(ctx context.Context, w http.ResponseWriter, r *http.Request) (incomingJSON, error) {
	var userReq incomingJSON

	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.NotFound(w, r)
		return userReq, err
	}

	// validate our zulip-bot token (manually put into the database before deployment)
	doc, err := client.Collection("auth").Doc("bot").Get(ctx)
	if err != nil {
		log.Println("Something weird happened trying to read the auth token from the database")
		return userReq, err
	}

	token := doc.Data()["token"]
	if userReq.Token != token {
		http.NotFound(w, r)
		return userReq, errors.New("unauthorized interaction attempt")
	}

	return userReq, err
}

func parseCmd(cmdStr string) (string, []string, error) {
	var err error
	var cmdList = []string{
		"cancel",
		"config",
		"help",
		"schedule",
		"skip",
		"subscribe",
		"unskip",
		"unsubscribe",
	}

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
		return cmd[0], nil, err

	// if there's not a valid command
	default:
		err = errors.New("the user-issued command wasn't valid")
		return "help", nil, err
	}
}

// TODO: Docstring here!
func dispatch(ctx context.Context, cmd string, cmdArgs []string, userID string, userEmail string, userName string) (string, error) {
	var response string
	var recurser Recurser

	// get the users "document" (database entry) out of firestore and temporarily keep it in 'doc'
	doc, err := client.Collection("recursers").Doc(userID).Get(ctx)

	// this says "if there's an error, and if that error was not document-not-found"
	if err != nil && grpc.Code(err) != codes.NotFound {
		response = botMessages.ReadError
		return response, err
	}

	// if there's a db entry, that means they were already subscribed to pairing bot
	isSubscribed := doc.Exists()

	// if the user is in the database, get their current state in a struct
	if isSubscribed {
		if err = doc.DataTo(&recurser); err != nil {
			log.Fatal(err)
		}
	}

	// here's the actual actions. command input from the user has already been sanitized,
	// so we can trust that cmd and cmdArgs only have valid stuff in them
	switch cmd {

	case "config":
		response = config(userID, recurser, isSubscribed)
		break

	case "schedule":
		response = schedule(userID, recurser, isSubscribed, ctx)
		break

	case "cancel":
		response = cancel(userID, recurser, isSubscribed, ctx)
		break

	case "subscribe":
		response = subscribe(userID, userName, userEmail, recurser, isSubscribed, ctx)
		break

	case "unsubscribe":
		response = unsubscribe(userID, recurser, isSubscribed, ctx)
		break

	case "skip":
		response = skip(userID, recurser, isSubscribed, ctx)
		break

	case "unskip":
		response = unskip(userID, recurser, isSubscribed, ctx)
		break

	case "help":
		response = botMessages.Help
		break

	default:
		// this won't execute because all input has been sanitized
		// by parseCmd() and all cases are handled explicitly above
	}
	return response, err
}

func config(userID string, recurser Recurser, isSubscribed bool) string {
	if isSubscribed == false {
		return botMessages.NotSubscribed
	}
	// Provide current settings as well as user-specific URL for config
	var response string
	response = recurser.stringifyUserConfig()
	response += fmt.Sprintf("[Click here to make changes](%s/config/%s)", gcloudServerURL, userID)

	return response
}

func schedule(userID string, recurser Recurser, isSubscribed bool, ctx context.Context) string {
	if isSubscribed == false {
		return botMessages.NotSubscribed
	}
	if !recurser.isConfigured() {
		return botMessages.NotConfigured
	}

	recurser.IsPairingTomorrow = true
	_, err := client.Collection("recursers").Doc(userID).Set(ctx, structs.Map(recurser), firestore.MergeAll)
	if err != nil {
		return botMessages.WriteError
	}

	response := "You're all set for a mock interview session! You'll be contacted shortly with all the pertinent details!\n\n"
	response += getQueueStatus(recurser, ctx)

	return response
}

func getQueueStatus(recurser Recurser, ctx context.Context) string {
	iter := client.Collection("recursers").Where("isPairingTomorrow", "==", true).Documents(ctx)
	recursersList := iterToRecurserList(iter)
	possibleMatches := 0

	for _, r := range recursersList {
		if isValidMatch(r, recurser) {
			possibleMatches += 1
		}
	}

	return fmt.Sprintf("Of the %v other Recursers in the queue, %v are a valid match!", len(recursersList)-1, possibleMatches-1)
}

func cancel(userID string, recurser Recurser, isSubscribed bool, ctx context.Context) string {
	if isSubscribed == false {
		return botMessages.NotSubscribed
	}
	if !recurser.IsPairingTomorrow {
		return "You are not signed up to pair!"
	}
	recurser.IsPairingTomorrow = false
	_, err := client.Collection("recursers").Doc(userID).Set(ctx, structs.Map(recurser), firestore.MergeAll)
	if err != nil {
		return botMessages.WriteError
	}

	return `Tomorrow: cancelled. I feel you. **I will not match you** for pairing tomorrow <3`
}

func subscribe(userID string, userName string, userEmail string, recurser Recurser, isSubscribed bool, ctx context.Context) string {
	if isSubscribed {
		return "You're already subscribed!"
	}

	recurser = newRecurser(userID, userName, userEmail)
	_, err := client.Collection("recursers").Doc(userID).Set(ctx, structs.Map(recurser), firestore.MergeAll)
	if err != nil {
		return botMessages.WriteError
	}

	sessions := map[string]interface{}{
		"sessions": []interface{}{},
	}

	_, err = client.Collection("soloSessions").Doc(userID).Create(ctx, sessions)
	if err != nil {
		return botMessages.WriteError
	}

	_, err = client.Collection("pairingSessions").Doc(userID).Create(ctx, sessions)
	if err != nil {
		return botMessages.WriteError
	}

	return botMessages.Subscribe
}

func unsubscribe(userID string, recurser Recurser, isSubscribed bool, ctx context.Context) string {
	if isSubscribed == false {
		return botMessages.NotSubscribed
	}

	_, err := client.Collection("recursers").Doc(userID).Delete(ctx)
	if err != nil {
		return botMessages.WriteError
	}
	_, err = client.Collection("soloSessions").Doc(userID).Delete(ctx)
	if err != nil {
		return botMessages.WriteError
	}
	_, err = client.Collection("pairingSessions").Doc(userID).Delete(ctx)
	if err != nil {
		return botMessages.WriteError
	}

	return botMessages.Unsubscribe
}

func skip(userID string, recurser Recurser, isSubscribed bool, ctx context.Context) string {
	if isSubscribed == false {
		return botMessages.NotSubscribed
	}
	recurser.IsSkippingTomorrow = true
	_, err := client.Collection("recursers").Doc(userID).Set(ctx, structs.Map(recurser), firestore.MergeAll)
	if err != nil {
		return botMessages.WriteError
	}

	return `Tomorrow: skipped. I feel you. **I will not contact you** with a question tomorrow <3`
}

func unskip(userID string, recurser Recurser, isSubscribed bool, ctx context.Context) string {
	if isSubscribed == false {
		return botMessages.NotSubscribed
	}

	recurser.IsSkippingTomorrow = false
	_, err := client.Collection("recursers").Doc(userID).Set(ctx, structs.Map(recurser), firestore.MergeAll)
	if err != nil {
		return botMessages.WriteError
	}

	return "Tomorrow: unskipped! Heckin *yes*! **I will contact you** with a question tomorrow :)"
}
