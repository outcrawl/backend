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
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

type AuthenticatedHandlerFunc func(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request)

var keySet map[string][]byte

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
		var pem []byte

		// check local variable
		pem = keySet[kid]
		// check cache
		if len(pem) == 0 {
			pem = getCachedKey(ctx, kid)
		}
		// fetch new keys
		if len(pem) == 0 {
			pem = fetchNewKey(ctx, kid)
		}

		if len(pem) > 0 {
			key, err := jwt.ParseRSAPublicKeyFromPEM(pem)
			if err != nil {
				log.Infof(ctx, "%v", err)
				return nil, err
			}
			return key, nil
		}
		return nil, errors.New("PEM Key not found")
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func getCachedKey(ctx context.Context, kid string) []byte {
	if item, err := memcache.Get(ctx, "key:"+kid); err == nil {
		if keySet == nil {
			keySet = make(map[string][]byte)
		}
		keySet[kid] = item.Value
		return item.Value
	}
	return nil
}

func fetchNewKey(ctx context.Context, kid string) []byte {
	// fetch keys
	client := urlfetch.Client(ctx)
	resp, err := client.Get(googleCertsURL)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return nil
	}

	// read keys
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return nil
	}
	var keysJson map[string]string
	err = json.Unmarshal(data, &keysJson)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return nil
	}

	// read max age
	cacheControl := resp.Header.Get("Cache-Control")
	re := regexp.MustCompile(`max-age=[0-9]+`)
	maxAgeString := re.FindString(cacheControl)[8:]
	maxAge, err := strconv.ParseInt(maxAgeString, 10, 64)
	if err == nil {
		// cache keys
		for k, v := range keysJson {
			item := &memcache.Item{
				Key:        "key:" + k,
				Value:      []byte(v),
				Expiration: time.Duration(maxAge-60) * time.Second,
			}
			memcache.Add(ctx, item)
		}
	}

	if keySet == nil {
		keySet = make(map[string][]byte)
	}
	for k, v := range keysJson {
		keySet[k] = []byte(v)
	}

	return keySet[kid]
}

func cacheKey(ctx context.Context, kid string, exp time.Duration, key []byte) {
	item := &memcache.Item{
		Key:        "key:" + kid,
		Value:      key,
		Expiration: exp,
	}
	memcache.Add(ctx, item)
}
