package integrations

import (
	"context"
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
		log.Warn("Failed to read the file")
	} else {
		log.Infof("Loaded GitHub secret.")
	}

	client = github.NewClient(&http.Client{Transport: itr})

	return
}

func CreateGithubIssue(title string, description string) {
	newIssueRequest := github.IssueRequest{
		Title:    github.String(title),
		Body:     github.String(description),
		Labels:   &[]string{"support"},
		Assignee: github.String("mislavmatijevic"),
		State:    github.String("open"),
	}

	issue, response, err := client.Issues.Create(context.Background(), "mislavmatijevic", "prog-demos-frontend", &newIssueRequest)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "github_integration", "issue": issue, "response": response}).Panic("Failed to create issue")
	}
}
