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
var ms = &MagazineState{}
var userStates = make(map[int64]*UserState)
var groupStates = make(map[int64]*GroupState)
var studentStates = make(map[int64]*StudentState)
var generateStates = make(map[int64]*GenerateState)
var scanStates = make(map[int64]*ScanState)
var magazineStates = make(map[int64]*MagazineState)

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
	para   string
	step   int
	repeat int
}

type ScanState struct {
	step int
}

type MagazineState struct {
	date string
	pair string
	step int
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
			sendMenu(bot, update.Message.Chat.ID, "Выбирете роль:", []string{"Студент", "Преподаватель", "Администратор"})
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
				sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Вернуться в главное меню"})
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

func (gqs *GenerateState) markStudents(update tgbotapi.Update, bot *tgbotapi.BotAPI, timerControl chan bool) error {
	t := time.Now().UTC()
	date := t.Format("02.01.2006")
	username := update.Message.From.UserName
	generateState, ok := generateStates[update.Message.Chat.ID]
	if !ok {
		generateState = &GenerateState{}
		generateStates[update.Message.Chat.ID] = generateState
	}
	if os.Getenv("DB_SWITCH") == "on" {
		switch generateState.step {
		case 0:
			sendMenu(bot, update.Message.Chat.ID, "Выберите номер пары", []string{"1", "2", "3", "4", "5", "6", "7"})
			generateState.step++
		case 1:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Номер пары не может быть пустым. Пожалуйста, введите название тэга:")
				return nil
			}
			generateState.para = update.Message.Text

			sendMenu(bot, update.Message.Chat.ID, "Нажмите стоп, когда закончите отмечать", []string{"Стоп"})

			allVars := fmt.Sprintf("%s, %s, %s, %d", date, gqs.para, username, gqs.repeat)
			qrCodeData, err := generateQRCode(allVars)
			if err != nil {
				log.Println("Ошибка при генерации QR-кода:", err)
				return err
			}
			err = sendQRToTelegramChat(bot, update.Message.Chat.ID, qrCodeData)
			if err != nil {
				log.Println("Ошибка при отправке QR-кода в чат:", err)
				return err
			}
			//Запись в базу данных
			err = recordToDatabase(username, date, gqs.para, gqs.repeat)
			if err != nil {
				log.Println("Ошибка при записи в базу данных:", err)
				return err
			}

			go func() {
				ticker := time.NewTicker(1 * time.Minute)
				for {
					select {
					case <-ticker.C:
						gqs.repeat++
						var allVars = fmt.Sprintf("%s, %s, %s, %d", date, gqs.para, username, gqs.repeat)
						qrCodeData, err := generateQRCode(allVars)
						if err != nil {
							log.Println("Ошибка при генерации QR-кода:", err)
							return
						}
						err = sendQRToTelegramChat(bot, update.Message.Chat.ID, qrCodeData)
						if err != nil {
							log.Println("Ошибка при отправке QR-кода в чат:", err)
							return
						}
						// Запись в базу данных

						if err = recordToDatabase(username, date, gqs.para, gqs.repeat); err != nil {
							log.Println("Ошибка при записи в базу данных:", err)
							return
						}
					case <-timerControl:
						ticker.Stop()
						return
					}
				}
			}()

			delete(generateStates, update.Message.Chat.ID)
		}
	}
	return nil
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
	username := update.Message.From.UserName
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
			if update.Message.Photo != nil {
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

				err = compareWithDatabase(result.String(), username, update, bot)
				if err != nil {
					log.Println("Ошибка при сравнении данных:", err)
					return err
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Результат сканирования: %s", result))
				bot.Send(msg)

				delete(scanStates, update.Message.Chat.ID)
				//scanState.step++
			} else {
				sendMessage(bot, update.Message.Chat.ID, "Пожалуйста, отправьте фото QR-кода.")
			}
		}
	}
	return nil
}

func (ms *MagazineState) lookMagazine(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	username := update.Message.From.UserName
	magazineState, ok := magazineStates[update.Message.Chat.ID]
	if !ok {
		//Если состояние пользователя не найдено, создаем новое состояние
		magazineState = &MagazineState{}
		magazineStates[update.Message.Chat.ID] = magazineState
	}
	if os.Getenv("DB_SWITCH") == "on" {
		switch magazineState.step {
		case 0:
			sendMessage(bot, update.Message.Chat.ID, "Введите дату в формате ДД.ММ.ГГГГ")
			magazineState.step++
		case 1:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Дата не модет быть пустой. Пожалуйста, введите дату:")
				return nil
			}
			magazineState.date = update.Message.Text
			sendMenu(bot, update.Message.Chat.ID, "Выберите номер пары", []string{"1", "2", "3", "4", "5", "6", "7"})
			magazineState.step++
		case 2:
			if update.Message.Text == "" {
				sendMessage(bot, update.Message.Chat.ID, "Номер пары не может быть пустым. Пожалуйста, введите название тэга:")
				return nil
			}
			magazineState.pair = update.Message.Text
			// Получаем студентов из базы данных
			students, err := getStudents(username, magazineState.date, magazineState.pair)
			if err != nil {
				log.Println("Ошибка при получении студентов из базы данных:", err)
				return err
			}

			// Если студентов нет, отправляем сообщение об этом
			if len(students) == 0 {
				sendMessage(bot, update.Message.Chat.ID, "Студентов нет.")
			} else {
				// Иначе, выводим имена студентов
				for _, student := range students {
					sendMessage(bot, update.Message.Chat.ID, student)
				}
			}
			delete(magazineStates, update.Message.Chat.ID)
			sendMenu(bot, update.Message.Chat.ID, "Выбирете действие:", []string{"Вернуться в главное меню"})
		}
	}
	return nil
}
