package hydra

import (
	"fmt"
	"time"

	"github.com/hugolgst/github-hydra-bot/github"
)

const apiURL = "http://localhost:3000"

func login() {

}

func WaitForStatus(statusChannel chan github.Status, project, jobset string) {
	// Get the initial evaluations reply
	initialResponse, err := GetEvaluations(project, jobset)
	if err != nil {
		statusChannel <- github.ErrorStatus
		return
	}
	fmt.Println("Got the initial evaluations response.")

	// Trigger the jobset
	err = TriggerJobset([]string{
		fmt.Sprintf("%s:%s", project, jobset),
	})
	if err != nil {
		statusChannel <- github.ErrorStatus
		return
	}
	fmt.Println("Triggered the jobset.")

	// Initialize an infinite loop to request the API every 5 seconds until we get a positive/negative response
	for i := 0; i < 250; i++ {
		// Process the new evaluation
		response, err := GetEvaluations(project, jobset)
		if err != nil {
			statusChannel <- github.ErrorStatus
			return
		}

		// If one evaluation more has been found
		if len(initialResponse.Evaluations) < len(response.Evaluations) {
			fmt.Println("New evaluation found.")
			latestEvaluation := response.Evaluations[len(response.Evaluations)-1]
			status := ProcessBuildStatus(latestEvaluation.Builds)

			statusChannel <- status
			return
		}

		time.Sleep(5 * time.Second)
	}

	statusChannel <- github.ErrorStatus
}
