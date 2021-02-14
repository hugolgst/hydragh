package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/go-github/v33/github"
	"github.com/gorilla/mux"
	"github.com/hugolgst/hydragh/configuration"
	botGithub "github.com/hugolgst/hydragh/github"
	"github.com/hugolgst/hydragh/hydra"
)

func writeFailureStatus(client botGithub.BotClient, event github.CheckSuiteEvent, name string) {
	client.WriteStatus(
		*event.Repo.Owner.Login,
		*event.Repo.Name,
		*event.CheckSuite.HeadSHA,
		name,
		botGithub.FailureStatus,
	)
}

func handleCheckSuite(event github.CheckSuiteEvent, job configuration.Job) {
	// Get Client from the WebHook installation
	client := botGithub.GetClientFromInstallationID(event.Installation.GetID())

	// Write the pending status
	jobset := fmt.Sprintf("%s-%s", job.Name, (*event.CheckSuite.HeadSHA)[0:6])
	client.WriteStatus(
		*event.Repo.Owner.Login,
		*event.Repo.Name,
		*event.CheckSuite.HeadSHA,
		jobset,
		botGithub.PendingStatus,
	)

	// Login into Hydra
	cookie, err := hydra.Login(os.Getenv("HYDRA_USERNAME"), os.Getenv("HYDRA_PASSWORD"))
	if err != nil {
		fmt.Println(err)
		writeFailureStatus(client, event, jobset)
		return
	}

	// Create the Jobset
	err = hydra.CreateJobset(configuration.Configuration.Project, jobset, cookie, hydra.Jobset{
		Enabled:            1,
		Visible:            1,
		NixExpressionInput: job.ExpressionInput,
		NixExpressionPath:  job.ExpressionPath,
		JobsetInputs:       hydra.GenerateJobsetInputs(job.Inputs, *event.CheckSuite.HeadBranch),
	})
	if err != nil {
		fmt.Println(err)
		writeFailureStatus(client, event, jobset)
		return
	}

	// Wait for the status to change
	status := make(chan botGithub.Status)
	go hydra.WaitForStatus(status, configuration.Configuration.Project, jobset)

	responseStatus := <-status
	client.WriteStatus(
		*event.Repo.Owner.Login,
		*event.Repo.Name,
		*event.CheckSuite.HeadSHA,
		jobset,
		responseStatus,
	)
	fmt.Printf("Status written on %s.", *event.CheckSuite.HeadSHA)
}

// EventHandler handles the income of WebHook requests
func EventHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Parse the WebHook payload using go-github
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		panic(err)
	}

	switch event := event.(type) {
	case *github.CheckSuiteEvent:
		if action := *event.Action; action != "requested" {
			return
		}

		// Handle the check suite for each registered job
		for _, job := range configuration.Configuration.Jobs {
			handleCheckSuite(*event, job)
		}
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/event_handler", EventHandler).Methods("POST")

	err := http.ListenAndServe(":10000", router)
	if err != nil {
		panic(err)
	}
}
