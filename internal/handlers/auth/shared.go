package auth

import (
	"net/http"

	"github.com/mislavmatijevic/prog-demos-backend/internal/handlers/api"
	"github.com/mislavmatijevic/prog-demos-backend/internal/utils/security"
)

type authInputErrorCode int

const (
	EXEC_ERR_INFO_INVALID authInputErrorCode = iota + 1
	EXEC_ERR_USERNAME_TAKEN
	EXEC_ERR_RECAPTCHA_REQUIRES_CHALLENGE
)

func (execErrCode authInputErrorCode) String() string {
	return [...]string{
		"Given information is not valid for registration.",
		"Username or email already taken.",
		"Login did not score well at ReCaptcha, challenge user.",
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

func handleRecaptchaError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case security.ErrFailedToProcess.Error():
		api.InternalErrorHandlerCustomMsg(w, "Recaptcha is not available.")
	case security.ErrMustChallenge.Error():
		respondForErrorCode(w, EXEC_ERR_RECAPTCHA_REQUIRES_CHALLENGE)
	case security.ErrInvalid.Error():
		fallthrough
	default:
		api.RequestErrorHandlerCustomMsg(w, "Failed recaptcha.")
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
