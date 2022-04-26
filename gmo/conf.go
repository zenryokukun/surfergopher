package gmo

import (
	"encoding/json"
	"log"
	"os"
)

type authKeys struct {
	Api    string `json:"api_key"`
	Secret string `json:"api_secret"`
}

func (k *authKeys) Keys() (string, string) {
	return k.Api, k.Secret
}

func NewAuthKeys(path string) *authKeys {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	keys := &authKeys{}
	err = json.Unmarshal(data, keys)
	if err != nil {
		log.Fatal(err)
	}
	return keys
}
