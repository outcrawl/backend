package newsletter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type SubscribeRequest struct {
	Email     string `json:"email"`
	Recaptcha string `json:"recaptcha"`
}

var (
	mgAPIKey             = os.Getenv("MG_API_KEY")
	mgMailingListAddress = os.Getenv("MG_MAILING_LIST_ADDRESS")
	mgDomain             = os.Getenv("MG_DOMAIN")
	reCaptchaKey         = os.Getenv("RE_CAPTCHA_KEY")
)

func HandleSubscribe(req SubscribeRequest) error {
	if len(req.Email) == 0 {
		return errors.New("invalid email")
	}
	if err := checkCaptcha(req.Recaptcha); err != nil {
		log.Print(err)
		return errors.New("invalid captcha")
	}
	return subscribe(req.Email)
}

func checkCaptcha(response string) error {
	// Call recaptcha service
	endpoint := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s", reCaptchaKey, response)
	resp, err := http.Post(endpoint, "text/plain", nil)
	if err != nil {
		return err
	}

	// Read response body
	defer func() {
		_ = resp.Body.Close()
	}()
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
	return errors.New("invalid recaptcha")
}

func subscribe(email string) error {
	values := url.Values{
		"address": {email},
		"vars":    {fmt.Sprintf(`{"subscribed_at": "%d"}`, time.Now().UnixNano())},
	}
	endpoint := fmt.Sprintf(
		"https://api.mailgun.net/v3/lists/%s/members?%s",
		mgMailingListAddress,
		values.Encode(),
	)
	req, _ := http.NewRequest(http.MethodPost, endpoint, nil)
	req.SetBasicAuth("api", mgAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		if err := sendWelcomeMail(email); err != nil {
			log.Print(err)
		}
		return nil
	}
	if err != nil {
		log.Print(err)
	} else {
		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			log.Print(err)
		} else {
			log.Print(respBody["message"])
		}
	}

	return errors.New("member not added")
}

func sendWelcomeMail(email string) error {
	values := url.Values{
		"template": {"outcrawl-welcome"},
		"from":     {"Outcrawl <no-reply@outcrawl.com>"},
		"to":       {email},
		"subject":  {"Hello from Outcrawl"},
	}
	endpoint := fmt.Sprintf(
		"https://api.mailgun.net/v3/%s/messages?%s",
		mgDomain,
		values.Encode(),
	)
	req, _ := http.NewRequest(http.MethodPost, endpoint, nil)
	req.SetBasicAuth("api", mgAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}
	if err != nil {
		return err
	} else {
		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			log.Print(err)
			return err
		} else {
			return errors.New(respBody["message"].(string))
		}
	}
}
