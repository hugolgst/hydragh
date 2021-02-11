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
	orgID    = "your organization id"
	certPath = "/home/hl/Downloads/hydra-github-bot.2021-02-08.private-key.pem"

	installationID int64
	itr            *ghinstallation.Transport
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

// func processPR(ID string) {
// 	atr, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, appID, certPath)
// 	if err != nil {
// 		log.Fatal("error creating GitHub app client")
// 	}

// 	installation, _, err := github.NewClient(&http.Client{Transport: atr}).Apps.FindUserInstallation(context.TODO(), "hugolgst")
// 	if err != nil {
// 		log.Fatalf("error finding organization installation: %v", err)
// 	}

// 	installationID = installation.GetID()
// 	itr = ghinstallation.NewFromAppsTransport(atr, installationID)

// 	client := github.NewClient(&http.Client{Transport: itr})

// 	WriteStatus(
// 		client,
// 		"hugolgst",
// 		"github-hydra-bot",
// 		"6193b434b0bb87427a267785e8cfbc6e3f90759e",
// 		"failure",
// 		"vinixos/overlay",
// 	)
// }

func eventHandler(w http.ResponseWriter, r *http.Request) {
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
	case *github.PullRequestEvent:
		if action := *event.Action; action != "opened" {
			return
		}

		fmt.Println(event.Installation)
	default:
		fmt.Println(event)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/event_handler", eventHandler).Methods("POST")

	err := http.ListenAndServe(":10000", router)
	if err != nil {
		panic(err)
	}
}
