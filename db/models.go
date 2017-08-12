package db

import "time"

type User struct {
	ID     string `json:"id" datastore:"-"`
	Banned bool   `json:"banned"`
	Admin  bool   `json:"admin" datastore:"-"`
	Email  string `json:"email"`
}

type Thread struct {
	ID       string    `json:"id" datastore:"-"`
	Comments []Comment `json:"comments" datastore:"-"`
	Closed   bool      `json:"closed"`
}

type Comment struct {
	ID        string    `json:"id" datastore:"-"`
	ThreadID  string    `json:"threadId"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ReplyTo   string    `json:"replyTo,omitempty"`
	Text      string    `json:"text"`
}
