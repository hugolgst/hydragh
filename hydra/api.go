package hydra

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/hugolgst/github-hydra-bot/github"
)

const apiURL = "http://localhost:3000"

type triggeredJobsetsResponse struct {
	JobsetsTriggered []string `json:"jobsetsTriggered"`
}

type buildStatusResponse struct {
	BuildStatus int `json:"buildstatus`
}

// EvaluationsResponse contains the array of evaluations
type EvaluationsResponse struct {
	Evaluations []Evaluation `json:"evals"`
}

// Evaluation contains the array of build IDs of a specific evaluation
type Evaluation struct {
	Builds []int `json:"builds"`
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

// processBuildStatus returns a GitHub status for a set of build IDs
func processBuildStatus(builds []int) github.Status {
	for build := range builds {
		value, err := getBuildStatus(build)
		if err != nil {
			return github.ErrorStatus
		}

		if value > 0 {
			return github.FailureStatus
		}
	}

	return github.SuccessStatus
}

// getBuildStatus returns the build status of a Hydra build:
// 0 success
// 1 failed
// 2 build error
func getBuildStatus(buildID int) (int, error) {
	url := fmt.Sprintf("%s/build/%d", apiURL, buildID)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 2, err
	}
	request.Header.Set("Accept", "application/json")

	// Send the HTTP request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 2, err
	}
	defer response.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 2, err
	}

	var jsonResponse buildStatusResponse
	json.Unmarshal(body, &jsonResponse)

	return jsonResponse.BuildStatus, nil
}

func WaitForStatus(statusChannel *chan github.Status, project, jobset string) {
	// Get the initial evaluations reply
	initialResponse, err := GetEvaluations(project, jobset)
	if err != nil {
		*statusChannel <- github.ErrorStatus
		return
	}
	fmt.Println("Got the initial evaluations response.")

	// Trigger the jobset
	err = TriggerJobset([]string{
		fmt.Sprintf("%s:%s", project, jobset),
	})
	if err != nil {
		*statusChannel <- github.ErrorStatus
		return
	}
	fmt.Println("Triggered the jobset.")

	// Initialize an infinite loop to request the API every 5 seconds until we get a positive/negative response
	for i := 0; i < 250; i++ {
		// Process the new evaluation
		response, err := GetEvaluations(project, jobset)
		if err != nil {
			*statusChannel <- github.ErrorStatus
			return
		}

		// If one evaluation more has been found
		if len(initialResponse.Evaluations) < len(response.Evaluations) {
			fmt.Println("New evaluation found.")
			latestEvaluation := response.Evaluations[len(response.Evaluations)-1]
			status := processBuildStatus(latestEvaluation.Builds)
			fmt.Println(status)
			*statusChannel <- status
			return
		}

		time.Sleep(5 * time.Second)
	}

	*statusChannel <- github.ErrorStatus
}

// GetEvaluations retrieves the latest evaluations of a specific jobset and returns them
func GetEvaluations(project, jobset string) (jsonResponse EvaluationsResponse, err error) {
	url := fmt.Sprintf("%s/jobset/%s/%s/evals", apiURL, project, jobset)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	request.Header.Set("Accept", "application/json")

	// Send the HTTP request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	json.Unmarshal(body, &jsonResponse)
	return jsonResponse, nil
}
