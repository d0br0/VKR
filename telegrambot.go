package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/makiuchi-d/gozxing"
	gozxingqr "github.com/makiuchi-d/gozxing/qrcode"

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
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Сканирование Qr-code", "Вернуться в главное меню"})
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
				handleQRCodeMessage(bot, update)
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

	// Получаем состояние группы из карты по ID чата
	groupState, ok := groupStates[update.Message.Chat.ID]
	if !ok {
		// Если состояние группы не найдено, создаем новое состояние
		groupState = &GroupState{}
		groupStates[update.Message.Chat.ID] = groupState
	}

	if os.Getenv("DB_SWITCH") == "on" {
		switch groupState.step {
		case 0:
			sendMessage(bot, update.Message.Chat.ID, "Введите название группы:")
			groupState.step++
		case 1:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название группы не может быть пустым. Пожалуйста, введите название группы:")
				return nil
			}
			groupState.groupName = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя классного руководителя:")
			groupState.step++
		case 2:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя классного руководителя не может быть пустым. Пожалуйста, введите имя классного руководителя:")
				return nil
			}
			groupState.classLeader = update.Message.Text

			if err := collectDataGroup(groupState.groupName, groupState.classLeader); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Database error, but bot still working.")
				return fmt.Errorf("collectDataGroup failed: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Группа успешно создана!")
				groupState.step = 0
				groupState.groupName = ""
				groupState.classLeader = ""
			}
		}
		isProcessing = false
	}
	return nil
}

func (us *UserState) makeUser(wg *sync.WaitGroup, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	defer wg.Done()
	isProcessing = true

	// Получаем состояние пользователя из карты по ID чата
	userState, ok := userStates[update.Message.Chat.ID]
	if !ok {
		// Если состояние пользователя не найдено, создаем новое состояние
		userState = &UserState{}
		userStates[update.Message.Chat.ID] = userState
	}

	if os.Getenv("DB_SWITCH") == "on" {
		switch userState.step {
		case 0:
			sendMessage(bot, update.Message.Chat.ID, "Введите тэг пользователя:")
			userState.step++
		case 1:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название тэга не может быть пустым. Пожалуйста, введите название тэга:")
				return nil
			}
			userState.username = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите название роли:")
			userState.step++
		case 2:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название роли не может быть пустым. Пожалуйста, введите название роли:")
				return nil
			}
			userState.role = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите ФИО:")
			userState.step++
		case 3:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "ФИО не может быть пустым. Пожалуйста, введите ФИО:")
				return nil
			}
			userState.fio = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя группы:")
			userState.step++
		case 4:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя группы не может быть пустым. Пожалуйста, введите имя группы:")
				return nil
			}
			userState.groupName = update.Message.Text

			if err := collectDataUsers(userState.username, userState.role, userState.fio, userState.groupName); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Database error, but bot still working.")
				return fmt.Errorf("collectDataGroup failed: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Пользователь успешно создан!")
				userState.step = 0
				userState.groupName = ""
				userState.username = ""
				userState.role = ""
				userState.fio = ""
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

func handleQRCodeMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
	isProcessing = true
	if update.Message.Photo != nil {
		fileID := (*update.Message.Photo)[len(*update.Message.Photo)-1].FileID
		fileURL, err := bot.GetFileDirectURL(fileID)
		if err != nil {
			log.Println("Error getting file URL:", err)
			return
		}
		sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
		resp, err := http.Get(fileURL)
		if err != nil {
			log.Println("Error getting image:", err)
			return
		}
		defer resp.Body.Close()
		sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Println("Error decoding image:", err)
			return
		}
		sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			log.Println("Error converting image to binary bitmap:", err)
			return
		}
		sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
		qrReader := gozxingqr.NewQRCodeReader()
		result, err := qrReader.Decode(bmp, nil)
		if err != nil {
			log.Println("Error reading QR code:", err)
			return
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, result.GetText())
		bot.Send(msg)
		isProcessing = false
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
