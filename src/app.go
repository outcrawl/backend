package newsletter

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	apiKey             string
	publicAPIKey       string
	domain             string
	mailingListAddress string
)

func init() {
	apiKey = os.Getenv("API_KEY")
	publicAPIKey = os.Getenv("PUBLIC_API_KEY")
	domain = os.Getenv("DOMAIN")
	mailingListAddress = os.Getenv("MAILING_LIST_ADDRESS")

	r := mux.NewRouter()

	r.HandleFunc("/subscribe", subscribeHandler).
		Methods("POST").
		Queries("email", "{email}")
	r.HandleFunc("/send", sendMailHandler).
		Methods("POST")

	http.Handle("/", r)
}
