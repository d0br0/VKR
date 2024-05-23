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
	"time"

	"github.com/makiuchi-d/gozxing"
	gozxingqr "github.com/makiuchi-d/gozxing/qrcode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	qrcode "github.com/skip2/go-qrcode"
)

var us = &UserState{}
var gs = &GroupState{}
var ss = &StudentState{}
var gqs = &GenerateState{}
var sqs = &ScanState{}
var userStates = make(map[int64]*UserState)
var groupStates = make(map[int64]*GroupState)
var studentStates = make(map[int64]*StudentState)
var generateStates = make(map[int64]*GenerateState)
var scanStates = make(map[int64]*ScanState)

type QrCodeResponse struct {
	Data string `json:"data"`
}

type UserState struct {
	username  string
	role      string
	fio       string
	groupName string
	step      int
}

type GroupState struct {
	nameGroup   string
	classLeader string
	step        int
}

type StudentState struct {
	username  string
	fio       string
	groupName string
	step      int
}

type GenerateState struct {
	para int
	step int
}

type ScanState struct {
	step int
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

		if update.Message.Text != "" {
			if role == "Администратор" {
				switch update.Message.Text {
				case "/start":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Число пользователей", "Вернуться в главное меню"})
				case "Вернуться в главное меню":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Число пользователей", "Вернуться в главное меню"})
				case "Число пользователей":
					handleNumberOfUsers(update, bot)
				case "Создание группы":
					gs.makeGroup(update, bot)
				case "Создание пользователя":
					us.makeUser(update, bot)
				case "Стоп":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание пользователя", "Вернуться в главное меню", "Число пользователей"})
					timerControl <- true
				case "Отметить присутствующих":
					gqs.markStudents(bot, update, timerControl)
				default:
					sendMessage(bot, update.Message.Chat.ID, "Извините, на такую команду я не запрограмирован.")
				}
			} else if role == "Преподаватель" {
				switch update.Message.Text {
				case "/start":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание студента", "Посмотреть журнал"})
				case "Вернуться в главное меню":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание студента", "Посмотреть журнал"})
				case "Создание студента":
					ss.makeStudent(update, bot)
				case "Стоп":
					sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Отметить присутствующих", "Создание группы", "Создание студента", "Вернуться в главное меню"})
					timerControl <- true
				case "Отметить присутствующих":
					gqs.markStudents(bot, update, timerControl)
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

func (gs *GroupState) makeGroup(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {

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
			groupState.nameGroup = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя классного руководителя:")
			groupState.step++
		case 2:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя классного руководителя не может быть пустым. Пожалуйста, введите имя классного руководителя:")
				return nil
			}
			groupState.classLeader = update.Message.Text

			if err := collectDataGroup(groupState.nameGroup, groupState.classLeader); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Database error, but bot still working.")
				return fmt.Errorf("collectDataGroup failed: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Группа успешно создана!")
				groupState.step = 0
				groupState.nameGroup = ""
				groupState.classLeader = ""
				delete(groupStates, update.Message.Chat.ID)
			}
		}
	}
	return nil
}

func (us *UserState) makeUser(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {

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
				delete(userStates, update.Message.Chat.ID)
			}
		}
	}
	return nil
}

