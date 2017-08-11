package db

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func DatabaseTransaction(ctx context.Context, f func(context.Context) error) error {
	return datastore.RunInTransaction(ctx, f, nil)
}
