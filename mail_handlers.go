package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"
	"github.com/outcrawl/backend/db"
	"github.com/outcrawl/backend/util"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type sendRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	vars := mux.Vars(r)
	email := strings.TrimSpace(vars["email"])
	recaptcha := vars["recaptcha"]

	if err := checkCaptcha(ctx, recaptcha); err != nil {
		util.ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(email) == 0 {
		util.ResponseError(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if err := subscribe(ctx, email); err == nil {
		util.ResponseJSON(w, "")
	} else {
		log.Infof(ctx, "%v", err)
		util.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func checkCaptcha(ctx context.Context, response string) error {
	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s", reCaptchaKey, response)
	client := urlfetch.Client(ctx)
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New("Invalid captcha")
}

func sendMailHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sr sendRequest
	err = json.Unmarshal(data, &sr)
	if err != nil {
		util.ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := send(ctx, sr.Subject, sr.Message); err == nil {
		util.ResponseJSON(w, "")
	} else {
		util.ResponseError(w, err.Error(), http.StatusInternalServerError)
	}
}
