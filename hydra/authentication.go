package hydra

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// AuthenticationRequest contains the username and password for the Login request
type AuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login sends an authentication request to Hydra's API and returns the cookie and error
func Login(username, password string) (string, error) {
	jsonString, _ := json.Marshal(AuthenticationRequest{username, password})

	// Create the HTTP request
	request, err := http.NewRequest("POST", apiURL+"/login", bytes.NewBuffer(jsonString))
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Referer", apiURL)

	// Send the HTTP request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var jsonResponse map[string]interface{}
	json.Unmarshal(body, &jsonResponse)

	if jsonResponse["error"] != nil {
		return "", errors.New("wrong credentials")
	}

	return response.Header.Get("Set-Cookie"), nil
}
