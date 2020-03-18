package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/outcrawl/backend/newsletter"
)

func main() {
	http.HandleFunc("/subscribe", func(writer http.ResponseWriter, request *http.Request) {
		var req newsletter.SubscribeRequest
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(&req); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		if err := newsletter.HandleSubscribe(req); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
