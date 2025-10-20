package auth

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
	log "github.com/sirupsen/logrus"
)

type authInputErrorCode int

const (
	EXEC_ERR_INFO_INVALID authInputErrorCode = iota + 1
	EXEC_ERR_USERNAME_TAKEN
	EXEC_ERR_CAPTCHA_FAILED
)

func (execErrCode authInputErrorCode) String() string {
	return [...]string{
		"Given information is not valid for registration.",
		"Username or email already taken.",
		"Captcha rejected request.",
	}[execErrCode-1]
}

func (execErrCode authInputErrorCode) EnumIndex() int {
	return int(execErrCode)
}

type errorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"errorCode"`
}

func handleCaptchaError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case captcha.ErrFailedToProcess.Error():
		api.InternalErrorHandlerCustomMsg(w, "Error while trying to process captcha token.")
	case captcha.ErrInvalid.Error():
		respondForErrorCode(w, EXEC_ERR_CAPTCHA_FAILED)
	default:
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "captcha", "full_error": err.Error()})
		api.InternalErrorHandlerCustomMsg(w, "Unknown captcha error.")
	}
}

func respondForErrorCode(w http.ResponseWriter, errorCode authInputErrorCode) {
	var res = errorResponse{
		Success:   false,
		Message:   errorCode.String(),
		ErrorCode: errorCode.EnumIndex(),
	}

	api.RespondWithStatus(w, res, http.StatusBadRequest)
}
