package web

import (
	"encoding/json"
	"log"
	"net/http"
)

type ResponseBody struct {
	Data    interface{} `json:"data,omitempty"`
	Status  int         `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Self    string      `json:"self,omitempty"`
	Prev    string      `json:"prev,omitempty"`
	Next    string      `json:"next,omitempty"`
}

func respondCreated(url string, w http.ResponseWriter) {
	if url != "" {
		w.Header().Set("location", url)
	}
	send(http.StatusCreated, nil, w)
}

func respondNoContent(w http.ResponseWriter) {
	send(http.StatusNoContent, nil, w)
}

func respondOK(body interface{}, w http.ResponseWriter) {
	sendData(http.StatusOK, body, w)
}

func respondStatus(status int, w http.ResponseWriter) {
	sendMessage(status, "", w)
}

func badRequest(message string, w http.ResponseWriter) {
	sendMessage(http.StatusBadRequest, message, w)
}

func respondError(status int, err error, w http.ResponseWriter) {
	sendMessage(status, err.Error(), w)
}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers",
		"Origin, X-Requested-With, Content-Type, Accept")
}

func sendMessage(status int, message string, w http.ResponseWriter) error {
	return send(status, ResponseBody{Status: status, Message: message}, w)
}

func sendData(status int, data interface{}, w http.ResponseWriter) error {
	return send(status, ResponseBody{Status: status, Data: data}, w)
}

func send(status int, body interface{}, w http.ResponseWriter) error {
	if status < 200 || status >= 300 {
		log.Printf("RESPONSE: status %d, BODY: %#v\n", status, body)
	}

	addCorsHeaders(w)
	if body == nil || status == http.StatusNoContent {
		w.WriteHeader(status)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	content, err := json.Marshal(body)
	if err == nil {
		if _, err = w.Write(content); err != nil {
			log.Printf("Error writing response: %#v\n", body)
		}
	}
	return err
}
