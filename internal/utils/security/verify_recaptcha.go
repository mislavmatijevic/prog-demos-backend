package security

import (
	"errors"
	"os"
	"strings"

	"context"
	"fmt"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

var (
	recaptchaSiteKey = ""
	recaptchaApiKey  = ""
	projectID        = "prog-demos-website"
)

var (
	ErrFailedToProcess = errors.New("failed to process recaptcha token")
	ErrInvalid         = errors.New("recaptcha failed")
	ErrMustChallenge   = errors.New("recaptcha was suspicious and user should be forced to retry")
)

func Initialize() {
	recaptchaSiteKey = os.Getenv("RECAPTCHA_SITE_KEY")
	recaptchaApiKey = os.Getenv("RECAPTCHA_API_KEY")
}

func VerifyRecaptcha(action string, clientToken string, fullRemoteAddress string) error {
	ctx := context.Background()
	pureIp := strings.Split(fullRemoteAddress, ":")[0]

	client, err := recaptcha.NewClient(ctx, option.WithAPIKey(recaptchaApiKey))
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "high",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha"},
		).Error("Error creating ReCaptcha client!")
		return ErrFailedToProcess
	}
	defer client.Close()

	event := &recaptchapb.Event{
		Token:   clientToken,
		SiteKey: recaptchaSiteKey,
	}

	assessment := &recaptchapb.Assessment{
		Event: event,
	}

	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: assessment,
		Parent:     fmt.Sprintf("projects/%s", projectID),
	}

	response, err := client.CreateAssessment(ctx, request)

	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "high",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha"},
		).Error("Error calling ReCaptcha's CreateAssessment!")
		return ErrFailedToProcess
	}

	if !response.TokenProperties.Valid {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "high",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha",
				"reason":   fmt.Sprintf("%v", response.TokenProperties.InvalidReason)},
		).Error("Error calling ReCaptcha's CreateAssessment!")
		return ErrFailedToProcess
	}

	if response.TokenProperties.Action != action {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "low",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha",
				"reason":   fmt.Sprintf("%s != %s", response.TokenProperties.Action, action)},
		).Error("The action attribute in reCAPTCHA tag did not match the action expected to be scored.")
		return ErrInvalid
	}

	tokenHostname := response.TokenProperties.GetHostname()
	myHostname := thisHost()
	if tokenHostname != myHostname {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "low",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha",
				"reason":   fmt.Sprintf("%s != %s", tokenHostname, myHostname)},
		).Error("Hostname attribute in reCAPTCHA tag did not match the hostname expected to be scored.")
		return ErrInvalid
	}

	if response.RiskAnalysis.Score < 0.5 {
		log.WithFields(
			log.Fields{
				"priority": "medium",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha"},
		).Warning("Low Recaptcha score!")
		return ErrMustChallenge
	}

	return nil
}

func thisHost() string {
	if utils.IsProd() {
		return "progdemos.com"
	} else {
		return "localhost"
	}
}
