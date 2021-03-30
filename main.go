package main

import "C"
import (
	"fmt"
	"log"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
	//"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	//"github.com/giorgisio/goav/avutil"
)

func main() {

	for n := [2]int64{0, 1}; n[0] <= 99999999999; n[1], n[0] = n[1]+n[0], n[1] {
		fmt.Println(n[0], float64(n[0])/float64(n[1]))
	}

	// Register all audio-video formats and codecs
	avformat.AvRegisterAll()

	settings := tb.Settings{
		URL:    "https://olkhovoy.com:8081", //os.Getenv("TELEGRAM_BOT_API_URL"), // if field is empty it equals to "https://api.telegram.org".
		Token:  os.Getenv("TELEGRAM_BOT_API_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
	if len(settings.Token) == 0 {
		log.Fatalf("environment variable TELEGRAM_BOT_API_TOKEN should have a value: %v", settings)
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

	bot.Handle("/video", func(msg *tb.Message) {

		// Get chat
		if msg.Chat == nil {
			bot.Send(msg.Sender, "Could not get target chat")
			return
		}

		// Get filename from message, compose url
		url := "/home/ao/V/PIK183/" + msg.Payload
		filename := "/home/ao/V/PIK183/pik_183_2021-03-28_00-05-49.ts.mp4"

		// Open video file
		var d *avformat.Dictionary
		avctx := avformat.AvformatAllocContext()
		err := avformat.AvformatOpenInput(&avctx, filename, nil, &d)
		if err != 0 {
			bot.Send(msg.Sender, "Could not open video: "+url)
			return
		}

		// Retrieve stream information
		if avctx.AvformatFindStreamInfo(&d) < 0 {
			bot.Send(msg.Sender, "Error: couldn't find stream information in the file: "+filename)

			// Close input file and free context
			avctx.AvformatCloseInput()
			return
		}

		// Select best stream (by number of pixels)
		avstreams := avctx.Streams()
		var beststream *avformat.Stream
		if len(avstreams) > 0 {
			var bestarea int
			for _, stream := range avstreams {
				codec := stream.Codec()
				if codec != nil {
					width, height := codec.GetWidth(), codec.GetHeight()
					area := width * height
					if area > bestarea { // compare stream resolutions [Width * Height]
						bestarea = area
						beststream = stream
					}
				}
			}
		}
		if beststream == nil {
			bot.Send(msg.Sender, "Error: could not find any video streams in the file: "+filename)
			return
		}
		codec := beststream.Codec()
		video := &tb.Video{
			File:              tb.FromURL("https://olkhovoy.com/" + filename),
			MIME:              "video/mp4",
			Width:             codec.GetWidth(),
			Height:            codec.GetHeight(),
			Caption:           filename,
			FileName:          filename,
			Duration:          int(beststream.NbFrames() * int64(beststream.AvgFrameRate().Num()) / int64(beststream.AvgFrameRate().Den())),
			SupportsStreaming: true,
		}
		bot.Send(bot.Me, video)

		//vid.Caption = avctx.Metadata()
		//vid.Duration = int(beststream.Duration() * int64(beststream.AvgFrameRate().Num()) / int64(beststream.AvgFrameRate().Den()))
		//vid.Width = beststream.Codec().GetWidth()
		//vid.Height = beststream.Codec().GetHeight()

		//vid := &tb.Video{File: tb.FromDisk(url)}
		//bot.Send(msg.Sender, vid)
		//bot.Send(msg.Seznder, "Сам, Привет.")
	})

	bot.Handle(tb.OnVideo, func(msg *tb.Message) {

		bot.Send(msg.Chat, msg.Video)

	})

	bot.Start()
}
