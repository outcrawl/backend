package backend

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var (
	adminUserID          string
	adminToken           string
	adminEmail           string
	googleClientID       string
	googleCertsURL       string
	reCaptchaKey         string
	mgAPIKey             string
	mgPublicAPIKey       string
	mgDomain             string
	mgMailingListAddress string
)

func init() {
	adminUserID = os.Getenv("ADMIN_USER_ID")
	adminToken = os.Getenv("ADMIN_TOKEN")
	adminEmail = os.Getenv("ADMIN_EMAIL")
	googleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleCertsURL = os.Getenv("GOOGLE_CERTS_URL")
	reCaptchaKey = os.Getenv("RE_CAPTCHA_KEY")
	mgAPIKey = os.Getenv("MG_API_KEY")
	mgPublicAPIKey = os.Getenv("MG_PUBLIC_API_KEY")
	mgDomain = os.Getenv("MG_DOMAIN")
	mgMailingListAddress = os.Getenv("MG_MAILING_LIST_ADDRESS")

	r := mux.NewRouter()

	r.HandleFunc("/api/signin", Authenticate(signInHandler)).
		Methods("POST")

	// users
	r.HandleFunc("/api/users/{id}/ban", Authenticate(banUserHandler)).
		Methods("POST")
	r.HandleFunc("/api/users/{id}/unban", Authenticate(unbanUserHandler)).
		Methods("POST")

	// mail
	r.HandleFunc("/api/mail/subscribe/{email}/{recaptcha}", subscribeHandler).
		Methods("POST")
	r.HandleFunc("/api/mail/send", Authenticate(sendMailHandler)).
		Methods("POST")

	// comments
	r.HandleFunc("/api/threads/{id}", Authenticate(createThreadHandler)).
		Methods("POST")
	r.HandleFunc("/api/threads/{id}", readThreadHandler).
		Methods("GET")
	r.HandleFunc("/api/threads/{id}", Authenticate(deleteThreadHandler)).
		Methods("DELETE")
	r.HandleFunc("/api/threads/{id}/close", Authenticate(closeThreadHandler)).
		Methods("POST")
	r.HandleFunc("/api/threads/{id}/comments", Authenticate(createCommentHandler)).
		Methods("POST")
	r.HandleFunc("/api/threads/{threadId}/comments/{id}", Authenticate(deleteCommentHandler)).
		Methods("DELETE")

	http.Handle("/", cors.AllowAll().Handler(r))
}
