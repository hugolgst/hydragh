package hydra

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

// TriggeredJobsetsResponse contains the response when triggering a jobset
type TriggeredJobsetsResponse struct {
	JobsetsTriggered []string `json:"jobsetsTriggered"`
}

// CreateJobset requests Hydra's API to create a jobset
func CreateJobset(project, name, data string) error {
	url := fmt.Sprintf("%s/jobset/%s/%s", apiURL, project, name)

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")

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

	fmt.Println(string(body))

	return nil
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

	var jsonResponse TriggeredJobsetsResponse
	json.Unmarshal(body, &jsonResponse)

	if !reflect.DeepEqual(jsonResponse.JobsetsTriggered, slugs) {
		return errors.New("One or more of the jobsets were not triggered")
	}

	return nil
}
