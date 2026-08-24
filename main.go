go func() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Bot is running!")
    })
    http.ListenAndServe(":"+port, nil)
}()package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	urlStore  = make(map[string]string)
	userStore = make(map[int64]string)
	mu        sync.Mutex
)

func getShortID(url string) string {
	hasher := md5.New()
	hasher.Write([]byte(url))
	short := hex.EncodeToString(hasher.Sum(nil))[:8]

	mu.Lock()
	urlStore[short] = url
	mu.Unlock()

	return short
}

func getExecPaths() (string, string) {
	homeDir, _ := os.UserHomeDir()
	localYt := homeDir + "/.local/bin/yt-dlp"
	
	if _, err := os.Stat(localYt); err == nil {
		return localYt, homeDir + "/.local/bin"
	}
	return "yt-dlp", ""
}

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		botToken = "8510258357:AAFbZZOu1gmnGP34mPqLnxMpIS8-ibKtIpE"
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Bot ishga tushdi: %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Botni qayta ishga tushirish"},
		{Command: "help", Description: "Botdan foydalanish bo'yicha yordam"},
	}
	bot.Request(tgbotapi.NewSetMyCommands(commands...))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			text := update.Message.Text
			chatID := update.Message.Chat.ID

			if text == "/start" {
				msg := tgbotapi.NewMessage(chatID, "Assalomu alaykum! 🎬\n\nMenga Instagram, TikTok yoki YouTube havolasini yuboring!")
				bot.Send(msg)
				continue
			}

			if text == "/help" {
				msg := tgbotapi.NewMessage(chatID, "Botdan foydalanish usuli:\n\n1. Video havolasini yuboring.\n2. Pastdagi menyudan kerakli tugmani bosing.")
				bot.Send(msg)
				continue
			}

			if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
				mu.Lock()
				userStore[chatID] = text
				mu.Unlock()

				replyKeyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("🎬 Videoni yuklash"),
						tgbotapi.NewKeyboardButton("🎵 Musiqasini yuklash"),
					),
				)
				replyKeyboard.ResizeKeyboard = true

				msg := tgbotapi.NewMessage(chatID, "Havola qabul qilindi! Nimani yuklab olishni xohlaysiz?")
				msg.ReplyMarkup = replyKeyboard
				bot.Send(msg)
				continue
			}

			mu.Lock()
			lastURL, exists := userStore[chatID]
			mu.Unlock()

			if exists {
				shortID := getShortID(lastURL)
				if text == "🎬 Videoni yuklash" {
					go handleVideoDownload(bot, chatID, lastURL, shortID)
				} else if text == "🎵 Musiqasini yuklash" {
					go handleAudioDownload(bot, chatID, lastURL, shortID)
				}
			} else if text == "🎬 Videoni yuklash" || text == "🎵 Musiqasini yuklash" {
				msg := tgbotapi.NewMessage(chatID, "❌ Avval video havolasini yuboring!")
				bot.Send(msg)
			}
		}

		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			chatID := callback.Message.Chat.ID
			data := callback.Data

			callbackConfig := tgbotapi.NewCallback(callback.ID, "Tanlovingiz qabul qilindi!")
			bot.Request(callbackConfig)

			parts := strings.SplitN(data, "|", 2)
			if len(parts) == 2 {
				action := parts[0]
				shortID := parts[1]

				mu.Lock()
				videoURL, exists := urlStore[shortID]
				mu.Unlock()

				if exists {
					if action == "v" {
						go handleVideoDownload(bot, chatID, videoURL, shortID)
					} else if action == "a" {
						go handleAudioDownload(bot, chatID, videoURL, shortID)
					}
				} else {
					msg := tgbotapi.NewMessage(chatID, "❌ Havola muddati o'tgan. Iltimos, havolani qayta yuboring.")
					bot.Send(msg)
				}
			}
		}
	}
}

func handleVideoDownload(bot *tgbotapi.BotAPI, chatID int64, videoURL string, shortID string) {
	statusMsg := tgbotapi.NewMessage(chatID, "⏳ Video yuklanmoqda, iltimos kuting...")
	sentMsg, _ := bot.Send(statusMsg)

	outputFile := fmt.Sprintf("video_%d_%d.mp4", chatID, time.Now().Unix())
	ytDlpPath, binPath := getExecPaths()

	cmd := exec.Command(ytDlpPath,
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"-f", "b[filesize<50M]/best",
		"-o", outputFile,
		videoURL,
	)
	if binPath != "" {
		cmd.Env = append(os.Environ(), "PATH="+binPath+":"+os.Getenv("PATH"))
	}

	err := cmd.Run()
	defer os.Remove(outputFile)

	if err != nil {
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, "❌ Videoni yuklab bo'lmadi.")
		bot.Send(editMsg)
		return
	}

	bot.Send(tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID))

	inlineButtons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎵 Musiqasini yuklash", "a|"+shortID),
		),
	)

	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(outputFile))
	video.Caption = "🎬 Videongiz tayyor!"
	video.ReplyMarkup = inlineButtons
	bot.Send(video)
}

func handleAudioDownload(bot *tgbotapi.BotAPI, chatID int64, videoURL string, shortID string) {
	statusMsg := tgbotapi.NewMessage(chatID, "⏳ Musiqa ajratib olinmoqda, iltimos kuting...")
	sentMsg, _ := bot.Send(statusMsg)

	outputFile := fmt.Sprintf("audio_%d_%d.mp3", chatID, time.Now().Unix())
	ytDlpPath, binPath := getExecPaths()

	args := []string{
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"-x",
		"--audio-format", "mp3",
		"-o", outputFile,
		videoURL,
	}
	if binPath != "" {
		args = append([]string{"--ffmpeg-location", binPath}, args...)
	}

	cmd := exec.Command(ytDlpPath, args...)
	if binPath != "" {
		cmd.Env = append(os.Environ(), "PATH="+binPath+":"+os.Getenv("PATH"))
	}

	err := cmd.Run()
	defer os.Remove(outputFile)

	if err != nil {
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, "❌ Musiqani ajratib bo'lmadi.")
		bot.Send(editMsg)
		return
	}

	bot.Send(tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID))

	inlineButtons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎬 Videoni yuklash", "v|"+shortID),
		),
	)

	audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(outputFile))
	audio.Caption = "🎵 Musiqangiz tayyor!"
	audio.ReplyMarkup = inlineButtons
	bot.Send(audio)
}
