package bot

import (
	"math/rand"
	"time"
)

func contains(list []string, cmd string) bool {
	for _, v := range list {
		if v == cmd {
			return true
		}
	}
	return false
}

// this shuffles our recursers.
// TODO: source of randomness is time, but this runs at the
// same time each day. Is that ok?
func shuffle(slice []map[string]interface{}) []map[string]interface{} {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]map[string]interface{}, len(slice))
	perm := r.Perm(len(slice))
	for i, randIndex := range perm {
		ret[i] = slice[randIndex]
	}
	return ret
}
