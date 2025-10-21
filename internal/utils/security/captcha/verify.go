package captcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mislavmatijevic/prog-demos-backend/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	ErrFailedToProcess = errors.New("failed to process captcha token")
	ErrInvalid         = errors.New("captcha failed")
)

var (
	captchaKeys          = make(map[string]string)
	expectedHostname     = ""
	minScoreForChallenge float32
)

type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
}

func Initialize() {
	if utils.IsProd() {
		expectedHostname = "www.progdemos.com"
	} else {
		expectedHostname = "example.com"
	}

	var captchaAllKeys = os.Getenv("CAPTCHA_KEYS")
	var keysValuePairs = strings.Split(captchaAllKeys, ",")

	for _, pair := range keysValuePairs {
		action := strings.Split(pair, ":")[0]
		key := strings.Split(pair, ":")[1]
		captchaKeys[action] = key
	}

	log.WithField("captcha_keys", len(captchaKeys)).Infof("Loaded %d Captcha keys.", len(captchaKeys))
}

func Verify(action string, clientToken string, fullRemoteAddress string) error {
	if strings.TrimSpace(clientToken) == "" {
		return ErrFailedToProcess
	}

	pureIp := strings.Split(fullRemoteAddress, ":")[0]

	secret := captchaKeys[action]
	if secret == "" {
		err := errors.New("Missing action")
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "low",
				"action":   action,
				"ip":       pureIp,
				"context":  "captcha",
				"reason":   fmt.Sprintf("Missing action: '%s'", action)},
		)
		return err
	}

	jsonRequestObject, _ := json.Marshal(map[string]string{"secret": secret, "response": clientToken, "remoteip": pureIp})
	turnstileResponse, err := http.Post("https://challenges.cloudflare.com/turnstile/v0/siteverify", "application/json", bytes.NewBuffer(jsonRequestObject))

	if err != nil || turnstileResponse.StatusCode >= 500 {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "high",
				"action":   action,
				"ip":       pureIp,
				"context":  "captcha"},
		).Error("Error contacting Turnstile's API!")
		return ErrFailedToProcess
	}

	defer turnstileResponse.Body.Close()

	response := TurnstileResponse{}

	bodyBytes, err := io.ReadAll(turnstileResponse.Body)
	if err == nil {
		err = json.Unmarshal(bodyBytes, &response)
	}
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"priority": "high",
				"action":   action,
				"ip":       pureIp,
				"context":  "captcha"},
		).Error("Error reading Turnstile's response body!")
		return ErrFailedToProcess
	}

	log.WithFields(log.Fields{"turnstile-response": string(bodyBytes)}).Debug("Turnstile responded")

	if !response.Success {
		return handleFailedresponse(response, err, action, pureIp, bodyBytes)
	}

	if response.Hostname != expectedHostname {
		log.WithFields(
			log.Fields{
				"priority":        "high",
				"action":          action,
				"response_action": response.Action,
				"ip":              pureIp,
				"context":         "captcha"},
		).Errorf("Turnstile's response contained unexpected hostname: '%s'", response.Hostname)
		return ErrInvalid
	}

	if utils.IsProd() && response.Action != action {
		log.WithFields(
			log.Fields{
				"priority":        "high",
				"action":          action,
				"response_action": response.Action,
				"expected_action": action,
				"ip":              pureIp,
				"context":         "captcha"},
		).Errorf("Turnstile's response wasn't sent for correct action: %s != %s", response.Action, action)
		return ErrInvalid
	}

	log.Debugf("Captcha passed.")
	return nil
}

func handleFailedresponse(response TurnstileResponse, err error, action string, pureIp string, bodyBytes []byte) error {
	for _, code := range response.ErrorCodes {
		switch code {
		case "invalid-input-response":
			fallthrough
		case "timeout-or-duplicate":
			{
				return ErrInvalid
			}
		case "missing-input-secret":
			fallthrough
		case "invalid-input-secret":
			fallthrough
		case "missing-input-response":
			fallthrough
		case "bad-request":
			fallthrough
		case "internal-error":
			fallthrough
		default:
			{
				log.WithError(err).WithFields(
					log.Fields{
						"priority":      "high",
						"action":        action,
						"ip":            pureIp,
						"response":      string(bodyBytes),
						"reason-failed": code,
						"context":       "captcha"},
				).Error("An error occured during interaction with Turnstile!")
			}
		}
	}

	return ErrFailedToProcess
}
