package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v33/github"
	"github.com/gorilla/mux"
	"github.com/hugolgst/github-hydra-bot/hydra"
)

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

		err := hydra.TriggerJobset([]string{
			"vinixos:checkers",
		})
		if err != nil {
			fmt.Println("Hydra jobset cannot be triggered.")
			return
		}
		fmt.Println("Jobsets triggered.")

		// Get Client from the WebHook installation
		// client := botGithub.GetClientFromInstallationID(event.Installation.GetID())
		// // Write the status to the Repository retrieved from the Webhook
		// client.WriteStatus(
		// 	*event.Repo.Owner.Login,
		// 	*event.Repo.Name,
		// 	*event.CheckSuite.HeadSHA,
		// 	"jobname",
		// 	botGithub.SuccessStatus,
		// )
		// fmt.Printf("Status written on %s.", *event.CheckSuite.HeadSHA)
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
