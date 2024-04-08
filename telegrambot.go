package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
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

		//Проверяем что от пользователья пришло именно текстовое сообщение
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

			switch update.Message.Text {
			case "/start":

				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a wikipedia bot, i can search information in a wikipedia, send me something what you want find in Wikipedia.")
				bot.Send(msg)
			case "/number_of_users":

				if os.Getenv("DB_SWITCH") == "on" {

					//Присваиваем количество пользоватьелей использовавших бота в num переменную
					num, err := getNumberOfUsers()
					if err != nil {

						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error.")
						bot.Send(msg)
					}

					//Создаем строку которая содержит колличество пользователей использовавших бота
					ans := fmt.Sprintf("%d peoples used me for search information in Wikipedia", num)

					//Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, ans)
					bot.Send(msg)
				} else {

					//Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database not connected, so i can't say you how many peoples used me.")
					bot.Send(msg)
				}
			default:

				//Устанавливаем язык для поиска в википедии
				language := os.Getenv("LANGUAGE")

				//Создаем url для поиска
				ms, _ := urlEncoded(update.Message.Text)

				url := ms
				request := "https://" + language + ".wikipedia.org/w/api.php?action=opensearch&search=" + url + "&limit=3&origin=*&format=json"

				//Присваем данные среза с ответом в переменную message
				message := wikipediaAPI(request)

				if os.Getenv("DB_SWITCH") == "on" {

					//Отправляем username, chat_id, message, answer в БД
					if err := collectData(update.Message.Chat.UserName, update.Message.Chat.ID, update.Message.Text, message); err != nil {

						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error, but bot still working.")
						bot.Send(msg)
					}
				}

				//Проходим через срез и отправляем каждый элемент пользователю
				for _, val := range message {

					//Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, val)
					bot.Send(msg)
				}
			}
		} else {

			//Отправлем сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Use the words for search.")
			bot.Send(msg)
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
