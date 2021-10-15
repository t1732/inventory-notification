package handler

import (
	"log"
	"net/http"
	"os"
	"regexp"
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

	client := &http.Client{}

        req, err := http.NewRequest("GET", targetUrl, nil)
        if err != nil {
                log.Fatalln(err)
        }

	userAgent := os.Getenv("USER_AGENT")
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36"
	}
        req.Header.Set("User-Agent", userAgent)

        res, err := client.Do(req)
        if err != nil {
                log.Fatalln(err)
        }

        defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	force := r.FormValue("ping") == "true"
	if !force {
		title := doc.Find("title").Text()
		log.Printf(title)

		availability := doc.Find("#availability").Text()
		availability = strings.TrimSpace(availability)
		log.Printf(availability)
		r := regexp.MustCompile(`お取り扱いできません`)
		if r.MatchString(availability) {
			log.Printf("在庫なし")
			return
		}

		merchantInfo := doc.Find("#merchant-info").Text()
		merchantInfo = strings.TrimSpace(merchantInfo)
		log.Printf("merchantInfo: %s", merchantInfo)
		r = regexp.MustCompile(`Amazon\.co\.jp`)
		if !r.MatchString(merchantInfo) {
			log.Printf("Amazon 出品ではない")
			return
		}
	}

	n, err := notifier.New()
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(500)
		return
	}

	texts := []string{"在庫が復活したかもしれません", targetUrl}
	if _, err = n.BroadcastMessage(texts).Do(); err != nil {
		log.Fatal(err)
	}
}
