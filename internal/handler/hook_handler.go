package handler

import (
	"os"
	"net/http"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/t1732/inventory-notification/internal/notifier"
)

func HookHandler(w http.ResponseWriter, r *http.Request) {
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
