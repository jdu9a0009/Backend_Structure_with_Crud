package commands

import (
	"encoding/json"
	"log"
)

func MapToJson(m map[string]string) (string, error) {
	by, err := json.Marshal(m)
	if err != nil {
		log.Println("error at making json from map:", err)
		return "", err
	}

	return string(by), err
}

func JsonToMap(by string) (map[string]string, error) {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(by), &m); err != nil {
		log.Println("error at making map from json:", err)
		return nil, err
	}

	return m, nil
}
