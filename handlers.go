package newsletter

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

type SendRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

const welcomeEmail = `<html><div style="font-family:-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol';text-align:center;background:#F5F5F5;padding:2rem;"> <h1 class="margin-top:0">Welcome to Outcrawl!</h1> <p> Awesome stuff coming right up! </p><p> If you received this email by mistake, simply unsubscribe. </p><a href="%unsubscribe_url%">Unsubscribe</a></div></html>`

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	v := r.URL.Query()
	email := v.Get("email")

	log.Infof(ctx, r.RemoteAddr)
	if err := checkSubscribeLimit(ctx, r.RemoteAddr); err != nil {
		responseError(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := subscribe(ctx, email); err == nil {
		updateSubscribeLimit(ctx, r.RemoteAddr)
		//sendTo(ctx, "Welcome!", welcomeEmail, email)
		responseJSON(w, "ok")
	} else {
		responseError(w, err.Error(), http.StatusBadRequest)
	}
}

func checkSubscribeLimit(ctx context.Context, addr string) error {
	if item, err := memcache.Get(ctx, addr); err == nil {
		if n, err := strconv.Atoi(string(item.Value)); err == nil && n >= 2 {
			return errors.New("limit reached")
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

func sendMailHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sr SendRequest
	err = json.Unmarshal(data, &sr)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	if err := send(ctx, sr.Subject, sr.Message); err == nil {
		responseJSON(w, "ok")
	} else {
		responseError(w, err.Error(), http.StatusInternalServerError)
	}
}
