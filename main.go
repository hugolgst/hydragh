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

// Status is a GitHub API status containing a name and a description
type Status struct {
	Name        string
	Description string
}

// All default status
var (
	successStatus = Status{
		Name:        "success",
		Description: "All the jobs were successful.",
	}
	pendingStatus = Status{
		Name:        "pending",
		Description: "Currently building the jobs...",
	}
	failureStatus = Status{
		Name:        "failure",
		Description: "At least one of the jobs failed.",
	}
	errorStatus = Status{
		Name:        "error",
		Description: "An error occured.",
	}
)

// WriteStatus creates a new status via the provided client on a specific commit of a client's repository
// The description is automatically written and the target URL is hard-coded.
func WriteStatus(client *github.Client, owner, repo, rev, jobName string, status Status) {
	client.Repositories.CreateStatus(
		context.TODO(),
		owner,
		repo,
		rev,
		&github.RepoStatus{
			State:       String(status.Name),
			Description: String(status.Description),
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
			"jobname",
			successStatus,
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
