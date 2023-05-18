package httpErrors

import (
	"net/http"
	"strings"

	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

var (
	BadRequest            = errors.New("Bad request")
	AlreadyExists         = errors.New("Already Existed")
	WrongCredentials      = errors.New("Wrong Credentials")
	NotFound              = errors.New("Not Found")
	NoContent             = errors.New("No Content")
	Unauthorized          = errors.New("Unauthorized")
	Forbidden             = errors.New("Forbidden")
	PermissionDenied      = errors.New("Permission Denied")
	ExpiredCSRFError      = errors.New("Expired CSRF token")
	WrongCSRFToken        = errors.New("Wrong CSRF token")
	CSRFNotPresented      = errors.New("CSRF not presented")
	NotRequiredFields     = errors.New("No such required fields")
	BadQueryParams        = errors.New("Invalid query params")
	InternalServerError   = errors.New("Internal Server Error")
	RequestTimeoutError   = errors.New("Request Timeout")
	ExistsEmailError      = errors.New("User with given email already exists")
	InvalidJWTToken       = errors.New("Invalid JWT token")
	InvalidJWTClaims      = errors.New("Invalid JWT claims")
	NotAllowedImageHeader = errors.New("Not allowed image header")
	NoCookie              = errors.New("not found cookie header")
	DuplicateResource     = errors.New("Duplicate Resource")
)

type IRestError interface {
	Status() int
	Code() string
	Error() string
	Text() string
}

type RestError struct {
	ErrStatus  int    `json:"status"`
	ErrCode    string `json:"code"`
	ErrMessage string `json:"message"`
	ErrText    string `json:"text"`
}

func (e RestError) Status() int {
	return e.ErrStatus
}

func (e RestError) Code() string {
	return e.ErrCode
}

func (e RestError) Error() string {
	return e.ErrMessage
}

func (e RestError) Text() string {
	return e.ErrText
}

func NewRestError(status int, err error, code ErrorCode, text string) IRestError {
	t := code.GetText()
	if text != "" {
		t = text
	}
	log.Info(t)
	return RestError{
		ErrStatus:  status,
		ErrCode:    string(code),
		ErrMessage: err.Error(),
		ErrText:    t,
	}
}

func NewBadRequestError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusBadRequest, err, ErrorCode(code), text)
}
func NewUnauthorizedError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusUnauthorized, err, ErrorCode(code), text)
}
func NewInternalServerError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusInternalServerError, err, ErrorCode(code), text)
}
func NewNotFoundError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusNotFound, err, ErrorCode(code), text)
}
func NewNoContentError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusNoContent, err, ErrorCode(code), text)
}
func NewConflictError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusConflict, err, ErrorCode(code), text)
}
func NewForbiddenError(err error, code string, text string) IRestError {
	return NewRestError(http.StatusForbidden, err, ErrorCode(code), text)
}

/*
func NewTestError(err error, code string, v ...interface{}) IRestError {
	errCode := ErrorCode(code)
	return RestError{
		ErrStatus:  http.StatusForbidden,
		ErrCode:    code,
		ErrMessage: err.Error(),
		ErrText:    errCode.GetText(),
	}
}
*/

func parseErrors(err error) IRestError {
	switch {
	case strings.Contains(err.Error(), "SQLSTATE"):
		return parseSqlError(err)
	default:
		if restErr, ok := err.(IRestError); ok {
			return restErr
		}
		return NewInternalServerError(err, "", "")
	}
}

func parseSqlError(err error) IRestError {
	return NewInternalServerError(errors.Wrap(err, "SQL ERROR"), "", "")
}

func ErrorResponse(err error) (IRestError, int) {
	restError := parseErrors(err)
	return restError, restError.Status()
}
