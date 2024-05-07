package main

import (
	"database/sql"
	"fmt"
	"os"

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
func collectDataUsers(username string, role string, fio string, groupName string) error {

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаем SQL запрос
	data := `INSERT INTO users(username, role, fio, groupName) VALUES($1, $2, $3, $4);`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data, `@`+username, role, fio, groupName); err != nil {
		return err
	}

	return nil
}

func collectDataGroup(chatID int64, groupName string, classLeader string) error {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// SQL запрос для добавления новой группы в таблицу "group"
	query := `INSERT INTO structure(group_name, class_leader, chat_id) VALUES($1, $2, $3)`

	// Выполнение SQL запроса
	if _, err := db.Exec(query, groupName, classLeader, chatID); err != nil {
		return err
	}

	return nil
}

func collectTesting(apples string, chear string) error {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// SQL запрос для добавления новой группы в таблицу "group"
	query := `INSERT INTO structure(apples, chear) VALUES($1, $2)`

	// Выполнение SQL запроса
	if _, err := db.Exec(query, apples, chear); err != nil {
		return err
	}

	return nil
}

// Создаем таблицы: users, magazine, group в БД при подключении к ней
func createTable() error {

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаём таблицу users
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(ID SERIAL PRIMARY KEY, USERNAME TEXT, ROLE TEXT, FIO TEXT, GROUP_NAME TEXT);`); err != nil {
		return err
	}
	//Создаём таблицу magazine
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS magazine(ID SERIAL PRIMARY KEY, DATE DATE, TIME TIME, STUDENT_ID INT);`); err != nil {
		return err
	}
	//Создаём таблицу group
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS  structure(ID SERIAL PRIMARY KEY, GROUP_NAME TEXT, CLASS_LEADER TEXT, CHAT_ID INT,);`); err != nil {
		return err
	}
	//Создаём таблицу testing
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS testing(ID SERIAL PRIMARY KEY, APPLES TEXT, CHEAR TEXT);`); err != nil {
		return err
	}

	_, err = db.Exec("ALTER TABLE structure ADD COLUMN CHAT_ID INT")
	if err != nil {
		return err
	}

	return nil
}

func getNumberOfUsers() (int64, error) {

	var count int64

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 1, err
	}
	defer db.Close()

	//Отправляем запрос в БД для подсчета числа уникальных пользователей
	row := db.QueryRow("SELECT COUNT(DISTINCT username) FROM users;")
	err = row.Scan(&count)
	if err != nil {
		return 2, err
	}

	return count, nil
}
