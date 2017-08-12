package db

import (
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func PutOrGetThread(ctx context.Context, thread *Thread) error {
	key := datastore.NewKey(ctx, "Thread", thread.ID, 0, nil)
	if err := datastore.Get(ctx, key, thread); err == nil {
		return nil
	}
	if _, err := datastore.Put(ctx, key, thread); err != nil {
		return err
	}
	return nil
}

func GetThreadWithComments(ctx context.Context, thread *Thread) error {
	key := datastore.NewKey(ctx, "Thread", thread.ID, 0, nil)
	if err := datastore.Get(ctx, key, thread); err != nil {
		return err
	}

	q := datastore.NewQuery("Comment").
		Filter("ThreadID =", thread.ID)
	thread.Comments = []Comment{}
	if keys, err := q.GetAll(ctx, &thread.Comments); err != nil {
		return err
	} else {
		for i := 0; i < len(keys); i++ {
			thread.Comments[i].ID = strconv.FormatInt(keys[i].IntID(), 10)
		}
		return nil
	}
}

func CloseThread(ctx context.Context, thread *Thread) error {
	key := datastore.NewKey(ctx, "Thread", thread.ID, 0, nil)
	thread.Closed = true
	if _, err := datastore.Put(ctx, key, thread); err != nil {
		return err
	}
	return nil
}

func DeleteThread(ctx context.Context, thread *Thread) error {
	key := datastore.NewKey(ctx, "Thread", thread.ID, 0, nil)
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	q := datastore.NewQuery("Comment").
		Filter("ThreadID =", thread.ID).
		KeysOnly()
	if keys, err := q.GetAll(ctx, nil); err != nil {
		return err
	} else {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			return err
		}
	}

	return nil
}
