package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v33/github"
	"github.com/gorilla/mux"
)

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }

var (
	appID    = int64(99645) // your app id goes here
	certPath = "/home/hl/Downloads/hydra-github-bot.2021-02-08.private-key.pem"
)

// Contains all the states descriptions
var descriptions = map[string]string{
	"success": "All the jobs were successful.",
	"pending": "Currently building the jobs...",
	"failure": "At least one of the jobs failed.",
	"error":   "Error",
}

// WriteStatus creates a new status via the provided client on a specific commit of a client's repository
// The description is automatically written and the target URL is hard-coded.
func WriteStatus(client *github.Client, owner, repo, rev, state, jobName string) {
	client.Repositories.CreateStatus(
		context.TODO(),
		owner,
		repo,
		rev,
		&github.RepoStatus{
			State:       String(state),
			Description: String(descriptions[state]),
			TargetURL:   String("https://hydra.visium.ch"),
			Context:     String(jobName),
		},
	)
}

// GetClientFromInstallationID returns the GitHub Client object from the Installation ID
func GetClientFromInstallationID(id int64) *github.Client {
	appsTransport, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, appID, certPath)
	if err != nil {
		panic("Error creating GitHub App client")
	}

	transport := ghinstallation.NewFromAppsTransport(
		appsTransport,
		id,
	)

	return github.NewClient(&http.Client{
		Transport: transport,
	})
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

		// Get Client from the WebHook installation
		client := GetClientFromInstallationID(event.Installation.GetID())
		// Write the status to the Repository retrieved from the Webhook
		WriteStatus(
			client,
			*event.Repo.Owner.Login,
			*event.Repo.Name,
			*event.CheckSuite.HeadSHA,
			"pending",
			"test-job",
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
