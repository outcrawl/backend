package backend

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"db"
	"util"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

type sendRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func subscribeHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if err := checkSubscribeLimit(ctx, r.RemoteAddr); err != nil {
		util.ResponseError(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := subscribe(ctx, user.Email); err == nil {
		updateSubscribeLimit(ctx, r.RemoteAddr)
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

func checkSubscribeLimit(ctx context.Context, addr string) error {
	if item, err := memcache.Get(ctx, addr); err == nil {
		if n, err := strconv.Atoi(string(item.Value)); err == nil && n >= 2 {
			return errors.New("Limit reached")
		}
	}
	return nil
}

func updateSubscribeLimit(ctx context.Context, addr string) {
	n := 0
	if item, err := memcache.Get(ctx, addr); err == nil {
		n, _ = strconv.Atoi(string(item.Value))
	}
	n++
	memcache.Set(ctx, &memcache.Item{
		Key:        addr,
		Value:      []byte(strconv.Itoa(n)),
		Expiration: 24 * time.Hour,
	})
}
