package handler

import (
	"net/http"
	"log"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handler call")
}
