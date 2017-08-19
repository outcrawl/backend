package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/outcrawl/backend/db"
	"github.com/outcrawl/backend/util"

	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
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
			if err := getKeySet(ctx); err != nil {
				util.ResponseError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			claims, err := validateToken(ctx, token)
			if err != nil {
				util.ResponseError(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if claims["aud"] != googleClientID {
				util.ResponseError(w, "Invalid client ID", http.StatusUnauthorized)
				return
			}
			if claims["iss"] != "https://accounts.google.com" && claims["iss"] != "accounts.google.com" {
				util.ResponseError(w, "Token not issued by Google", http.StatusUnauthorized)
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

func validateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid := token.Header["kid"].(string)
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keySet[kid]))
		if err != nil {
			memcache.Delete(ctx, "server:certs")
			return nil, err
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func getKeySet(ctx context.Context) error {
	if keySet == nil {
		keySet = make(map[string]string)
	} else {
		return nil
	}

	// check cache
	if item, err := memcache.Get(ctx, "server:certs"); err == nil {
		return json.Unmarshal(item.Value, &keySet)
	}

	// fetch keys
	client := urlfetch.Client(ctx)
	resp, err := client.Get(googleCertsURL)
	if err != nil {
		return err
	}

	// read keys
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &keySet)
	if err != nil {
		return err
	}

	// cache keys
	cacheControl := resp.Header.Get("Cache-Control")
	re := regexp.MustCompile(`max-age=[0-9]+`)
	maxAgeString := re.FindString(cacheControl)[8:]
	maxAge, err := strconv.ParseInt(maxAgeString, 10, 64)
	if err != nil {
		return errors.New("Could not parse Cache-Control header")
	}

	item := &memcache.Item{
		Key:        "server:certs",
		Value:      data,
		Expiration: time.Duration(maxAge-60) * time.Second,
	}
	memcache.Add(ctx, item)
	return nil
}

func parsePublicKey() {

}
