package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/smarx/go-cowsay"
)

const maxRemainingCount = 240

type Item struct {
	actions        map[string]bool
	remainingCount int
	text           string
	channel        string
	timestamp      string
	currentText    string
}

func Process(item *Item, api *slack.Client) {
	text := item.text
	if item.actions["marquee"] {
		text += "     "
		howMuch := (maxRemainingCount - item.remainingCount) % len(text)
		text = text[howMuch:] + text[:howMuch]
	}
	if item.actions["blink"] {
		if item.remainingCount%2 == 0 {
			text = strings.Repeat(" ", len(text))
		}
	}
	if item.actions["cow"] {
		text = "```\n" + cowsay.Format(text) + "\n```"
	} else if item.actions["blink"] || item.actions["marquee"] {
		text = "```" + text + "```"
	}
	item.remainingCount -= 1
	if item.currentText == text {
		item.remainingCount = 0
	}
	item.currentText = text
	go api.UpdateMessage(item.channel, item.timestamp, text)
}

func ProcessForever(incoming chan *Item, api *slack.Client) {
	ticker := time.Tick(250 * time.Millisecond)
	items := make([]*Item, 0, 10)
	for {
		select {
		case <-ticker:
			newItems := make([]*Item, 0, len(items))
			for i := range items {
				Process(items[i], api)
				if items[i].remainingCount > 0 {
					newItems = append(newItems, items[i])
				}
			}
			items = newItems
		case item := <-incoming:
			items = append(items, item)
		}
	}
}

func FindTags(text string) (tags map[string]bool, stripped string) {
	tags = make(map[string]bool)
	stripped = text
	var possible = []string{"blink", "marquee", "important", "cow"}

	for i := range possible {
		var tag = possible[i]
		var re = regexp.MustCompile(fmt.Sprintf("&lt;%[1]s&gt;(.*)&lt;/%[1]s&gt;", tag))
		var newtext = re.ReplaceAllString(stripped, "$1")
		if newtext != stripped {
			tags[tag] = true
			stripped = newtext
		}
	}
	return
}

func main() {
	api := slack.New(os.Getenv("TOKEN"))

	itemc := make(chan *Item)
	go ProcessForever(itemc, api)

	response, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
	}

	UserID := response.UserID

	var important *Item
	ackID := -1

	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.AckMessage:
				if ev.ReplyTo == ackID {
					important.timestamp = ev.Timestamp
					ackID = -1
				}
			case *slack.MessageEvent:
				if len(ev.SubType) > 0 || (important != nil && ev.Channel == important.channel && ev.Timestamp == important.timestamp) {
					continue
				}
				if important != nil {
					message := rtm.NewOutgoingMessage(important.currentText, important.channel)
					ackID = message.ID
					rtm.SendMessage(message)
					go api.DeleteMessage(important.channel, important.timestamp)
				}
				if ev.User != UserID {
					continue
				}
				tags, text := FindTags(ev.Text)
				if len(tags) > 0 {
					item := Item{tags, maxRemainingCount, text, ev.Channel, ev.Timestamp, text}
					if item.actions["important"] {
						important = &item
					}
					if (item.actions["cow"] || item.actions["important"]) && len(item.actions) == 1 {
						item.remainingCount = 1
					}
					if len(item.actions) > 0 {
						itemc <- &item
					}
				}
			case *slack.InvalidAuthEvent:
				fmt.Println("Invalid credentials")
				break Loop
			default:
				// ignore everything else
			}
		}
	}
}
