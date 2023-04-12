package httpErrors

import (
	"net/http"
	"strings"

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
	Causes() interface{}
}

type RestError struct {
	ErrStatus  int         `json:"status"`
	ErrCode    string      `json:"code"`
	ErrMessage string      `json:"message"`
	ErrCauses  interface{} `json:"-"`
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

func (e RestError) Causes() interface{} {
	return e.ErrCauses
}

func NewRestError(status int, code string, err error) IRestError {
	return RestError{
		ErrStatus:  status,
		ErrCode:    code,
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
}

func NewBadRequestError(err error) IRestError {
	return RestError{
		ErrStatus:  http.StatusBadRequest,
		ErrCode:    "",
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
}

func NewUnauthorizedError(err error) IRestError {
	return RestError{
		ErrStatus:  http.StatusUnauthorized,
		ErrCode:    "",
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
}

func NewInternalServerError(err error) IRestError {
	result := RestError{
		ErrStatus:  http.StatusInternalServerError,
		ErrCode:    "",
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
	return result
}

func NewNotFoundError(err error) IRestError {
	return RestError{
		ErrStatus:  http.StatusNotFound,
		ErrCode:    "",
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
}

func NewNoContentError(err error) IRestError {
	return RestError{
		ErrStatus:  http.StatusNoContent,
		ErrCode:    "",
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
}

func NewConflictError(err error) IRestError {
	return RestError{
		ErrStatus:  http.StatusConflict,
		ErrCode:    "",
		ErrMessage: err.Error(),
		ErrCauses:  err,
	}
}

/*
func NewForbiddenError(causes interface{}) RestErr {
	return ErrorJson{
		ErrStatus: http.StatusForbidden,
		ErrError:  Forbidden.Error(),
		ErrCauses: causes,
	}
}


func reflectVariableName(interface{}) string {
	return ""
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
		return NewInternalServerError(err)
	}
}

func parseSqlError(err error) IRestError {
	return NewInternalServerError(errors.Wrap(err, "SQL ERROR"))
}

func ErrorResponse(err error) (IRestError, int) {
	restError := parseErrors(err)
	return restError, restError.Status()
}
