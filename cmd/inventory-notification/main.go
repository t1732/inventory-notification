package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/PuerkitoBio/goquery"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handler call")
}

func lineHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("lineHandler call")

	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		log.Fatal(err)
		return
	}

	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			log.Fatal(err)
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		switch event.Type {
		// フレンド登録
		case linebot.EventTypeFollow:
			userID := event.Source.UserID
			text := linebot.NewTextMessage("Amazonの在庫を通知します")
			if _, err := bot.PushMessage(userID, text).Do(); err != nil {
				log.Fatal(err)
			}
		// フレンド解除
		case linebot.EventTypeUnfollow:
			userID := event.Source.UserID
			text := linebot.NewTextMessage("またどうぞ！")
			if _, err := bot.PushMessage(userID, text).Do(); err != nil {
				log.Fatal(err)
			}
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Printf("receive message: %s", message.Text)
				text := linebot.NewTextMessage("そうですね")
				if _, err := bot.ReplyMessage(event.ReplyToken, text).Do(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("hookHandler call")

	targetUrl := os.Getenv("TARGET_URL")
	if targetUrl == "" {
		log.Printf("target URL empty.")
		return
	}

	doc, err := goquery.NewDocument(targetUrl)
	if err != nil {
		log.Fatal(err)
		return
	}

	force := r.FormValue("ping") == "true"

	if !force {
		selection := doc.Find("#availability > span.a-color-price")
		if strings.Contains(strings.TrimSpace(selection.Text()), "在庫切れ") {
			log.Printf("在庫なし")
			return
		}

		selection = doc.Find("#merchant-info a")
		if !strings.Contains(strings.TrimSpace(selection.Text()), "Amazon.co.jp") {
			log.Printf("Amazon 出品ではない")
			return
		}
	}

	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		log.Fatal(err)
		return
	}

	texts := []linebot.SendingMessage{}
	texts = append(texts, linebot.NewTextMessage("在庫が復活しました！"), linebot.NewTextMessage(targetUrl))
	if _, err := bot.BroadcastMessage(texts...).Do(); err != nil {
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