func (ss *StudentState) makeStudent(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {

	// Получаем состояние пользователя из карты по ID чата
	studentState, ok := studentStates[update.Message.Chat.ID]
	if !ok {
		// Если состояние пользователя не найдено, создаем новое состояние
		studentState = &StudentState{}
		studentStates[update.Message.Chat.ID] = studentState
	}

	if os.Getenv("DB_SWITCH") == "on" {
		switch studentState.step {
		case 0:
			sendMessage(bot, update.Message.Chat.ID, "Введите тэг пользователя:")
			studentState.step++
		case 1:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Название тэга не может быть пустым. Пожалуйста, введите название тэга:")
				return nil
			}
			studentState.username = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите ФИО:")
			studentState.step++
		case 2:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "ФИО не может быть пустым. Пожалуйста, введите ФИО:")
				return nil
			}
			studentState.fio = update.Message.Text
			sendMessage(bot, update.Message.Chat.ID, "Введите имя группы:")
			studentState.step++
		case 3:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Имя группы не может быть пустым. Пожалуйста, введите имя группы:")
				return nil
			}
			studentState.groupName = update.Message.Text

			if err := collectDataUsers(studentState.username, "Студент", studentState.fio, studentState.groupName); err != nil {
				sendMessage(bot, update.Message.Chat.ID, "Database error, but bot still working.")
				return fmt.Errorf("collectDataGroup failed: %w", err)
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Пользователь успешно создан!")
				studentState.step = 0
				studentState.groupName = ""
				studentState.username = ""
				studentState.fio = ""
				delete(studentStates, update.Message.Chat.ID)
			}
		}
	}
	return nil
}

func (gqs *GenerateState) markStudents(bot *tgbotapi.BotAPI, update tgbotapi.Update, timerControl chan bool) error {
	generateState, ok := generateStates[update.Message.Chat.ID]
	if !ok {
		// Если состояние пользователя не найдено, создаем новое состояние
		generateState = &GenerateState{}
		generateStates[update.Message.Chat.ID] = generateState
	}
	switch generateState.step {
	case 0:
		sendMenu(bot, update.Message.Chat.ID, "Выберите номер группы", []string{"1"})
		generateState.step++
	case 1:
		sendMenu(bot, update.Message.Chat.ID, "Выберите номер пары", []string{"1", "2", "3", "4", "5", "6", "7"})
		generateState.step++
	case 2:
		sendMenu(bot, update.Message.Chat.ID, "Нажмите стоп, когда закончите отмечать", []string{"Стоп"})
		qrCodeData, err := generateQRCode("Присутствующий")
		if err != nil {
			log.Println("Ошибка при генерации QR-кода:", err)
			return err
		}
		err = sendQRToTelegramChat(bot, update.Message.Chat.ID, qrCodeData)
		if err != nil {
			log.Println("Ошибка при отправке QR-кода в чат:", err)
			return err
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

func (sqs *ScanState) handleQRCodeMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	scanState, ok := scanStates[update.Message.Chat.ID]
	if !ok {
		//Если состояние пользователя не найдено, создаем новое состояние
		scanState = &ScanState{}
		scanStates[update.Message.Chat.ID] = scanState
	}
	if os.Getenv("DB_SWITCH") == "on" {
		switch scanState.step {
		case 0:
			sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
			scanState.step++
		case 1:
			sendMessage(bot, update.Message.Chat.ID, "Сделайте фото QR-Code и отправьте в чат.")
			if update.Message.Photo != nil {
				scanState.step++
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Пожалуйста, отправьте фото QR-кода.")
			}
		case 2:
			fileID := (*update.Message.Photo)[len(*update.Message.Photo)-1].FileID
			fileURL, err := bot.GetFileDirectURL(fileID)
			if err != nil {
				log.Println("Ошибка при получении URL файла:", err)
				return err
			}

			// download and decode image file
			resp, err := http.Get(fileURL)
			if err != nil {
				log.Println("Ошибка при загрузке изображения:", err)
				return err
			}
			defer resp.Body.Close()

			img, _, err := image.Decode(resp.Body)
			if err != nil {
				log.Println("Ошибка при декодировании изображения:", err)
				return err
			}

			// prepare BinaryBitmap
			bmp, err := gozxing.NewBinaryBitmapFromImage(img)
			if err != nil {
				log.Println("Ошибка при преобразовании изображения в двоичный растровый формат:", err)
				return err
			}

			// decode image
			qrReader := gozxingqr.NewQRCodeReader()
			result, err := qrReader.Decode(bmp, nil)
			if err != nil {
				log.Println("Ошибка при чтении QR-кода:", err)
				return err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Результат сканирования: %s", result))
			bot.Send(msg)
			scanState.step = 0
			delete(scanStates, update.Message.Chat.ID)
		}
	}

	return nil
}

func magazine()

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
