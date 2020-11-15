package handler

import (
	"net/http"
	"log"

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

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
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
