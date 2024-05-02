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

	err = db.Ping()
	if err != nil {
		return err
	}

	//Создаем SQL запрос
	data := `INSERT INTO users (username, role, fio, groupName) VALUES($1, $2, $3, $4);`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data, `@`+username, role, fio, groupName); err != nil {
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

	err = db.Ping()
	if err != nil {
		return err
	}

	// SQL запрос для добавления новой группы в таблицу "group"
	query := `INSERT INTO structure (group_name, class_leader) VALUES ($1, $2)`

	// Выполнение SQL запроса
	if _, err := db.Exec(query, groupName, classLeader); err != nil {
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
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS structure(ID SERIAL PRIMARY KEY, GROUP_NAME TEXT, CLASS_LEADER TEXT); `); err != nil {
		return err
	}

	return nil
}

func getNumberOfUsers() (int64, error) {

	var count int64

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	//Отправляем запрос в БД для подсчета числа уникальных пользователей
	row := db.QueryRow("SELECT COUNT(DISTINCT username) FROM users;")
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
