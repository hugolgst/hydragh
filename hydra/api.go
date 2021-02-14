package hydra

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

const apiURL = "http://localhost:3000"

type triggeredJobsetsResponse struct {
	JobsetsTriggered []string `json:"jobsetsTriggered"`
}

// TriggerJobset triggers some specific jobsets defined by their slug: `project:jobset`
func TriggerJobset(slugs []string) error {
	url := fmt.Sprintf("%s/api/push?jobsets=%s", apiURL, strings.Join(slugs, ","))

	request, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	// Send the HTTP request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var jsonResponse triggeredJobsetsResponse
	json.Unmarshal(body, &jsonResponse)

	if !reflect.DeepEqual(jsonResponse.JobsetsTriggered, slugs) {
		return errors.New("One or more of the jobsets were not triggered")
	}

	return nil
}
