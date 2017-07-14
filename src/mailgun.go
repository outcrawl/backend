package newsletter

import (
	"bytes"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

func subscribe(ctx context.Context, email string) error {
	client := urlfetch.Client(ctx)
	endpoint := "https://api.mailgun.net/v3/lists/" + mailingListAddress + "/members"
	data := url.Values{
		"subscribed": {"True"},
		"address":    {email},
	}

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	req.SetBasicAuth("api", apiKey)

	_, err := client.Do(req)
	return err
}
