package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
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

func processPR(ID string) {
	atr, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, appID, certPath)
	if err != nil {
		log.Fatal("error creating GitHub app client")
	}

	installation, _, err := github.NewClient(&http.Client{Transport: atr}).Apps.FindUserInstallation(context.TODO(), "hugolgst")
	if err != nil {
		log.Fatalf("error finding organization installation: %v", err)
	}

	installationID = installation.GetID()
	itr = ghinstallation.NewFromAppsTransport(atr, installationID)

	client := github.NewClient(&http.Client{Transport: itr})

	client.Repositories.CreateStatus(
		context.TODO(),
		"hugolgst",
		"github-hydra-bot",
		ID,
		&github.RepoStatus{
			State:       String("success"),
			Description: String("test"),
		},
	)
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: err=%s\n", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		processPR(*event.Commits[0].ID)
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
