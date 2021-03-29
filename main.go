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
			log.Println("Error: Couldn't find stream information.")

			// Close input file and free context
			avctx.AvformatCloseInput()
			return
		}

		avstreams := avctx.Streams()
		if len(avstreams) == 0 {
			bot.Send(msg.Sender, "Could not find any streams: "+filename)
			return
		}
		beststream := avstreams[0]
		if len(avstreams) > 1 {
			bestarea := 0
			for _, stream := range avstreams {
				codec := stream.Codec()
				ct := codec.GetCodecType()
				if ct == 1 {
					ct = 0
				}
				width, height := codec.GetWidth(), codec.GetHeight()
				area := width * height
				if area >= bestarea {
					beststream = stream
					bestarea = area
				}
			}
		}

		//md := avctx.Metadata()
		//vid.Caption = md["title"]
		//if md["title"]; !ok { vid.Caption = filename }
		vid := &tb.Video{
			File:              tb.FromDisk(filename),
			Width:             beststream.Codec().GetWidth(),
			Height:            beststream.Codec().GetHeight(),
			Duration:          int(beststream.Duration() * int64(beststream.AvgFrameRate().Num()) / int64(beststream.AvgFrameRate().Den())),
			Caption:           filename,
			Thumbnail:         nil,
			SupportsStreaming: true,
			MIME:              "video/mp4",
			FileName:          filename,
		}

		//vid.Caption = avctx.Metadata()
		//vid.Duration = int(beststream.Duration() * int64(beststream.AvgFrameRate().Num()) / int64(beststream.AvgFrameRate().Den()))
		//vid.Width = beststream.Codec().GetWidth()
		//vid.Height = beststream.Codec().GetHeight()

		//vid := &tb.Video{File: tb.FromDisk(url)}
		bot.Send(msg.Sender, vid)
		//bot.Send(msg.Sender, "Сам, Привет.")
	})

	bot.Start()
}
