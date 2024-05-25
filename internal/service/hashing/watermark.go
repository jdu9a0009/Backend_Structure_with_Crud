package hashing

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Data struct {
	Path   string `json:"path"`
	Author string `json:"author"`
}

func Watermark(data Data) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// POST watermark
	req, err := http.NewRequest("POST", "http://localhost:8090/stamp", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	return nil
}
