package backend

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/outcrawl/backend/db"
	"github.com/outcrawl/backend/util"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type sendRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func subscribeHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if err := subscribe(ctx, user.Email); err == nil {
		util.ResponseJSON(w, user)
	} else {
		log.Infof(ctx, "%v", err)
		util.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
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
