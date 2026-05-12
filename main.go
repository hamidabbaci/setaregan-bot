package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// --- توابع کمکی ---
func sendBotMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("خطا در ارسال پیام به %d: %v", chatID, err)
	}
}

func saveUsers(users map[int64]bool) {
	bytes, _ := json.Marshal(users)
	os.WriteFile("users.json", bytes, 0644)
}

func loadUsers() map[int64]bool {
	users := make(map[int64]bool)
	bytes, err := os.ReadFile("users.json")
	if err == nil {
		json.Unmarshal(bytes, &users)
	}
	return users
}

// --- ساختار و توابع تاریخچه (مرحله قبل) ---
type PushMessage struct {
	Time string
	Text string
}

func saveHistory(history []PushMessage) {
	bytes, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile("history.json", bytes, 0644)
}

func loadHistory() []PushMessage {
	var history []PushMessage
	bytes, err := os.ReadFile("history.json")
	if err == nil {
		json.Unmarshal(bytes, &history)
	}
	return history
}

// --- تابع اصلی برنامه ---
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("خطا: فایل .env پیدا نشد!")
	}

	botToken := os.Getenv("BOT_TOKEN")
	adminIDStr := os.Getenv("ADMIN_ID")

	myAdminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		log.Fatal("خطا: آیدی ادمین باید فقط شامل عدد باشد!")
	}

	baleEndpoint := "https://tapi.bale.ai/bot%s/%s"

	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(botToken, baleEndpoint)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	users := loadUsers()

	// --- تغییر ۱: خواندن تاریخچه از فایل در زمان شروع ---
	msgHistory := loadHistory()
	// ---------------------------------------------------

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		senderID := update.Message.Chat.ID
		msgText := update.Message.Text

		if senderID != myAdminID {
			if msgText == "/start" {
				if users[senderID] == false {
					users[senderID] = true
					saveUsers(users)
				}
				sendBotMessage(bot, senderID, "سلام! شما عضو ربات ستارگان ترید شدید و پوش مسیج‌ها را دریافت خواهید کرد.")
			}
		} else {
			if strings.HasPrefix(msgText, "push ") {
				finalMessage := strings.TrimPrefix(msgText, "push ")

				// --- تغییر ۲: ثبت زمان و ذخیره پیام در تاریخچه ---
				currentTime := time.Now().Format("2006-01-02 15:04:05")
				newRecord := PushMessage{
					Time: currentTime,
					Text: finalMessage,
				}

				msgHistory = append(msgHistory, newRecord)
				saveHistory(msgHistory)
				// -------------------------------------------------

				for userID := range users {
					sendBotMessage(bot, userID, finalMessage)
					time.Sleep(35 * time.Millisecond)
				}

				sendBotMessage(bot, myAdminID, "پوش مسیج با موفقیت ارسال شد و در تاریخچه ثبت گردید!")

				// --- تغییر ۳: اضافه شدن دستور history برای ادمین ---
			} else if msgText == "history" {

				if len(msgHistory) == 0 {
					sendBotMessage(bot, myAdminID, "هنوز هیچ پیامی در تاریخچه ثبت نشده است.")
				} else {
					var historyText string = "📜 لیست پیام‌های ارسال شده:\n\n"

					for index, record := range msgHistory {
						historyText += fmt.Sprintf("شماره %d - در تاریخ [%s]:\n%s\n\n", index+1, record.Time, record.Text)
					}

					sendBotMessage(bot, myAdminID, historyText)
				}
				// --------------------------------------------------

			} else {
				sendBotMessage(bot, myAdminID, "برای ارسال پیام بنویسید:\npush پیام شما\n\nبرای دیدن تاریخچه بنویسید:\nhistory")
			}
		}
	}
}
