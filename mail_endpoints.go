package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

type subscribeQuery struct {
	Email     string `form:"email"`
	Recaptcha string `form:"recaptcha"`
}

func subscribeEndpoint(c *gin.Context) {
	var query subscribeQuery
	if c.ShouldBindQuery(&query) == nil {
		if len(query.Email) == 0 {
			errorResponse(c, ErrInvalidEmail)
			return
		}
		if err := checkCaptcha(c, query.Recaptcha); err != nil {
			errorResponse(c, ErrIncorrectRecaptcha)
			return
		}
		if err := subscribe(c, query.Email); err == nil {
			c.String(http.StatusOK, "")
		} else {
			errorResponse(c, ErrNotSubscribed)
		}
	} else {
		errorResponse(c, ErrInvalidParameters)
	}
}

func checkCaptcha(c *gin.Context, response string) error {
	// Call recaptcha service
	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s", reCaptchaKey, response)
	client := urlfetch.Client(appengine.NewContext(c.Request))
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// Read response body
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var recaptchaData map[string]interface{}
	err = json.Unmarshal(data, &recaptchaData)
	if err != nil {
		return err
	}
	if recaptchaData["success"].(bool) {
		return nil
	}
	return errors.New("Invalid recaptcha")
}
