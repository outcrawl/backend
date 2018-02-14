package backend

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	code     int
	httpCode int
	message  string
}

func (e Error) Error() string {
	return e.message
}

var (
	ErrInvalidParameters  = Error{1, http.StatusBadRequest, "Invalid parameters"}
	ErrInvalidEmail       = Error{2, http.StatusBadRequest, "Invalid email"}
	ErrIncorrectRecaptcha = Error{3, http.StatusBadRequest, "Incorrect recaptcha"}
	ErrNotSubscribed      = Error{4, http.StatusBadRequest, "Could not subscribe"}
)

func errorResponse(c *gin.Context, err Error) {
	c.JSON(err.httpCode, gin.H{
		"code":    err.code,
		"message": err.message,
	})
}
