package bot

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Messenger struct {
	Help          string `json:"help"`
	Subscribe     string `json:"subscribe"`
	Unsubscribe   string `json:"unsubscribe"`
	NotSubscribed string `json:"notSubscribed"`
	OddOneOut     string `json:"oddOneOut"`
	Matched       string `json:"matched"`
	WriteError    string `json:"writeError"`
	ReadError     string `json:"readError"`
}

func InitMessenger(filename string) Messenger {
	var messenger Messenger
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal([]byte(file), &messenger)
	if err != nil {
		log.Println(err)
	}
	return messenger
}
