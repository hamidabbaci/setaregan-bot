package main

import (
	"encoding/json"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ساختار دیتایی که قرار است از فرانت‌اند دریافت کنیم
type PushRequest struct {
	Message string `json:"message"`
}

// تابع مدیریت درخواست API برای دریافت تاریخچه
func getHistoryAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	history := loadHistory()
	json.NewEncoder(w).Encode(history)
}

// تابع مدیریت درخواست فرانت‌اند برای ارسال پیام
func sendPushAPI(bot *tgbotapi.BotAPI, users map[int64]bool, adminID int64, history *[]PushMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "فقط متد POST مجاز است", http.StatusMethodNotAllowed)
			return
		}

		var req PushRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Message == "" {
			http.Error(w, "دیتای ارسالی نامعتبر است", http.StatusBadRequest)
			return
		}

		currentTime := time.Now().Format("2006-01-02 15:04:05")
		newRecord := PushMessage{
			Time: currentTime,
			Text: req.Message,
		}

		*history = append(*history, newRecord)
		saveHistory(*history)

		for userID := range users {
			sendBotMessage(bot, userID, req.Message)
			time.Sleep(35 * time.Millisecond)
		}

		sendBotMessage(bot, adminID, "ارسال از پنل وب:\n"+req.Message)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
