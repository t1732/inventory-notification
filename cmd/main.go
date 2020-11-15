package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/t1732/inventory-notification/internal/notifier"
)

func handleFollow(client *notifier.LineClient, userID string) {
	if _, err := client.PushMessage(userID, "Amazonの在庫を通知します").Do(); err != nil {
		log.Fatal(err)
	}
}

func handleUnfollow(client *notifier.LineClient, userID string) {
	if _, err := client.PushMessage(userID, "またどうぞ！").Do(); err != nil {
		log.Fatal(err)
	}
}

func handleMessage(client *notifier.LineClient, replyToken string) {
	if _, err := client.ReplyMessage(replyToken, "そうですね").Do(); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handler call")
}

func lineHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("lineHandler call")

	client, err := notifier.New()
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(500)
		return
	}

	events, err := client.ParseRequest(w, r)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(500)
		return
	}

	eventAction := &notifier.EventAction{
		Follow: handleFollow,
		Unfollow: handleUnfollow,
		Message: handleMessage,
	}
	client.HandleEvent(events, eventAction)
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("hookHandler call")

	targetUrl := os.Getenv("TARGET_URL")
	if targetUrl == "" {
		log.Printf("target URL empty.")
		return
	}
	log.Printf(targetUrl)

	doc, err := goquery.NewDocument(targetUrl)
	if err != nil {
		log.Fatal(err)
		return
	}

	force := r.FormValue("ping") == "true"
	if !force {
		title := doc.Find("title").Text()
		log.Printf(title)

		availability := doc.Find("#availability > span.a-color-price").Text()
		availability = strings.TrimSpace(availability)
		log.Printf("availability: %s", availability)
		if strings.Contains(availability, "在庫切れ") {
			log.Printf("在庫なし")
			return
		}

		merchantInfo := doc.Find("#merchant-info a").Text()
		merchantInfo = strings.TrimSpace(merchantInfo)
		log.Printf("merchantInfo: %s", merchantInfo)
		if !strings.Contains(merchantInfo, "Amazon.co.jp") {
			log.Printf("Amazon 出品ではない")
			return
		}
	}

	client, err := notifier.New()
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(500)
		return
	}

	texts := []string{"在庫が復活しました！", targetUrl}
	if _, err = client.BroadcastMessage(texts).Do(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Print("starting server...")

	http.HandleFunc("/", handler)
	http.HandleFunc("/callback", lineHandler)
	http.HandleFunc("/hook", hookHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
