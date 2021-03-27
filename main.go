package main

import (
    "log"
	"os"
	"time"

    tb "gopkg.in/tucnak/telebot.v2"
)

func main() {

	//env := os.Environ()
	settings := tb.Settings{
		URL:    "",//os.Getenv("SHERCAMBOT_API_URL"), // if field is empty it equals to "https://api.telegram.org".
		Token:  os.Getenv("SHERCAMBOT_API_TOKEN"),
		Poller: &tb.LongPoller{

			Timeout: 10 * time.Second,
		},
	}
	if len(settings.Token) == 0 {
		log.Fatalf("environment valiable SHERCAMBOT_API_TOKEN should have a value: %v", settings)
		return
	}

    bot, err := tb.NewBot(settings)
	if err != nil {
		log.Fatalf("could not create Telegram Bot instance: %v", err)
		return
	}

    bot.Handle("/привет", func(msg *tb.Message) {
		bot.Send(msg.Sender, "Сам, Привет.")
    })

    bot.Start()
}
