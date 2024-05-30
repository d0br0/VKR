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

func collectDataUsers(userName string, role string, fio string, groupName string, childName string) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Проверяем, существует ли пользователь
	userExists := `SELECT EXISTS(SELECT 1 FROM users WHERE user_Name=$1);`
	var exists bool
	err = db.QueryRow(userExists, userName).Scan(&exists)
	if err != nil {
		log.Printf("Error executing query: %v\n", err)
		return err
	}

	if exists {
		log.Printf("User %s already exists\n", userName)
		return nil
	}

	//Создаем SQL запрос
	data := `INSERT INTO users(user_Name, role, fio, group_Name, child_Name) VALUES($1, $2, $3, $4, $5);`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data, userName, role, fio, groupName, childName); err != nil {
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
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return "", err
	}
	defer db.Close()

	var role string
	err = db.QueryRow("SELECT ROLE FROM users WHERE USER_NAME = $1", username).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", username)
			return "", fmt.Errorf("user not found")
		}
		log.Printf("Error executing query: %v", err)
		return "", err
	}

	log.Printf("Found role for user %s: %s", username, role)

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

func getStudents(teacherName string, date string, pairNumber string) ([]string, error) {
	// Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Выполняем SQL запрос, который объединяет таблицы magazine и users по полю STUDENT_NAME
	// и выбирает поле FIO из таблицы users для каждого студента
	rows, err := db.Query(`SELECT users.FIO 
		FROM magazine 
		JOIN users ON magazine.STUDENT_NAME = users.USER_NAME
		WHERE DATE = $1 AND PAIR_NUMBER = $2 AND TEACHER_NAME = $3`, date, pairNumber, teacherName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []string
	// Проходимся по всем строкам результата запроса
	for rows.Next() {
		var student string
		// Считываем значение поля FIO в переменную student
		if err := rows.Scan(&student); err != nil {
			return nil, err
		}
		// Добавляем student в список students
		students = append(students, student)
	}

	// Проверяем, не было ли ошибок при чтении строк
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Возвращаем список students
	return students, nil
}

func compareWithDatabase(qrData string, username string, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Запрашиваем данные из таблицы magazine
	rows, err := db.Query(`SELECT DATE, PAIR_NUMBER, TEACHER_NAME, REPEAT FROM magazine;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	//Сравниваем данные из QR-кода с данными из таблицы
	var matchFound bool
	for rows.Next() {
		var date, pairNumber, teacherName, repeat string
		err = rows.Scan(&date, &pairNumber, &teacherName, &repeat)
		if err != nil {
			return err
		}

		//Если данные совпадают, записываем их в новую строку вместе с username
		if qrData == fmt.Sprintf("%s, %s, %s, %s", date, pairNumber, teacherName, repeat) {
			_, err = db.Exec(`INSERT INTO magazine (DATE, PAIR_NUMBER, TEACHER_NAME, REPEAT, STUDENT_NAME) VALUES ($1, $2, $3, $4, $5);`, date, pairNumber, teacherName, repeat, username)
			if err != nil {
				return err
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Данные успешно записаны.")
			bot.Send(msg)
			matchFound = true
			break
		}
	}

	if !matchFound {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Данные не совпадают.")
		bot.Send(msg)
	}

	return nil
}

func lookStudent(teacherName string, date string, pairNumber string) ([]string, error) {
	var studentName string
	var absentStudentsUsernames []string

	// Запрос в таблицу magazine
	err := db.QueryRow("SELECT STUDENT_NAME FROM magazine WHERE DATE = $1 AND PAIR_NUMBER = $2 AND TEACHER_NAME = $3", date, pairNumber, teacherName).Scan(&studentName)
	if err != nil {
		return nil, err
	}

	if studentName != "" {
		var groupName string

		// Запрос в таблицу users
		err = db.QueryRow("SELECT GROUP_NAME FROM users WHERE username = $1", studentName).Scan(&groupName)
		if err != nil {
			return nil, err
		}

		rows, err := db.Query("SELECT username FROM structure WHERE GROUP_NAME = $1", groupName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				return nil, err
			}

			if username != studentName {
				var userName string
				err = db.QueryRow("SELECT user_name FROM users WHERE CHILD_NAME = $1", username).Scan(&userName)
				if err != nil {
					return nil, err
				}
				absentStudentsUsernames = append(absentStudentsUsernames, userName)
			}
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return absentStudentsUsernames, nil
}

func createTable() error {

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаём таблицу users
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (ID SERIAL PRIMARY KEY, USER_NAME TEXT, ROLE TEXT, FIO TEXT, GROUP_NAME TEXT, CHILD_NAME TEXT);`); err != nil {
		return err
	}
	//Создаём таблицу magazine
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS magazine (ID SERIAL PRIMARY KEY, DATE TEXT, PAIR_NUMBER TEXT, TEACHER_NAME TEXT, REPEAT TEXT, STUDENT_NAME TEXT);`); err != nil {
		return err
	}
	//Создаём таблицу group
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS  structure (ID SERIAL PRIMARY KEY, GROUP_NAME TEXT, CLASS_LEADER TEXT);`); err != nil {
		return err
	}

	return nil
}
