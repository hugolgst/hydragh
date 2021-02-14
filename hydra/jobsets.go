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

// Jobset describes the request sent to Hydra's API
type Jobset struct {
	Enabled            int                    `json:"enabled"`
	NixExpressionInput string                 `json:"nixexprinput"`
	NixExpressionPath  string                 `json:"nixexprpath"`
	JobsetInputs       map[string]JobsetInput `json:"jobsetinputs"`
}

// JobsetInput is a list of "Git checkout" Hydra inputs
type JobsetInput struct {
	JobsetInputAlts []string `json:"jobsetinputalts"`
}

// CreateJobset requests Hydra's API to create a jobset
func CreateJobset(project, jobset, cookie string, data Jobset) error {
	url := fmt.Sprintf("%s/jobset/%s/%s", apiURL, project, jobset)
	stringData, _ := json.Marshal(data)

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(stringData))
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", cookie)

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
