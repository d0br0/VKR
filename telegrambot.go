package main

import (
	"log"
	"os"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func telegramBot() {

	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Button 1"),
				tgbotapi.NewKeyboardButton("Button 2"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Button 3"),
				tgbotapi.NewKeyboardButton("Button 4"),
			),
		)

		bot.Send(msg)
	}
}

func main() {

	time.Sleep(1 * time.Minute)

	//Создаем таблицу
	if os.Getenv("CREATE_TABLE") == "yes" {

		if os.Getenv("DB_SWITCH") == "on" {

			if err := createTable(); err != nil {

				panic(err)
			}
		}
	}

	time.Sleep(1 * time.Minute)

	//Вызываем бота
	telegramBot()
}
