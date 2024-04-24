package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var adminPassword string = "1029384756"

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

		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

			switch update.Message.Text {
			case "/start":
				sendMenu(bot, update.Message.Chat.ID, "Здравствуй! Я бот для учёта посещаемости. Выбери кто ты.", []string{"Преподаватель", "Студент", "Администратор"})
			case "Преподаватель":
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание студента", "Вернуться в главное меню"})
			case "Студент":
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Сканирование QR-code", "Вернуться в главное меню"})
			case "Администратор":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите пароль администратора."))
			case adminPassword:
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Ввести название учебного заведения", "Создание группы", "Создание студента", "Вернуться в главное меню", "Число пользователей"})
			case "Вернуться в главное меню":
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Преподаватель", "Студент", "Администратор"})
			case "Число пользователей":
				handleNumberOfUsers(update, bot)
			case "Создание группы":
				makeGroup(update, bot)
			default:
				sendDB(update, bot)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На такую комманду я не запрограммирован..")
				bot.Send(msg)
			}
		}
	}
}

func sendMenu(bot *tgbotapi.BotAPI, chatID int64, message string, options []string) {
	msg := tgbotapi.NewMessage(chatID, message)
	var keyboard [][]tgbotapi.KeyboardButton
	for _, option := range options {
		row := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(option)}
		keyboard = append(keyboard, row)
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard...)
	bot.Send(msg)
}

func handleNumberOfUsers(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	if os.Getenv("DB_SWITCH") == "on" {

		//Присваиваем количество пользоватьелей использовавших бота в num переменную
		num, err := getNumberOfUsers()
		if err != nil {

			//Отправлем сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка базы данных")
			bot.Send(msg)
		}

		//Создаем строку которая содержит колличество пользователей использовавших бота
		ans := fmt.Sprintf("%d Число пользователей:", num)

		//Отправлем сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, ans)
		bot.Send(msg)
	} else {

		//Отправлем сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "База данных не подключена, я не могу сообщить число пользователей :(")
		bot.Send(msg)
	}
	return nil
}

func makeGroup(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	var groupName string
	var classLeader string
	var step string
	if os.Getenv("DB_SWITCH") == "on" {
		if step == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите название группы:")
			bot.Send(msg)
			step = "groupName"
			return nil
		}

		if step == "groupName" {
			groupName = update.Message.Text
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите имя классного руководителя:")
			bot.Send(msg)
			step = "classLeader"
			return nil
		}

		if step == "classLeader" {
			classLeader = update.Message.Text

			if err := collectDataGroup(groupName, classLeader); err != nil {
				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error, but bot still working.")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Группа успешно создана!")
				bot.Send(msg)
			}

			groupName = ""
			classLeader = ""
			step = ""
		}
	}
	return nil
}

func sendDB(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	if os.Getenv("DB_SWITCH") == "on" {

		//Отправляем username, chat_id, message, answer в БД
		if err := collectDataUsers(update.Message.Chat.UserName, update.Message.Chat.ID, update.Message.Text); err != nil {

			//Отправлем сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error, but bot still working.")
			bot.Send(msg)
		}
	}
	return nil
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
