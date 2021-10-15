package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/t1732/inventory-notification/internal/handler"
)

func main() {
	log.Print("starting server...")

	http.HandleFunc("/", handler.RootHandler)
	http.HandleFunc("/callback", handler.CallbackHandler)
	http.HandleFunc("/hook", handler.HookHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
