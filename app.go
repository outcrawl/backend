package newsletter

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var (
	apiKey             string
	publicAPIKey       string
	domain             string
	mailingListAddress string

	publicToken  string
	privateToken string
)

func init() {
	apiKey = os.Getenv("API_KEY")
	publicAPIKey = os.Getenv("PUBLIC_API_KEY")
	domain = os.Getenv("DOMAIN")
	mailingListAddress = os.Getenv("MAILING_LIST_ADDRESS")

	publicToken = os.Getenv("PUBLIC_TOKEN")
	privateToken = os.Getenv("PRIVATE_TOKEN")

	r := mux.NewRouter()

	r.HandleFunc("/subscribe", subscribeHandler).
		Methods("POST").
		Queries("email", "{email}", "token", publicToken)
	r.HandleFunc("/send", sendMailHandler).
		Methods("POST").
		Queries("token", privateToken)

	http.Handle("/", cors.AllowAll().Handler(r))
}
