package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v33/github"
	"github.com/gorilla/mux"
	botGithub "github.com/hugolgst/github-hydra-bot/github"
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

		// Get Client from the WebHook installation
		client := botGithub.GetClientFromInstallationID(event.Installation.GetID())

		jobset := fmt.Sprintf("checkers-%s", (*event.CheckSuite.HeadSHA)[0:6])
		client.WriteStatus(
			*event.Repo.Owner.Login,
			*event.Repo.Name,
			*event.CheckSuite.HeadSHA,
			jobset,
			botGithub.PendingStatus,
		)

		hydra.CreateJobset("test", jobset, `{
			"emailoverride" : "",
			"enabled" : 1,
			"errormsg" : "",
			"fetcherrormsg" : null,
			"jobsetinputs" : {
					"nixpkgs" : {
						"jobsetinputalts" : [
								"git@github.com:NixOS/nixpkgs nixos-20.09"
						]
					},
					"vinixos" : {
						"jobsetinputalts" : [
								"git@github.com:VisiumCH/vinixos.git main"
						]
					}
			},
			"nixexprinput" : "vinixos",
			"nixexprpath" : "hydra/checkers.nix"
		}`)

		status := make(chan botGithub.Status)
		hydra.WaitForStatus(&status, "test", jobset)

		responseStatus := <-status
		fmt.Println(responseStatus)

		client.WriteStatus(
			*event.Repo.Owner.Login,
			*event.Repo.Name,
			*event.CheckSuite.HeadSHA,
			jobset,
			responseStatus,
		)
		fmt.Printf("Status written on %s.", *event.CheckSuite.HeadSHA)
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
