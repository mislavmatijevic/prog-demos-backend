package misc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/authentication"
	"github.com/mislavmatijevic/prog-demos-backend/internal/database"
	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/integrations"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/utils_errors"
	log "github.com/sirupsen/logrus"
)

type reportIssueBody struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	IncludeUsername bool   `json:"includeUsername"`
	CaptchaToken    string `json:"captchaToken"`
}

type reportIssueResponse struct {
	Success     bool   `json:"success"`
	NewIssueUrl string `json:"newIssueUrl"`
}

func reportIssue(w http.ResponseWriter, r *http.Request) {
	var issueRequest reportIssueBody
	if err := json.NewDecoder(r.Body).Decode(&issueRequest); err != nil {
		api.RequestErrorHandlerGenericMsg(w, err)
		return
	}

	err := captcha.Verify("report-issue", issueRequest.CaptchaToken, r.RemoteAddr)
	if err != nil {
		utils_errors.HandleCaptchaError(w, err)
		return
	}

	if issueRequest.IncludeUsername {
		username := getUserFromAuth(r)
		if len(username) > 0 {
			issueRequest.Description = fmt.Sprintf("%s\n> reported by: `%s`", issueRequest.Description, username)
		}
	}

	url, err := integrations.CreateGithubIssue(issueRequest.Title, issueRequest.Description)
	if err != nil {
		api.InternalErrorHandlerCustomMsg(w, err.Error())
		return
	}

	api.RespondOk(w, reportIssueResponse{
		Success:     true,
		NewIssueUrl: url,
	})
}

func getUserFromAuth(r *http.Request) string {
	userId, err := authentication.GetUserIdFromRequest(r)
	if err != nil {
		log.WithFields(log.Fields{"priority": "low", "context": "report_issue"}).Warn("Failed to get user ID while reporting issue.")
		return ""
	}

	user, err := database.GetUserById(userId)
	if err != nil {
		log.WithFields(log.Fields{"priority": "medium", "context": "report_issue"}).Warnf("Failed to find user with user ID %d", userId)
		return ""
	}

	return user.Username
}
