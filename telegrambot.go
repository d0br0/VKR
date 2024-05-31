package main

import (
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	// Создаем канал для управления таймером
	timerControl := make(chan bool)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		role, err := getUserRole(update.Message.From.UserName)
		if err != nil {
			log.Printf("Error getting user role: %v\n", err)
			continue
		}

		userState, ok := userStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := userState.makeUser(update, bot)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		groupState, ok := groupStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := groupState.makeGroup(update, bot)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		studentState, ok := studentStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := studentState.makeStudent(update, bot)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		generateState, ok := generateStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := generateState.markStudents(update, bot, timerControl)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		scanState, ok := scanStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := scanState.handleQRCodeMessage(update, bot)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		magazineState, ok := magazineStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := magazineState.lookMagazine(update, bot)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		parentState, ok := parentStates[update.Message.Chat.ID]
		if ok {
			// Если есть, обрабатываем сообщение в контексте создания группы
			err := parentState.makeParent(update, bot)
			if err != nil {
				log.Printf("Error making group: %v\n", err)
			}
			continue
		}

		if update.Message.Text != "" {
			if role == "Администратор" {
				switch update.Message.Text {
				case "/start":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Создание родителя", "Журнал"})
				case "Вернуться в главное меню":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Создание родителя", "Журнал"})
				case "Создание группы":
					gs.makeGroup(update, bot)
				case "Создание пользователя":
					us.makeUser(update, bot)
				case "Стоп":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Создание родителя", "Журнал"})
					timerControl <- true
				case "Отметить присутствующих":
					gqs.markStudents(update, bot, timerControl)
				case "Журнал":
					ms.lookMagazine(update, bot)
				case "Создание родителя":
					ps.makeParent(update, bot)
				default:
					sendMessage(bot, update.Message.Chat.ID, "Извините, на такую команду я не запрограмирован.")
				}
			} else if role == "Преподаватель" {
				switch update.Message.Text {
				case "/start":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Журнал", "Создание студента", "Создание родителя"})
				case "Вернуться в главное меню":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Журнал", "Создание студента", "Создание родителя"})
				case "Создание студента":
					ss.makeStudent(update, bot)
				case "Стоп":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Журнал", "Создание студента", "Создание родителя"})
					timerControl <- true
				case "Отметить присутствующих":
					gqs.markStudents(update, bot, timerControl)
				case "Журнал":
					ms.lookMagazine(update, bot)
				case "Создание родителя":
					ps.makeParent(update, bot)
				default:
					sendMessage(bot, update.Message.Chat.ID, "Извините, на такую команду я не запрограмирован.")
				}
			} else if role == "Студент" {
				switch update.Message.Text {
				case "/start":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Сканирование Qr-code"})
				case "Вернуться в главное меню":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Сканирование Qr-code"})
				case "Сканирование Qr-code":
					sqs.handleQRCodeMessage(update, bot)
				default:
					sendMessage(bot, update.Message.Chat.ID, "Извините, на такую команду я не запрограмирован.")
				}
			} else if role == "Родитель" {
				switch update.Message.Text {
				case "/start":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Посмотреть посещаемость ребёнка"})
				case "Посмотреть посещаемость ребёнка":
					cs.lookChildren(update, bot)
				}
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
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
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
