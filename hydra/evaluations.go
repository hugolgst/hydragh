package hydra

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// EvaluationsResponse contains the array of evaluations
type EvaluationsResponse struct {
	Evaluations []Evaluation `json:"evals"`
}

// Evaluation contains the array of build IDs of a specific evaluation
type Evaluation struct {
	Builds []int `json:"builds"`
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
