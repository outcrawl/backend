package backend

import (
	"fmt"
	"net/http"

	"db"
	"util"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

type AuthenticatedHandlerFunc func(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request)

var keySet map[string]string

func Authenticate(handler AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		token := r.Header.Get("Authorization")

		if len(token) == 0 {
			util.ResponseError(w, "Authorization header not present", http.StatusUnauthorized)
			return
		}

		if token == adminToken {
			user := &db.User{
				ID:    adminUserID,
				Admin: true,
				Email: adminEmail,
			}
			handler(ctx, user, w, r)
		} else {
			if keySet == nil {
				getKeySet(ctx)
			}

			claims, err := validateToken(token)
			if err != nil {
				util.ResponseError(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if claims["aud"] != googleClientID {
				util.ResponseError(w, "Invalid client ID", http.StatusUnauthorized)
				return
			}

			userID := claims["sub"].(string)
			email := claims["email"].(string)
			user := &db.User{
				ID:    userID,
				Email: email,
				Admin: userID == adminUserID,
			}
			handler(ctx, user, w, r)
		}
	}
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid := token.Header["kid"].(string)
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keySet[kid]))
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func getKeySet(ctx context.Context) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(googleCertsURL)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return
	}
	if err := util.ReadJSON(resp.Body, &keySet); err != nil {
		log.Errorf(ctx, "%v", err)
	}
}
