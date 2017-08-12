package db

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func PutUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, "User", user.ID, 0, nil)
	if _, err := datastore.Put(ctx, key, user); err != nil {
		return err
	}
	return nil
}

func GetUser(ctx context.Context, user *User) error {
	key := datastore.NewKey(ctx, "User", user.ID, 0, nil)
	return datastore.Get(ctx, key, user)
}
