package core

import (
	"database/sql"
	"fmt"
)

func CreateRating(contact_id int, user_id int, rating int) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO ratings(contact_id, user_id, rating) VALUES ($1, $2, $3);`
	result, err := db.Exec(query, contact_id, user_id, rating)
	fmt.Println(result)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
