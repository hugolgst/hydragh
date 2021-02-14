package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/hugolgst/github-hydra-bot/configuration"
)

var (
	SuccessStatus = Status{
		Name:        "success",
		Description: "All the jobs were successful.",
	}
	PendingStatus = Status{
		Name:        "pending",
		Description: "Currently building the jobs...",
	}
	FailureStatus = Status{
		Name:        "failure",
		Description: "At least one of the jobs failed.",
	}
	ErrorStatus = Status{
		Name:        "error",
		Description: "An error occured.",
	}
)

// BotClient refers to the custom GitHub client for this bot
type BotClient struct {
	Client *github.Client
}

// Status is a GitHub API status containing a name and a description
type Status struct {
	Name        string
	Description string
}

// WriteStatus creates a new status via the provided client on a specific commit of a client's repository
// The description is automatically written and the target URL is hard-coded.
func (botClient BotClient) WriteStatus(owner, repo, rev, jobName string, status Status) {
	config := configuration.Configuration
	targetURL := fmt.Sprintf("%s/jobset/%s/%s", config.HydraURL, config.Project, jobName)

	botClient.Client.Repositories.CreateStatus(
		context.TODO(),
		owner,
		repo,
		rev,
		&github.RepoStatus{
			State:       String(status.Name),
			Description: String(status.Description),
			TargetURL:   String(targetURL),
			Context:     String(jobName),
		},
	)
}

// GetClientFromInstallationID returns the GitHub Client object from the Installation ID
func GetClientFromInstallationID(id int64) BotClient {
	appsTransport, err := ghinstallation.NewAppsTransportKeyFromFile(
		http.DefaultTransport,
		int64(configuration.Configuration.AppID),
		configuration.Configuration.CertificatePath,
	)
	if err != nil {
		panic("Error creating GitHub App client")
	}

	transport := ghinstallation.NewFromAppsTransport(
		appsTransport,
		id,
	)

	return BotClient{
		Client: github.NewClient(&http.Client{
			Transport: transport,
		}),
	}
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
