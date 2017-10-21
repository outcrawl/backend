package backend

import (
	"html/template"
	"net/http"
	"time"

	"github.com/outcrawl/backend/db"
	"github.com/outcrawl/backend/util"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"strings"
)

func createThreadHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}

	thread := &db.Thread{
		ID: mux.Vars(r)["id"],
	}
	if err := db.PutOrGetThread(ctx, thread); err != nil {
		log.Errorf(ctx, "%v", err)
		util.ResponseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		util.ResponseJSON(w, thread)
	}
}

func readThreadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	thread := &db.Thread{
		ID:       mux.Vars(r)["id"],
		Comments: []db.Comment{},
	}

	if err := getCachedItem(ctx, "thread:"+thread.ID, &thread); err == nil {
		util.ResponseJSON(w, thread)
		return
	}

	if err := db.GetThreadWithComments(ctx, thread); err != nil {
		log.Errorf(ctx, "%v", err)
		util.ResponseError(w, "Could not read thread", http.StatusInternalServerError)
		return
	}

	cacheItem(ctx, "thread:"+thread.ID, thread)
	util.ResponseJSON(w, thread)
}

func deleteThreadHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}
	thread := &db.Thread{
		ID: mux.Vars(r)["id"],
	}
	if err := db.DeleteThread(ctx, thread); err != nil {
		log.Errorf(ctx, "%v", err)
		util.ResponseError(w, "Could not delete thread", http.StatusInternalServerError)
		return
	}
	util.ResponseJSON(w, thread)
}

func closeThreadHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}
	thread := &db.Thread{
		ID: mux.Vars(r)["id"],
	}
	if err := db.CloseThread(ctx, thread); err != nil {
		log.Errorf(ctx, "%v", err)
		util.ResponseError(w, "Could not close thread", http.StatusInternalServerError)
		return
	}

	clearCachedItem(ctx, "thread:"+thread.ID)

	util.ResponseJSON(w, thread)
}

func createCommentHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if err := db.GetUser(ctx, user); err != nil || user.Banned {
		util.ResponseError(w, "Not allowed to comment", http.StatusUnauthorized)
		return
	}

	// Read comment
	var comment db.Comment
	if err := util.ReadJSON(r.Body, &comment); err != nil || len(comment.Text) == 0 {
		util.ResponseError(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if len(comment.Text) > 2048 {
		util.ResponseError(w, "Comment is too long", http.StatusBadRequest)
		return
	}

	// Update fields
	comment.UserID = user.ID
	comment.ThreadID = mux.Vars(r)["id"]
	comment.CreatedAt = time.Now().UTC()
	comment.Text = template.HTMLEscapeString(comment.Text)

	// Insert comment into database
	if err := db.PutComment(ctx, &comment); err != nil {
		log.Errorf(ctx, "%v", err)
		util.ResponseError(w, "Could not create comment", http.StatusInternalServerError)
		return
	}

	// Send reply email
	if len(comment.ReplyTo) != 0 {
		parentComment := &db.Comment{ID: comment.ReplyTo}
		if err := db.GetComment(ctx, parentComment); err == nil {
			recipient := &db.User{ID: parentComment.UserID}
			if err = db.GetUser(ctx, recipient); err == nil {
				if rateLimitEmailTo(ctx, recipient.Email) {
					email := strings.Replace(commentReplyEmail,
						"{{url}}",
						"https://outcrawl.com/"+comment.ThreadID+"/",
						1)
					sendTo(ctx, "You've got a reply", email, recipient.Email)
				}
			}
		}
	}

	clearCachedItem(ctx, "thread:"+comment.ThreadID)

	util.ResponseJSON(w, comment)
}

func deleteCommentHandler(ctx context.Context, user *db.User, w http.ResponseWriter, r *http.Request) {
	if !user.Admin {
		util.ResponseError(w, "Must be an admin", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID := vars["threadId"]
	commentID := vars["id"]

	comment := &db.Comment{
		ID:       commentID,
		ThreadID: threadID,
	}
	if err := db.DeleteComment(ctx, comment); err != nil {
		util.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clearCachedItem(ctx, "thread:"+threadID)

	util.ResponseJSON(w, comment)
}
