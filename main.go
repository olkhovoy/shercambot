package main

import "C"
import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	//"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	//"github.com/giorgisio/goav/avutil"
)

func main() {

	rootdir := os.Getenv("CAM_ROOT_DIR")
	if len(rootdir) == 0 {
		rootdir, _ = os.Getwd()
		if len(rootdir) == 0 {
			rootdir = "./"
		}
	}
	cams := GetCams(rootdir)
	log.Printf("Cameras root dir: %v", cams)

	// Register all audio-video formats and codecs
	avformat.AvRegisterAll()

	os.Environ()
	settings := tb.Settings{
		URL:    "http://127.0.0.1:8081",
		Token:  apiToken,
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

	err = bot.SetCommands([]tb.Command{
		{"/privet", "Как и /привет - проверка связи, бот должен ответить"},
		{"/photo", "Отправить фото"},
		{"/video", "Отправить видео (путь на диске сервера или URL: http:// или file://)"},
	})
	if err != nil {
		log.Fatalf("could not create Telegram Bot instance: %v", err)
		return
	}

	bot.Handle("/privet", func(msg *tb.Message) {
		bot.Send(msg.Sender, "сам privet")
	})

	bot.Handle("/привет", func(msg *tb.Message) {
		bot.Send(msg.Sender, "сам привет")
	})

	bot.Handle("/photo", func(msg *tb.Message) {
		arg := msg.Payload
		var fil tb.File
		if (arg[:4] == "http") || (arg[:4] == "file") {
			fil = tb.FromURL(arg)
		} else {
			fil = tb.FromDisk(arg)
		}
		pic := &tb.Photo{File: fil}
		res, err := bot.Send(msg.Sender, pic)
		var txt string
		if err == nil {
			txt = fmt.Sprintf("Ответ: %v", res)
		} else {
			txt = fmt.Sprintf("Ошиба: %v", err)
		}
		bot.Send(msg.Sender, txt)
	})

	bot.Handle("/video", func(msg *tb.Message) {
		/*
			fmt.Println(a.OnDisk()) // true
			fmt.Println(a.InCloud()) // false

			// Will upload the file from disk and send it to recipient
			bot.Send(recipient, a)

			// Next time you'll be sending this very *Audio, Telebot won't
			// re-upload the same file but rather utilize its Telegram FileID
			bot.Send(otherRecipient, a)

			fmt.Println(a.OnDisk()) // true
			fmt.Println(a.InCloud()) // true
			fmt.Println(a.FileID) // <telegram file id: ABC-DEF1234ghIkl-zyx57W2v1u123ew11>
		*/

		// Get chat
		if msg.Chat == nil {
			bot.Send(msg.Sender, "Could not get target chat")
			return
		}

		// Get filename from message, compose url
		vid := msg.Payload
		if len(vid) == 0 {
			bot.Send(msg.Sender, "Empty video filename")
			return
		}
		//str := msg.Payload // "pik_183_2021-03-28_00-05-49.ts.mp4"
		//vidre := regexp.MustCompile("(?<fullpath>(?<path>/?.*/|)(?<filename>(?<name>pik_(?<cam>\\d\\d\\d)_(?<datetime>(?<date>\\d\\d\\d\\d-\\d\\d-\\d\\d)_(?<time>\\d\\d-\\d\\d-\\d\\d))[\\._]?(?<videosuffix>.*))\\.(?<videoformat>.*?))")
		vidre := regexp.MustCompile("(/?.*/|)(pik_(\\d\\d\\d)_\\d\\d\\d\\d-\\d\\d-\\d\\d_\\d\\d-\\d\\d-\\d\\d)\\.(.*?)")
		m := vidre.FindAllStringSubmatchIndex(vid, -1)
		if m == nil {
			bot.Send(msg.Sender, "Could not parse filename from args")
			return
		}

		//str := msg.Payload // "pik_183_2021-03-28_00-05-49.ts.mp4"

		//videoRE := regexp.MustCompile("(?<videofilename>(?<videoname>pik_(?<cam>\\d\\d\\d)_(?<videodatetime>(?<videodate>\\d\\d\\d\\d-\\d\\d-\\d\\d)_(?<videotime>\\d\\d-\\d\\d-\\d\\d))[\\._]?(?<videosuffix>.*))\\.(?<videoformat>.*?))$")
		//m := videoRE.FindSubmatch([]byte(videofile))
		//text := fmt.Sprintf("%v", m)
		//reg := regexp.MustCompile("((https?://.*?|file:.*?//|)(/.*?)(pik_\d\d\d_.*?)\.mp4),?\W*?)+")
		//m := reg.FindAllStringSubmatchIndex(string(msg.Payload),1)
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

	bot.Handle(tb.OnPhoto, func(msg *tb.Message) {
		obj := msg.Photo
		info := fmt.Sprintf(
			"FileID: %d, FileLocal: %s, FileURL: %s\n\n%v",
			obj.FileID,
			obj.FileLocal,
			obj.FileURL,
			obj,
		)
		bot.Send(
			msg.Sender,
			info,
		)
	})

	bot.Handle(tb.OnVideo, func(msg *tb.Message) {
		obj := msg.Video
		info := fmt.Sprintf(
			"VideoFileID: %d, VideoLocal: %s, VideoURL: %s,ThumbFileID: %d, ThumbLocal: %s, ThumbURL: %s\n%v",
			obj.FileID,
			obj.FileLocal,
			obj.FileURL,
			obj.Thumbnail.FileID,
			obj.Thumbnail.FileLocal,
			obj.Thumbnail.FileURL,
			obj,
		)
		bot.Send(
			msg.Sender,
			info,
		)
	})

	bot.Handle(tb.OnDocument, func(msg *tb.Message) {
		obj := msg.Document
		info := fmt.Sprintf(
			"DocFileID: %v, DocLocal: %v, DocURL: %v, ThumbFileID: %v, ThumbLocal: %v, ThumbURL: %v\n\n%v",
			obj.File.FileID,
			obj.File.FileLocal,
			obj.File.FileURL,
			obj.Thumbnail.FileID,
			obj.Thumbnail.FileLocal,
			obj.Thumbnail.FileURL,
			obj,
		)
		bot.Send(
			msg.Sender,
			info,
		)
	})

	bot.Handle(tb.OnAnimation, func(msg *tb.Message) {
		bot.Send(
			msg.Sender,
			fmt.Sprint(msg.Animation),
		)
	})

	bot.Start()
}
