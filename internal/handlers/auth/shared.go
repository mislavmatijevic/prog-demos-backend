package auth

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security/captcha"
	log "github.com/sirupsen/logrus"
)

type authErrorCode int

const (
	NO_ERROR authErrorCode = iota
	ERR_INFO_INVALID
	ERR_USERNAME_TAKEN
	ERR_CAPTCHA_FAILED
	ERR_TOKEN_NOT_VALID
	ERR_TOKEN_NOT_FOUND
	ERR_TOKEN_EXPIRED
)

func (execErrCode authErrorCode) String() string {
	return [...]string{
		"given information is not valid for registration",
		"username or email already taken",
		"captcha rejected request",
		"token not valid",
		"token was not found",
		"token expired",
	}[execErrCode-1]
}

func (execErrCode authErrorCode) EnumIndex() int {
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
		respondForErrorCode(w, ERR_CAPTCHA_FAILED)
	default:
		log.WithError(err).WithFields(log.Fields{"priority": "high", "context": "captcha", "full_error": err.Error()})
		api.InternalErrorHandlerCustomMsg(w, "Unknown captcha error.")
	}
}

func respondForErrorCode(w http.ResponseWriter, errorCode authErrorCode) {
	var res = errorResponse{
		Success:   false,
		Message:   errorCode.String(),
		ErrorCode: errorCode.EnumIndex(),
	}

	api.RespondWithStatus(w, res, http.StatusBadRequest)
}
