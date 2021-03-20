package bot

import (
	"context"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func contains(list []string, cmd string) bool {
	for _, v := range list {
		if v == cmd {
			return true
		}
	}
	return false
}

func min(arr []string, strVals map[string]int) int {
	min := strVals[arr[0]]
	for i := 1; i < len(arr); i++ {
		val := strVals[arr[i]]
		if min > val {
			min = val
		}
	}
	return min
}

// this shuffles our recursers.
// TODO: source of randomness is time, but this runs at the
// same time each day. Is that ok?
func shuffle(slice []Recurser) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]Recurser, len(slice))
	perm := r.Perm(len(slice))
	for i, randIndex := range perm {
		ret[i] = slice[randIndex]
	}
	slice = ret
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

func isValidMatch(recurserOne Recurser, recurserTwo Recurser) bool {
	difficulties := map[string]int{
		"easy":   0,
		"medium": 1,
		"hard":   2,
	}

	return min(recurserOne.Config.PairingDifficulty, difficulties) <= difficulties[recurserTwo.Config.Experience] &&
		min(recurserTwo.Config.PairingDifficulty, difficulties) <= difficulties[recurserOne.Config.Experience]
}
