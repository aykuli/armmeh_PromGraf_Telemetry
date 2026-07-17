package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/lpernett/godotenv"
	"github.com/mymmrac/telego"
)

type GrafanaAlert struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	State    string `json:"state"` // Alerting, OK, No Data
	RuleName string `json:"ruleName"`
	RuleURL  string `json:"ruleUrl"`
}

func main() {
	_ = godotenv.Load() // local dev

	botToken := os.Getenv("BOT_TOKEN")
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatIDStr == "" {
		log.Fatal("Ошибка: BOT_TOKEN или TELEGRAM_CHAT_ID не заданы в .env")
	}

	targetChatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Ошибка: некорректный TELEGRAM_CHAT_ID: %v", err)
	}

	
	bot, err := telego.NewBot(botToken)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	if _, err = bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: targetChatID},
		Text:      "Привет, Айнур! Бот начал работать.",
		ParseMode: telego.ModeMarkdown,
	}); err != nil {
		log.Fatalf("Что-то не так. Проверь бота: %v", err)
	}

	log.Printf("err: %v\n", err)

	http.HandleFunc("/alert", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("alert worked")
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var alert GrafanaAlert
		if err := json.Unmarshal(body, &alert); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		statusEmoji := "🟢"
		if alert.State == "Alerting" {
			statusEmoji = "🔴"
		}

		text := fmt.Sprintf(
			"%s *Атеншн!*\n\n*Правило:* %s\n*Статус:* %s\n*Сообщение:* %s\n\n[Открыть в Grafana](%s)",
			statusEmoji, alert.RuleName, alert.State, alert.Message, alert.RuleURL,
		)

		_, err = bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID:    telego.ChatID{ID: targetChatID},
			Text:      text,
			ParseMode: telego.ModeMarkdown,
		})
		log.Printf("Text: %v\n", text)

		log.Printf("ChatID: %v\n", targetChatID)

		if err != nil {
			log.Printf("Ошибка отправки в Telegram: %v\n", err)
			http.Error(w, "Failed to send alert to Telegram", http.StatusInternalServerError)
			return
		}

		// Отвечаем Grafane, что всё ок
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	})

	port := "0.0.0.0:8081"
	log.Printf("Сервер бота запущен локально на http://%s/alert\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
