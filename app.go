package backend

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

var (
	adminToken           string
	reCaptchaKey         string
	mgAPIKey             string
	mgPublicAPIKey       string
	mgDomain             string
	mgMailingListAddress string
)

func init() {
	adminToken = os.Getenv("ADMIN_TOKEN")
	reCaptchaKey = os.Getenv("RE_CAPTCHA_KEY")
	mgAPIKey = os.Getenv("MG_API_KEY")
	mgPublicAPIKey = os.Getenv("MG_PUBLIC_API_KEY")
	mgDomain = os.Getenv("MG_DOMAIN")
	mgMailingListAddress = os.Getenv("MG_MAILING_LIST_ADDRESS")

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/mail/subscribe", subscribeEndpoint)
	}

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://outcrawl.com",
		},
		AllowCredentials: true,
	}).Handler(router)
	http.Handle("/", handler)
}
