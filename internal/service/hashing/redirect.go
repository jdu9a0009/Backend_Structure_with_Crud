package hashing

import (
	"bytes"
	"io"
	"net/http"
)

func Redirect(data []byte) ([]byte, error) {

	// POST redirect
	req, err := http.NewRequest("POST", "http://172.16.100.247/redirect.php", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	var resByte []byte
	resByte, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//
	//if resp.StatusCode != 200 || res.Data == nil {
	//	return "", errors.New(fmt.Sprintf("status code: %d and message: %s", resp.StatusCode, res.Message))
	//}

	return resByte, nil
}
