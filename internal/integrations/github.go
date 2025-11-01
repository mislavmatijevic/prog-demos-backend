package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

const GITHUB_ISSUES_APP_PRIVATE_KEY_FILE = "./prog-demos-bug-reporter.pem"

var client *github.Client = nil

func initializeGithubIntegration() {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 2196689, 92362358, GITHUB_ISSUES_APP_PRIVATE_KEY_FILE)
	if err != nil {
		log.Warnf("Failed to read the GitHub client private key file from %s.", GITHUB_ISSUES_APP_PRIVATE_KEY_FILE)
		return
	}

	client = github.NewClient(&http.Client{Transport: itr})
	log.Infof("Created GitHub client.")

	return
}

func CreateGithubIssue(title string, description string) (string, error) {
	if client == nil {
		log.Info("Skipping creating issue since GitHub client is not initialized.")
		return "", errors.New("GitHub client not initialized")
	}

	newIssueRequest := github.IssueRequest{
		Title:    github.String(title),
		Body:     github.String(description),
		Labels:   &[]string{"support"},
		Assignee: github.String("mislavmatijevic"),
		State:    github.String("open"),
	}

	issue, response, err := client.Issues.Create(context.Background(), "mislavmatijevic", "prog-demos-frontend", &newIssueRequest)
	if err != nil {
		issueBytes, err := json.Marshal(issue)
		responseBytes, err := json.Marshal(response)
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "github_integration", "issue": string(issueBytes), "response": string(responseBytes)}).Error("Failed to create issue")
		return "", err
	}

	return *issue.HTMLURL, nil
}
