package core

import (
	"database/sql"
	"fmt"
	"strconv"
)

type User struct {
	id            sql.NullInt64
	telegram_id   sql.NullInt64
	telegram_name sql.NullString
}

func createUser(telegram_id int, telegram_name string) (*User, error) {
	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	lastInsertId := 0
	query := `INSERT INTO users(telegram_id, telegram_name) VALUES($1, $2) RETURNING id;`
	result := db.QueryRow(query, strconv.Itoa(telegram_id), telegram_name)
	err = result.Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}

	fmt.Println("lastInsertId")
	fmt.Println(lastInsertId)

	user, err := getUserById(lastInsertId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func getUserById(user_id int) (*User, error) {

	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//Отправляем запрос в БД для подсчета числа уникальных пользователей
	row := db.QueryRow("SELECT id, telegram_id, telegram_name FROM users WHERE id = $1;", user_id)
	user := new(User)
	switch err := row.Scan(&user.id, &user.telegram_id, &user.telegram_name); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return user, nil
	default:
		return nil, err
	}
}

func getUserByTelegramId(telegram_id int) (*User, error) {

	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//Отправляем запрос в БД для подсчета числа уникальных пользователей
	row := db.QueryRow("SELECT id, telegram_id, telegram_name FROM users WHERE telegram_id = $1;", telegram_id)
	user := new(User)
	switch err := row.Scan(&user.id, &user.telegram_id, &user.telegram_name); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return user, nil
	default:
		return nil, err
	}
}
