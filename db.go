package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
)

var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var user = os.Getenv("POSTGRES_USER")
var password = os.Getenv("POSTGRES_PASSWORD")
var dbname = os.Getenv("POSTGRES_DB")
var sslmode = os.Getenv("SSLMODE")

var dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

// Собираем данные полученные ботом
func collectDataUsers(userName string, role string, fio string, groupName string) error {

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаем SQL запрос
	data := `INSERT INTO users(user_Name, role, fio, group_Name) VALUES($1, $2, $3, $4);`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data, userName, role, fio, groupName); err != nil {
		log.Printf("Error executing query: %v\n", err)
		return err
	}

	return nil
}

func collectDataGroup(groupName string, classLeader string) error {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// SQL запрос для добавления новой группы в таблицу "group"
	query := `INSERT INTO structure(group_name, class_leader) VALUES($1, $2);`

	// Выполнение SQL запроса
	if _, err := db.Exec(query, groupName, classLeader); err != nil {
		log.Printf("Error executing query: %v\n", err)
		return err
	}

	return nil
}

func getUserRole(username string) (string, error) {
	log.Printf("Getting role for user: %s", username) // Логирование имени пользователя

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Printf("Error opening database: %v", err) // Логирование ошибки при открытии базы данных
		return "", err
	}
	defer db.Close()

	var role string
	err = db.QueryRow("SELECT ROLE FROM users WHERE USER_NAME = $1", username).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", username) // Логирование, если пользователь не найден
			return "", fmt.Errorf("user not found")
		}
		log.Printf("Error executing query: %v", err) // Логирование ошибки при выполнении запроса
		return "", err
	}

	log.Printf("Found role for user %s: %s", username, role) // Логирование найденной роли

	return role, nil
}

func recordToDatabase(username string, date string, para string, repeat int) error {
	// Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// Подготавливаем запрос на обновление данных
	updateStmt, err := db.Prepare("UPDATE magazine SET REPEAT = $1 WHERE DATE = $2 AND PAIR_NUMBER = $3 AND TEACHER_NAME = $4 AND STUDENT_NAME = $5")
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	// Выполняем запрос на обновление
	result, err := updateStmt.Exec(repeat, date, para, username, "")
	if err != nil {
		return err
	}

	// Проверяем, была ли обновлена какая-либо строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Если ни одна строка не была обновлена, вставляем новую
	if rowsAffected == 0 {
		insertStmt, err := db.Prepare("INSERT INTO magazine(DATE, PAIR_NUMBER, TEACHER_NAME, REPEAT, STUDENT_NAME) VALUES($1, $2, $3, $4, $5)")
		if err != nil {
			return err
		}
		defer insertStmt.Close()

		_, err = insertStmt.Exec(date, para, username, repeat, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func compareWithDatabase(qrData string, username string, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	// Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// Извлекаем данные из базы данных
	row := db.QueryRow("SELECT * FROM magazine WHERE CONCAT(DATE, PAIR_NUMBER, TEACHER_NAME, REPEAT) = $1", qrData)
	var date, pairNumber, teacherName string
	var repeat int
	err = row.Scan(&date, &pairNumber, &teacherName, &repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если данных нет в базе данных
			log.Println("Данные QR-кода не найдены в базе данных")
			sendMessage(bot, update.Message.Chat.ID, "Запись не произошла, сфотографируйте заново qr code и пришлите сюда")
		} else {
			// Если произошла другая ошибка
			log.Println("Ошибка при извлечении данных из базы данных:", err)
		}
		return err
	}

	// Если данные найдены в базе данных
	log.Println("Данные QR-кода найдены в базе данных")

	// Вставляем новую строку с данными QR-кода и именем пользователя
	stmt, err := db.Prepare("INSERT INTO magazine(DATE, PAIR_NUMBER, TEACHER_NAME, REPEAT, STUDENT_NAME) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(date, pairNumber, teacherName, repeat, username)
	if err != nil {
		log.Println("Ошибка при вставке данных в базу данных:", err)
		return err
	}

	sendMessage(bot, update.Message.Chat.ID, "Запись прошла успешно")

	return nil
}

func createTable() error {

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаём таблицу users
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (ID SERIAL PRIMARY KEY, USER_NAME TEXT, ROLE TEXT, FIO TEXT, GROUP_NAME TEXT);`); err != nil {
		return err
	}
	//Создаём таблицу magazine
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS magazine (ID SERIAL PRIMARY KEY, DATE TEXT, PAIR_NUMBER TEXT, TEACHER_NAME TEXT, REPEAT INT, STUDENT_NAME TEXT);`); err != nil {
		return err
	}
	//Создаём таблицу group
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS  structure (ID SERIAL PRIMARY KEY, GROUP_NAME TEXT, CLASS_LEADER TEXT);`); err != nil {
		return err
	}

	return nil
}
