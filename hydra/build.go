package hydra

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hugolgst/github-hydra-bot/github"
)

// BuildStatusResponse contains the status of the build
type BuildStatusResponse struct {
	BuildStatus int `json:"buildstatus"`
}

// ProcessBuildStatus returns a GitHub status for a set of build IDs
func ProcessBuildStatus(builds []int) github.Status {
	for build := range builds {
		value, err := GetBuildStatus(build)
		if err != nil {
			return github.ErrorStatus
		}

		if value > 0 {
			return github.FailureStatus
		}
	}

	return github.SuccessStatus
}

// GetBuildStatus returns the build status of a Hydra build:
// 0 success
// 1 failed
// 2 build error
func GetBuildStatus(buildID int) (int, error) {
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

	var jsonResponse BuildStatusResponse
	json.Unmarshal(body, &jsonResponse)

	return jsonResponse.BuildStatus, nil
}
