package backend

import (
	"errors"
	"net/http"

	"db"
	"util"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func signInHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	err := db.DatabaseTransaction(ctx, func(ctx context.Context) error {
		if err := db.GetUser(ctx, user); err == nil {
			return nil
		}
		if err := db.PutUser(ctx, user); err == nil {
			return nil
		}
		return errors.New("Could not sign in")
	})
	if err != nil {
		util.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	util.ResponseJSON(w, user)
}

func banUserHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}
	userID := mux.Vars(r)["id"]
	u := &db.User{ID: userID}

	err := db.DatabaseTransaction(ctx, func(ctx context.Context) error {
		if err := db.GetUser(ctx, u); err != nil {
			return errors.New("User not found")
		}
		u.Banned = true
		if err := db.PutUser(ctx, u); err != nil {
			return errors.New("Could not ban user")
		}
		return nil
	})
	if err != nil {
		util.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.ResponseJSON(w, u)
}

func unbanUserHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}
	userID := mux.Vars(r)["id"]
	u := &db.User{ID: userID}

	err := db.DatabaseTransaction(ctx, func(ctx context.Context) error {
		if err := db.GetUser(ctx, u); err != nil {
			return errors.New("User not found")
		}
		u.Banned = false
		if err := db.PutUser(ctx, u); err != nil {
			return errors.New("Could not unban user")
		}
		return nil
	})
	if err != nil {
		util.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.ResponseJSON(w, u)
}
