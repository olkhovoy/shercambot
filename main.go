package main

import "C"
import (
	"log"
	"os"
	//"regexp"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
	//"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	//"github.com/giorgisio/goav/avutil"
)

func main() {
	/*
		ПЧ := [2]uint64{0, 1}
		for пч := [-2:]; пч := []uint64{0, 1}[:]; пч[0] < пч[1]; ПЧ = append(ПЧ, пч[0]+пч[1]) {}

		, ПЧ[1], ПЧ[2]; ж, Ж = Ж, жЖ; ПЧ[2] = ПЧ[0] + ПЧ[1] len; ПЧ[ж] > ПЧ[преж]; ПЧ[ж] = ПЧ[ж-1] + ПЧ[преЖ] {
			len(ПЧ)
		}; ПЧ[i] > ПЧ[i-1]ПЧ[< [2]int64{0, 1}; n[0] <= 99999999999; n[1], n[0] = ПЧ[1]+ПЧ[0], n[1] {
			ПЧ = append(ПЧ, )
			fmt.Println(n[0], float64(n[0]) / float64(n[1]))
		}
	*/
	// Register all audio-video formats and codecs
	avformat.AvRegisterAll()

	os.Environ()
	settings := tb.Settings{
		URL:    "https://olkhovoy.com:8081",
		Token:  "17405045__:AAFwExdvmmghJpjJN8HKe9odRU8rwxr2E__",
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
		bot.Send(msg.Sender, "сам привет")
	})

	bot.Handle("/video", func(msg *tb.Message) {

		// Get chat
		if msg.Chat == nil {
			bot.Send(msg.Sender, "Could not get target chat")
			return
		}

		// Get filename from message, compose url
		//str := msg.Payload // "pik_183_2021-03-28_00-05-49.ts.mp4"

		//videoRE := regexp.MustCompile("(?<videofilename>(?<videoname>pik_(?<cam>\\d\\d\\d)_(?<videodatetime>(?<videodate>\\d\\d\\d\\d-\\d\\d-\\d\\d)_(?<videotime>\\d\\d-\\d\\d-\\d\\d))[\\._]?(?<videosuffix>.*))\\.(?<videoformat>.*?))$")
		//m := videoRE.FindSubmatch([]byte(videofile))
		//text := fmt.Sprintf("%v", m)
		vid := msg.Payload
		web := bool(strings.Compare(vid[:4], "http") == 0) || bool(strings.Compare(vid[:4], "file") == 0)

		if !web {
			// Open video file
			avctx := avformat.AvformatAllocContext()
			err := avformat.AvformatOpenInput(&avctx, vid, nil, nil)
			if err != 0 {
				bot.Send(msg.Sender, "Could not open video: "+vid)
				return
			}

			// Retrieve stream information
			if avctx.AvformatFindStreamInfo(nil) < 0 {
				bot.Send(msg.Sender, "Error: couldn't find stream information in the file: "+vid)

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
				bot.Send(msg.Sender, "Error: could not find any video streams in the file: "+vid)
				return
			}
			var file tb.File
			if web {
				file = tb.FromURL(vid)
			} else {
				file = tb.FromDisk(vid)
			}
			codec := beststream.Codec()
			video := &tb.Video{
				File:              file,
				MIME:              "video/mp4",
				Width:             codec.GetWidth(),
				Height:            codec.GetHeight(),
				Caption:           vid,
				FileName:          vid,
				Duration:          int(beststream.NbFrames() * int64(beststream.AvgFrameRate().Den()) / int64(beststream.AvgFrameRate().Num())),
				SupportsStreaming: true,
			}
			bot.Send(msg.Sender, video)
		} else {

			video1 := &tb.Video{
				File:              tb.FromURL(vid),
				MIME:              "video/mp4",
				Width:             1280, //codec.GetWidth(),
				Height:            720,  //codec.GetHeight(),
				Caption:           vid,
				FileName:          vid,
				Duration:          118, //int(beststream.NbFrames() * int64(beststream.AvgFrameRate().Den()) / int64(beststream.AvgFrameRate().Num())),
				SupportsStreaming: true,
			}
			bot.Send(msg.Sender, video1)

		}
		//vid.Caption = avctx.Metadata()
		//vid.Duration = int(beststream.Duration() * int64(beststream.AvgFrameRate().Num()) / int64(beststream.AvgFrameRate().Den()))
		//vid.Width = beststream.Codec().GetWidth()
		//vid.Height = beststream.Codec().GetHeight()

		//vid := &tb.Video{File: tb.FromDisk(url)}
		//bot.Send(msg.Sender, vid)
		//bot.Send(msg.Seznder, "Сам, Привет.")
	})

	bot.Handle(tb.OnVideo, func(msg *tb.Message) {

		bot.Send(msg.Sender, msg.Video)

	})

	bot.Start()
}
