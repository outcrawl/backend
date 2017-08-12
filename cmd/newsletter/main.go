package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type SendMessage struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

const url = "https://outcrawl-newsletter.appspot.com/send"

func main() {
	file := os.Args[1]
	subject := os.Args[2]
	token := os.Args[3]

	content, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Cannot read file '%s'\n", file)
		return
	}

	send(subject, string(content), token)

	fmt.Println("success")
}

func send(subject, message, token string) error {
	buf, err := json.Marshal(SendMessage{subject, message})
	if err != nil {
		return err
	}
	_, err = http.Post(fmt.Sprintf("%s?token=%s", url, token), "application/json", bytes.NewBuffer(buf))
	return err
}
