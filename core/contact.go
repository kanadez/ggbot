package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Contact struct {
	Id       sql.NullInt64
	Contact  sql.NullString
	Rating   sql.NullFloat64
	Comments []*Comment
}

func FindContact(search_query string) ([]*Contact, error) {
	if utf8.RuneCountInString(search_query) > MAX_SEARCH_QUERY_LEN {
		return nil, errors.New("Search query is too big!")
	}

	log.Printf("\n\n\n findContact() search_query: %s \n\n\n", search_query)

	if utf8.RuneCountInString(search_query) == 0 {
		return nil, errors.New("Search query is empty!")
	}

	search_query_lowered := strings.ToLower(search_query)

	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sql_query := fmt.Sprintf(
		`
		SELECT contacts.id as id, contacts.contact as contact, AVG(ratings.rating) as rating
		FROM contacts
		LEFT JOIN ratings ON contacts.id = ratings.contact_id
		WHERE LOWER(contact) LIKE '%%%s%%'
		GROUP BY contacts.id;
		`, search_query_lowered)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	contacts := make([]*Contact, 0)
	for rows.Next() {
		contact := new(Contact)
		err := rows.Scan(&contact.Id, &contact.Contact, &contact.Rating)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(contacts) > 0 {
		for _, contact := range contacts {
			comments, err := getContactComments(contact)

			/*fmt.Println("Comments: ")
			for _, comment := range comments {
				fmt.Println(comment.Comment)
			}*/

			if err != nil {
				return nil, err
			}

			contact.Comments = comments
		}
	}

	/*for _, contact := range contacts {
		fmt.Println(contact)
	}*/

	return contacts, nil
}

func CreateContact(telegram_username string, added_by int) (*Contact, error) {
	if utf8.RuneCountInString(telegram_username) > MAX_CONTACT_LEN {
		return nil, errors.New("New contact name is too big!")
	}

	if utf8.RuneCountInString(telegram_username) == 0 {
		return nil, errors.New("New contact name is empty!")
	}

	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	lastInsertId := 0
	query := `INSERT INTO contacts(contact, added_by) VALUES($1, $2) RETURNING id;`
	result := db.QueryRow(query, telegram_username, strconv.Itoa(added_by))
	err = result.Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}

	fmt.Println("lastInsertId")
	fmt.Println(lastInsertId)

	contact, err := FindContactById(lastInsertId)
	if err != nil {
		return nil, err
	}

	return contact, nil
}

func FindContactById(contact_id int) (*Contact, error) {
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sql_query := fmt.Sprintf(
		`
		SELECT contacts.id as id, contacts.contact as contact, AVG(ratings.rating) as rating
		FROM contacts
		LEFT JOIN ratings ON contacts.id = ratings.contact_id
		WHERE contacts.id = %d
		GROUP BY contacts.id;
		`, contact_id)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	contacts := make([]*Contact, 0)
	for rows.Next() {
		contact := new(Contact)
		err := rows.Scan(&contact.Id, &contact.Contact, &contact.Rating)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	//fmt.Println(contacts[0])

	if len(contacts) == 0 {
		return nil, nil
	}

	return contacts[0], nil
}
