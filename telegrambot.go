package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var us = &UserState{}
var gs = &GroupState{}
var adminPassword string = "1029384756"

type UserState struct {
	username  string
	role      string
	fio       string
	groupName string
	step      string
}

type GroupState struct {
	groupName   string
	classLeader string
	step        string
}

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
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Ввести название учебного заведения", "Создание группы", "Создание пользователя", "Вернуться в главное меню", "Число пользователей"})
			case "Вернуться в главное меню":
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Преподаватель", "Студент", "Администратор"})
			case "Число пользователей":
				handleNumberOfUsers(update, bot)
			case "Создание группы":
				gs.makeGroup(update, bot)
			case "Создание пользователя":
				us.makeUser(update, bot)
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

func (gs *GroupState) makeGroup(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	if os.Getenv("DB_SWITCH") == "on" {
		if gs.step == "" {
			sendMessage(bot, update.Message.Chat.ID, "Введите название группы:")
			gs.step = "groupName"
		} else if gs.step == "groupName" {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название группы не может быть пустым. Пожалуйста, введите название группы:")
				return nil
			}
			gs.groupName = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя классного руководителя:")
			gs.step = "classLeader"
		} else if gs.step == "classLeader" {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя классного руководителя не может быть пустым. Пожалуйста, введите имя классного руководителя:")
				return nil
			}
			gs.classLeader = update.Message.Text

			if err := collectDataGroup(gs.groupName, gs.classLeader); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Database error, but bot still working.")
				return fmt.Errorf("collectDataGroup failed: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Группа успешно создана!")
			}

		}
	}
	return nil
}

func (us *UserState) makeUser(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	if os.Getenv("DB_SWITCH") == "on" {
		if us.step == "" {
			sendMessage(bot, update.Message.Chat.ID, "Введите тэг пользователя:")
			us.step = "username"
		} else if us.step == "username" {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название тэга не может быть пустым. Пожалуйста, введите название тэга:")
				return nil
			}
			sendMessage(bot, update.Message.Chat.ID, "Введите название роли:")
			us.step = "role"
		} else if us.step == "role" {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название роли не может быть пустым. Пожалуйста, введите название роли:")
				return nil
			}
			us.role = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите ФИО:")
			us.step = "fio"
		} else if us.step == "fio" {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "ФИО не может быть пустым. Пожалуйста, введите ФИО:")
				return nil
			}
			us.fio = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя группы:")
			us.step = "groupName"
		} else if us.step == "groupName" {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя группы не может быть пустым. Пожалуйста, введите имя группы:")
				return nil
			}
			us.groupName = update.Message.Text

			if err := collectDataUsers(us.username, us.role, us.fio, us.groupName); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Database error, but bot still working.")
				return fmt.Errorf("collectDataGroup failed: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Группа успешно создана!")
			}

		}
	}
	return nil
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
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
