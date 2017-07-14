package newsletter

import (
	"net/http"

	"google.golang.org/appengine"
)

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	v := r.URL.Query()
	email := v.Get("email")

	if err := subscribe(ctx, email); err == nil {
		responseJSON(w, "ok")
	} else {
		responseError(w, err.Error(), http.StatusInternalServerError)
	}
}
