package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	qrcode "github.com/skip2/go-qrcode"
)

var us = &UserState{}
var gs = &GroupState{}
var adminPassword string = "1029384756"
var isProcessing bool
var wg sync.WaitGroup
var userStates = make(map[int64]*UserState)
var groupStates = make(map[int64]*GroupState)

type UserState struct {
	username  string
	role      string
	fio       string
	groupName string
	step      int
}

type GroupState struct {
	groupName   string
	classLeader string
	step        int
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

	// Создаем канал для управления таймером
	timerControl := make(chan bool)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text != "" && !isProcessing {
			switch update.Message.Text {
			case "/start":
				sendMenu(bot, update.Message.Chat.ID, "Здравствуй! Я бот для учёта посещаемости. Выбери кто ты.", []string{"Преподаватель", "Студент", "Администратор"})
			case "Преподаватель":
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Вернуться в главное меню"})
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
				wg.Add(1)
				gs.makeGroup(&wg, update, bot)
				wg.Wait()
			case "Создание пользователя":
				wg.Add(1)
				us.makeUser(&wg, update, bot)
				wg.Wait()
			case "Стоп":
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание студента", "Вернуться в главное меню"})
				timerControl <- true
			case "Отметить присутствующих":
				sendMenu(bot, update.Message.Chat.ID, "Нажмите стоп, когда закончите отмечать", []string{"Стоп"})
				qrCodeData, err := generateQRCode("Присутствующий")
				if err != nil {
					log.Println("Ошибка при генерации QR-кода:", err)
					return
				}
				err = sendQRToTelegramChat(bot, update.Message.Chat.ID, qrCodeData)
				if err != nil {
					log.Println("Ошибка при отправке QR-кода в чат:", err)
					return
				}
				go func() {
					ticker := time.NewTicker(1 * time.Minute)
					for {
						select {
						case <-ticker.C:
							qrCodeData, err := generateQRCode("Присутствующий")
							if err != nil {
								log.Println("Ошибка при генерации QR-кода:", err)
								return
							}
							err = sendQRToTelegramChat(bot, update.Message.Chat.ID, qrCodeData)
							if err != nil {
								log.Println("Ошибка при отправке QR-кода в чат:", err)
								return
							}
						case <-timerControl:
							ticker.Stop()
							return
						}
					}
				}()
			case "Сканирование Qr-code":
				sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code, и отрпавьте в чат")
				qrText, err := scanQRCode(update.Message.Text)
				if err != nil {
					log.Printf("Ошибка при сканировании QR-кода: %v", err)
					continue
				}
				// Сохраняем информацию в базу данных
				// Здесь вы можете добавить свою логику для сохранения данных
				sendMessage(bot, update.Message.Chat.ID, "Информация из QR-кода: "+qrText)
			default:
				sendMessage(bot, update.Message.Chat.ID, "Извините, на такую команду я не запрограмирован.")
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

func (gs *GroupState) makeGroup(wg *sync.WaitGroup, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	defer wg.Done()
	isProcessing = true
	defer func() {
		isProcessing = false
	}()
	// Получаем состояние группы из карты по ID чата
	groupState, ok := groupStates[update.Message.Chat.ID]
	if !ok {
		// Если состояние группы не найдено, создаем новое состояние
		groupState = &GroupState{}
		groupStates[update.Message.Chat.ID] = groupState
	}

	if os.Getenv("DB_SWITCH") == "on" {
		switch gs.step {
		case 0:
			sendMessage(bot, update.Message.Chat.ID, "Введите название группы:")
			gs.step++
		case 1:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название группы не может быть пустым. Пожалуйста, введите название группы:")
				return nil
			}
			gs.groupName = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя классного руководителя:")
			gs.step++
		case 2:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя классного руководителя не может быть пустым. Пожалуйста, введите имя классного руководителя:")
				return nil
			}
			gs.classLeader = update.Message.Text
			// Здесь вы вызываете функцию collectDataGroup с параметрами groupName и classLeader.
			// Если она завершится успешно, вы отправите сообщение о успешном создании группы.
			// В противном случае вы сообщите об ошибке базы данных.
			if err := collectDataGroup(gs.groupName, gs.classLeader); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Ошибка базы данных, но бот продолжает работать.")
				return fmt.Errorf("collectDataGroup не удалось: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Группа успешно создана!")
				// Сбросим состояние, чтобы можно было создать новую группу.
				gs.step = 0
				gs.groupName = ""
				gs.classLeader = ""
			}
			isProcessing = false
		}
	}
	return nil
}

func (us *UserState) makeUser(wg *sync.WaitGroup, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	defer wg.Done()
	isProcessing = true
	defer func() {
		isProcessing = false
	}()
	userState, ok := userStates[update.Message.Chat.ID]
	if !ok {
		// Если состояние пользователя не найдено, создаем новое состояние
		userState = &UserState{}
		userStates[update.Message.Chat.ID] = userState
	}
	if os.Getenv("DB_SWITCH") == "on" {
		if us.step == 0 {
			sendMessage(bot, update.Message.Chat.ID, "Введите тэг пользователя:")
			//us.step++
		} else if us.step == 1 {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название тэга не может быть пустым. Пожалуйста, введите название тэга:")
				return nil
			}
			us.username = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите название роли:")
			//us.step++
		} else if us.step == 2 {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название роли не может быть пустым. Пожалуйста, введите название роли:")
				return nil
			}
			us.role = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите ФИО:")
			//us.step++
		} else if us.step == 3 {
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "ФИО не может быть пустым. Пожалуйста, введите ФИО:")
				return nil
			}
			us.fio = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя группы:")
			//us.step++
		} else if us.step == 4 {
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
				us.step = 0
				us.groupName = ""
				us.username = ""
				us.role = ""
				us.fio = ""
			}
		}
		isProcessing = false
	}
	return nil
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func generateQRCode(text string) ([]byte, error) {
	// Generate QR code
	qr, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		log.Println("Error generating QR code:", err)
		return nil, err
	}

	// Create a buffer to hold the PNG data
	buf := new(bytes.Buffer)

	// Write QR code to buffer as PNG
	err = qr.Write(256, buf)
	if err != nil {
		log.Println("Error saving QR code:", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func sendQRToTelegramChat(bot *tgbotapi.BotAPI, chatID int64, qrCodeData []byte) error {
	// Создание нового файла для отправки в чат
	fileBytes := tgbotapi.FileBytes{Name: "qr-code.png", Bytes: qrCodeData}
	photo := tgbotapi.NewPhotoUpload(chatID, fileBytes)

	// Отправка QR-кода в чат
	_, err := bot.Send(photo)
	if err != nil {
		log.Panic(err)
	}

	return nil
}

func scanQRCode(qrData string) (string, error) {
	qr, err := qrcode.New(qrData, qrcode.Highest)
	if err != nil {
		return "", err
	}
	return qr.ToSmallString(false), nil
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
