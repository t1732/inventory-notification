package notifier

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

type LineClient struct {
	bot *linebot.Client
}

type EventAction struct {
	Follow   func(client *LineClient, userID string)
	Unfollow func(client *LineClient, userID string)
	Message  func(client *LineClient, replyToken string)
}

type ErrParse struct {
	msg string
}

func (err *ErrParse) Error() string {
	return fmt.Sprintf(err.msg)
}

func New() (*LineClient, error) {
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	client := &LineClient{}
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		return client, err
	}
	client.bot = bot
	return client, nil
}

func (client *LineClient) ParseRequest(w http.ResponseWriter, r *http.Request) ([]*linebot.Event, error) {
	events, err := client.bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			return nil, &ErrParse{msg: "Invalid Signature."}
		} else {
			return nil, &ErrParse{msg: "Server error"}
		}
	}
	return events, nil
}

func (client *LineClient) HandleEvent(events []*linebot.Event, ea *EventAction) {
	for _, event := range events {
		switch event.Type {
		// フレンド登録
		case linebot.EventTypeFollow:
			userID := event.Source.UserID
			ea.Follow(client, userID)
		// フレンド解除
		case linebot.EventTypeUnfollow:
			userID := event.Source.UserID
			ea.Unfollow(client, userID)
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Printf("receive message: %s", message.Text)
				replyToken := event.ReplyToken
				ea.Message(client, replyToken)
			}
		}
	}
}

func (client *LineClient) PushMessage(targetId string, msg string) *linebot.PushMessageCall {
	text := linebot.NewTextMessage(msg)
	return client.bot.PushMessage(targetId, text)
}

func (client *LineClient) ReplyMessage(replyToken string, msg string) *linebot.ReplyMessageCall {
	text := linebot.NewTextMessage(msg)
	return client.bot.ReplyMessage(replyToken, text)
}

func (client *LineClient) BroadcastMessage(texts []string) *linebot.BroadcastMessageCall {
	messages := []linebot.SendingMessage{}
	for _, text := range texts {
		messages = append(messages, linebot.NewTextMessage(text))
	}
	return client.bot.BroadcastMessage(messages...)
}
