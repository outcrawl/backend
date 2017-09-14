package db

import (
	"errors"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func PutComment(ctx context.Context, comment *Comment) error {
	var thread Thread
	threadKey := datastore.NewKey(ctx, "Thread", comment.ThreadID, 0, nil)
	if err := datastore.Get(ctx, threadKey, &thread); err != nil {
		return err
	}

	if thread.Closed {
		return errors.New("Thread is closed")
	}

	return DatabaseTransaction(ctx, func(ctx context.Context) error {
		if len(comment.ReplyTo) != 0 {
			replyID, err := strconv.ParseInt(comment.ReplyTo, 10, 64)
			if err != nil {
				parentKey := datastore.NewKey(ctx, "Comment", "", replyID, nil)
				var parent Comment
				if err := datastore.Get(ctx, parentKey, &parent); err != nil {
					return err
				}
			}
		}
		key := datastore.NewIncompleteKey(ctx, "Comment", nil)
		if newKey, err := datastore.Put(ctx, key, comment); err != nil {
			return err
		} else {
			comment.ID = strconv.FormatInt(newKey.IntID(), 10)
			return nil
		}
	})
}

func GetComment(ctx context.Context, comment *Comment) error {
	id, err := strconv.ParseInt(comment.ID, 10, 64)
	if err != nil {
		return err
	}
	key := datastore.NewKey(ctx, "Comment", "", id, nil)
	return datastore.Get(ctx, key, comment)
}

func DeleteComment(ctx context.Context, comment *Comment) error {
	id, err := strconv.ParseInt(comment.ID, 10, 64)
	if err != nil {
		return err
	}
	key := datastore.NewKey(ctx, "Comment", "", id, nil)
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}

	var replies []*Comment
	if keys, err := datastore.NewQuery("Comment").
		Filter("ReplyTo =", comment.ID).
		KeysOnly().
		GetAll(ctx, &replies); err != nil {
		return err
	} else {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			return err
		}
	}

	return nil
}
