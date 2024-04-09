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
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "/start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Здравствуй, "+update.Message.From.FirstName+"! Я бот для учёта посещаемости. Выбери кто ты.")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Преподаватель"),
						tgbotapi.NewKeyboardButton("Студент"),
						tgbotapi.NewKeyboardButton("Администратор"),
					),
				)
				bot.Send(msg)
			}
		} else {
			switch update.Message.Text {
			case "Преподаватель":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбирете действие:")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Отметить присутствующих"),
						tgbotapi.NewKeyboardButton("Создание группы"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Создание студента"),
						tgbotapi.NewKeyboardButton("Вернуться в главное меню"),
					),
				)
				bot.Send(msg)
			case "Студент":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбирете действие:")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Сканирование QR-code"),
						tgbotapi.NewKeyboardButton("Вернуться в главное меню"),
					),
				)
				bot.Send(msg)
			case "Администратор":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбирете действие:")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Ввести название учебного заведения"),
						tgbotapi.NewKeyboardButton("Создание группы"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Создание студента"),
						tgbotapi.NewKeyboardButton("Вернуться в главное меню"),
					),
				)
				bot.Send(msg)
			case "Вернуться в главное меню":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбирете действие:")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Преподаватель"),
						tgbotapi.NewKeyboardButton("Студент"),
						tgbotapi.NewKeyboardButton("Администратор"),
					),
				)
				bot.Send(msg)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На такую комманду я не запрограммирован..")
				bot.Send(msg)
			}
		}
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
