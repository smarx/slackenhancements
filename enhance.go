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
	Actions        map[string]bool
	RemainingCount int
	Text           string
	Channel        string
	Timestamp      string
}

func Process(item *Item, api *slack.Client) {
	text := item.Text
	if item.Actions["marquee"] {
		text += "     "
		runes := []rune(text)
		howMuch := (maxRemainingCount - item.RemainingCount) % len(runes)
		text = string(runes[howMuch:]) + string(runes[:howMuch])
	}
	if item.Actions["blink"] {
		if item.RemainingCount%2 == 0 {
			text = strings.Repeat(" ", len([]rune(text)))
		}
	}
	if item.Actions["cow"] {
		text = "```\n" + cowsay.Format(text) + "\n```"
	} else if item.Actions["blink"] || item.Actions["marquee"] {
		text = "```" + text + "```"
	}
	item.RemainingCount -= 1
	go api.UpdateMessage(item.Channel, item.Timestamp, text)
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
				if items[i].RemainingCount > 0 {
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
	var possible = []string{"blink", "marquee", "cow"}

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
	token := os.Getenv("TOKEN")

	if len(token) == 0 {
		fmt.Println("Missing auth token. Please visit https://api.slack.com/docs/oauth-test-tokens and get a token for your account. Then set the TOKEN environment variable to that value.")
		return
	}

	api := slack.New(token)

	itemc := make(chan *Item)
	go ProcessForever(itemc, api)

	response, err := api.AuthTest()
	if err != nil {
		log.Fatal(err)
	}

	UserID := response.UserID

	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Enhanced! You can now use <blink>, <marquee>, and <cow> to annoy your coworkers.")
			case *slack.MessageEvent:
				if len(ev.SubType) > 0 || ev.ReplyTo > 0 {
					continue
				}
				if ev.User != UserID {
					continue
				}
				tags, text := FindTags(ev.Text)
				if len(tags) > 0 {
					item := Item{
						Actions:        tags,
						RemainingCount: maxRemainingCount,
						Text:           text,
						Channel:        ev.Channel,
						Timestamp:      ev.Timestamp}
					if item.Actions["cow"] && len(item.Actions) == 1 {
						// Don't keep editing the cow if it's not going to change.
						item.RemainingCount = 1
					}
					if len(item.Actions) > 0 {
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
