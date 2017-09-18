package backend

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

func subscribe(ctx context.Context, email string) error {
	client := urlfetch.Client(ctx)
	endpoint := fmt.Sprintf("https://api.mailgun.net/v3/lists/%s/members", mgMailingListAddress)
	data := url.Values{
		"address": {email},
		"vars":    {fmt.Sprintf(`{"subscribed_at": "%d"}`, time.Now().UnixNano())},
	}

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	req.SetBasicAuth("api", mgAPIKey)

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}

	if err != nil {
		return err
	}
	return errors.New("Member not added")
}

func send(ctx context.Context, subject string, message string) error {
	return sendTo(ctx, subject, message, mgMailingListAddress)
}

func sendTo(ctx context.Context, subject string, message string, to string) error {
	client := urlfetch.Client(ctx)
	endpoint := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", mgDomain)

	data := url.Values{
		"from":    {"Outcrawl <news@outcrawl.com>"},
		"to":      {to},
		"subject": {subject},
		"html":    {message},
	}

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	req.SetBasicAuth("api", mgAPIKey)

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New("Mail not sent")
}
