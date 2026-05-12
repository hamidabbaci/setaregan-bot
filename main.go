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

// تابع ارسال پیام
func sendBotMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("خطا در ارسال پیام به %d: %v", chatID, err)
	}
}

// تابع ذخیره کاربران در فایل
func saveUsers(users map[int64]bool) {
	bytes, _ := json.Marshal(users)
	os.WriteFile("users.json", bytes, 0644)
}

// تابع خواندن کاربران از فایل
func loadUsers() map[int64]bool {
	users := make(map[int64]bool)
	bytes, err := os.ReadFile("users.json")
	if err == nil {
		json.Unmarshal(bytes, &users)
	}
	return users
}

type PushMessage struct {
	Time string
	Text string

}

func loadHistory() []PushMessage {
	var history []PushMessage {
		bytes, err := os.ReadFile("history.json")
		if err == nil {
			json.Unmarshal(bytes, &history)
		}
		return history
	}
}

func main() {
	// بارگذاری فایل مخفی تنظیمات
	err := godotenv.Load()
	if err != nil {
		log.Fatal("خطا: فایل .env پیدا نشد!")
	}

	// خواندن توکن و آیدی از سیستم
	botToken := os.Getenv("BOT_TOKEN")
	adminIDStr := os.Getenv("ADMIN_ID")

	// تبدیل آیدی از متن به عدد
	myAdminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		log.Fatal("خطا: آیدی ادمین در فایل .env باید فقط شامل عدد باشد!")
	}

	baleEndpoint := "https://tapi.bale.ai/bot%s/%s"

	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(botToken, baleEndpoint)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	// خواندن لیست کاربران در زمان شروع
	users := loadUsers()

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
			// بررسی دستور استارت برای کاربران عادی
			if msgText == "/start" {
				if users[senderID] == false {
					users[senderID] = true
					saveUsers(users)
				}
				sendBotMessage(bot, senderID, "سلام! شما عضو ربات ستارگان ترید شدید و پوش مسیج‌ها را دریافت خواهید کرد.")
			}
		} else {
			// بررسی دستورات ادمین
			if strings.HasPrefix(msgText, "push ") {
				finalMessage := strings.TrimPrefix(msgText, "push ")

				for userID := range users {
					sendBotMessage(bot, userID, finalMessage)
					time.Sleep(35 * time.Millisecond)
				}

				sendBotMessage(bot, myAdminID, "پوش مسیج با موفقیت برای همه ارسال شد!")
			} else {
				sendBotMessage(bot, myAdminID, "برای ارسال پوش مسیج بنویسید:\npush پیام شما")
			}
		}
	}
}
