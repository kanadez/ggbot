package core

import (
	"database/sql"
	"errors"
	"fmt"
	"unicode/utf8"
)

type Comment struct {
	Id      sql.NullInt64
	Comment sql.NullString
}

func CreateComment(contact_id int, user_id int, comment string) error {
	if utf8.RuneCountInString(comment) > MAX_COMMENT_LEN {
		return errors.New("Comment length is more than 0!")
	}

	if utf8.RuneCountInString(comment) == 0 {
		return errors.New("Comment is empty!")
	}

	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO comments(contact_id, user_id, comment) VALUES ($1, $2, $3);`
	result, err := db.Exec(query, contact_id, user_id, comment)
	fmt.Println(result)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func getContactComments(contact *Contact) ([]*Comment, error) {
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	contact_id, _ := contact.Id.Value()
	sql_query := fmt.Sprintf(
		`
		SELECT comments.id as id, comments.comment as comment
		FROM comments
		WHERE comments.contact_id = %d;
		`, contact_id)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := new(Comment)
		err := rows.Scan(&comment.Id, &comment.Comment)
		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	//for _, contact := range contacts {
	//	fmt.Println(contact)
	//}

	return comments, nil
}

func GetContactCommentsPaginated(contact *Contact, current_page int) ([]*Comment, error) {
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	contact_id, _ := contact.Id.Value()
	offset_by_current_page := (current_page - 1) * COMMENTS_PAGE_SIZE
	sql_query := fmt.Sprintf(
		`
		SELECT comments.id as id, comments.comment as comment
		FROM comments
		WHERE comments.contact_id = %d
		LIMIT %d OFFSET %d;
		`, contact_id, COMMENTS_PAGE_SIZE, offset_by_current_page)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := new(Comment)
		err := rows.Scan(&comment.Id, &comment.Comment)
		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	//for _, contact := range contacts {
	//	fmt.Println(contact)
	//}

	return comments, nil
}
