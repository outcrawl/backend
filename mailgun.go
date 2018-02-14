package backend

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/appengine"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine/urlfetch"
)

func subscribe(c *gin.Context, email string) error {
	client := urlfetch.Client(appengine.NewContext(c.Request))
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

func sendEmail(c *gin.Context, subject string, message string) error {
	return sendEmailTo(c, subject, message, mgMailingListAddress)
}

func sendEmailTo(c *gin.Context, subject string, message string, to string) error {
	client := urlfetch.Client(c)
	endpoint := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", mgDomain)

	data := url.Values{
		"from":    {"Outcrawl <contact@outcrawl.com>"},
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
