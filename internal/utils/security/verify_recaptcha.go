package security

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

var (
	recaptchaSiteKey     = ""
	recaptchaApiKey      = ""
	expectedHostname     = ""
	projectID            = "prog-demos-website"
	minScoreForChallenge float32
)

var (
	ErrFailedToProcess = errors.New("failed to process recaptcha token")
	ErrInvalid         = errors.New("recaptcha failed")
	ErrMustChallenge   = errors.New("recaptcha was suspicious and user should be forced to retry")
)

func Initialize() {
	recaptchaSiteKey = os.Getenv("RECAPTCHA_SITE_KEY")
	recaptchaApiKey = os.Getenv("RECAPTCHA_API_KEY")
	expectedHostname = os.Getenv("FRONT_HOSTNAME")

	challengeScore, err := strconv.ParseFloat(os.Getenv("RECAPTCHA_CHALLENGE_SCORE"), 32)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "recaptcha"}).Panic("Minimal recaptcha score not set!")
	}
	minScoreForChallenge = float32(challengeScore)
	log.WithField("min_score_for_challenge", minScoreForChallenge).Info("Minimal reCAPTCHA score set!")
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
	if tokenHostname != expectedHostname {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "low",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha",
				"reason":   fmt.Sprintf("%s != %s", tokenHostname, expectedHostname)},
		).Error("Hostname attribute in reCAPTCHA tag did not match the hostname expected to be scored.")
		return ErrInvalid
	}

	if response.RiskAnalysis.Score < 0.7 {
		log.WithFields(
			log.Fields{
				"priority": "medium",
				"action":   action,
				"ip":       pureIp,
				"context":  "recaptcha",
				"score":    response.RiskAnalysis.Score},
		).Warning("Low Recaptcha score!")

		if response.RiskAnalysis.Score <= minScoreForChallenge {
			return ErrMustChallenge
		}
	}

	log.Debugf("RECAPTCHA OK %f!", response.RiskAnalysis.Score)

	return nil
}
