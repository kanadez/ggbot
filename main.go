// main
package main

import (
	"errors"
	//"time"
	"database/sql"
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	//"log"
	"net/http"
	//"net/url"
	"os"
	//"reflect"

	//"strings"
	//"time"

	"github.com/Syfaro/telegram-bot-api"
	_ "github.com/lib/pq"

	//"gopkg.in/go-telegram-bot-api/telegram-bot-api.v2"

	//db.go----------

	//"database/sql"
	//"fmt"
	//"os"
	"strings"
	//_ "github.com/lib/pq"
	"log"
	"math"
	"strconv"
	"unicode/utf8"
)

const COMMENTS_PAGE_SIZE = 3
const MAX_COMMENT_LEN = 1000
const MAX_SEARCH_QUERY_LEN = 255
const MAX_CONTACT_LEN = 255
const MAX_MESSAGE_LEN = 1000

// db.go =========================================

var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var user = os.Getenv("USER")
var password = os.Getenv("PASSWORD")
var dbname = os.Getenv("DBNAME")
var sslmode = os.Getenv("SSLMODE")

var dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

type Comment struct {
	id      sql.NullInt64
	comment sql.NullString
}

type Contact struct {
	id       sql.NullInt64
	contact  sql.NullString
	rating   sql.NullFloat64
	comments []*Comment
}

type User struct {
	id            sql.NullInt64
	telegram_id   sql.NullInt64
	telegram_name sql.NullString
}

type MessagesHistory struct {
	id                    sql.NullInt64
	user_id               sql.NullInt64
	telegram_chat_id      sql.NullString
	telegram_message_id   sql.NullInt64
	telegram_message_text sql.NullString
}

func findContact(search_query string) ([]*Contact, error) {
	if utf8.RuneCountInString(search_query) > MAX_SEARCH_QUERY_LEN {
		return nil, errors.New("Search query is too big!")
	}

	log.Printf("\n\n\n findContact() search_query: %s \n\n\n", search_query)

	if utf8.RuneCountInString(search_query) == 0 {
		return nil, errors.New("Search query is empty!")
	}

	search_query_lowered := strings.ToLower(search_query)

	db, err := sql.Open("postgres", dbInfo)
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
		err := rows.Scan(&contact.id, &contact.contact, &contact.rating)
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
				fmt.Println(comment.comment)
			}*/

			if err != nil {
				return nil, err
			}

			contact.comments = comments
		}
	}

	/*for _, contact := range contacts {
		fmt.Println(contact)
	}*/

	return contacts, nil
}

func createContact(telegram_username string, added_by int) (*Contact, error) {
	if utf8.RuneCountInString(telegram_username) > MAX_CONTACT_LEN {
		return nil, errors.New("New contact name is too big!")
	}

	if utf8.RuneCountInString(telegram_username) == 0 {
		return nil, errors.New("New contact name is empty!")
	}

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
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

	contact, err := findContactById(lastInsertId)
	if err != nil {
		return nil, err
	}

	return contact, nil
}

func findContactById(contact_id int) (*Contact, error) {
	db, err := sql.Open("postgres", dbInfo)
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
		err := rows.Scan(&contact.id, &contact.contact, &contact.rating)
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

func getContactComments(contact *Contact) ([]*Comment, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	contact_id, _ := contact.id.Value()
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
		err := rows.Scan(&comment.id, &comment.comment)
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

func createComment(contact_id int, user_id int, comment string) error {
	if utf8.RuneCountInString(comment) > MAX_COMMENT_LEN {
		return errors.New("Comment length is more than 0!")
	}

	if utf8.RuneCountInString(comment) == 0 {
		return errors.New("Comment is empty!")
	}

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
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

func createRating(contact_id int, user_id int, rating int) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
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

func getContactCommentsPaginated(contact *Contact, current_page int) ([]*Comment, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	contact_id, _ := contact.id.Value()
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
		err := rows.Scan(&comment.id, &comment.comment)
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

func saveMessageToHistory(chat_id int64, telegram_id int, telegram_name string, message_id int, message_text string) error {
	if utf8.RuneCountInString(message_text) > MAX_MESSAGE_LEN {
		return errors.New("Message is too big!")
	}

	if utf8.RuneCountInString(message_text) == 0 {
		return errors.New("Message is empty!")
	}

	log.Printf("saveMessageToHistory: %d, %d, %s, %d, %s", chat_id, telegram_id, telegram_name, message_id, message_text)
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// здесь надо взять из базы users объект текущего телеграм-юзера, либо создать его, если в базе нет
	user, err := getUserByTelegramId(telegram_id)
	if err != nil {
		fmt.Println("getUserByTelegramId() in saveMessageToHistory() err")
		return err
	}
	if user == nil {
		user, err = createUser(telegram_id, telegram_name)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	user_id, _ := user.id.Value()

	query := `INSERT INTO messages_history(user_id, telegram_chat_id, telegram_message_id, telegram_message_text) VALUES($1, $2, $3, $4);`
	result, err := db.Exec(query, fmt.Sprintf("%d", user_id), chat_id, message_id, message_text)
	fmt.Println(result)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func getMessagesHistoryByChatId(telegram_chat_id int64) ([]*MessagesHistory, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sql_query := fmt.Sprintf(
		`
		SELECT id, user_id, telegram_chat_id, telegram_message_id, telegram_message_text
		FROM messages_history
		WHERE telegram_chat_id = %d;
		`, telegram_chat_id)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	messages_history := make([]*MessagesHistory, 0)
	for rows.Next() {
		messages_history_record := new(MessagesHistory)
		err := rows.Scan(&messages_history_record.id, &messages_history_record.user_id, &messages_history_record.telegram_chat_id, &messages_history_record.telegram_message_id, &messages_history_record.telegram_message_text)
		if err != nil {
			return nil, err
		}

		messages_history = append(messages_history, messages_history_record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	//for _, message := range messages_history {
	//	fmt.Println(message)
	//}

	return messages_history, nil
}

func getPreLastMessageFromHistoryToCheckCallback(telegram_chat_id int64) (*MessagesHistory, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sql_query := fmt.Sprintf(
		`
		SELECT id, user_id, telegram_chat_id, telegram_message_id, telegram_message_text
		FROM messages_history
		WHERE telegram_chat_id = %d
		ORDER BY id desc
		LIMIT 1 OFFSET 2;
		`, telegram_chat_id)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	messages_history_record := new(MessagesHistory)
	for rows.Next() {
		err := rows.Scan(&messages_history_record.id, &messages_history_record.user_id, &messages_history_record.telegram_chat_id, &messages_history_record.telegram_message_id, &messages_history_record.telegram_message_text)
		if err != nil {
			return nil, err
		}
	}

	return messages_history_record, nil
}

func createUser(telegram_id int, telegram_name string) (*User, error) {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
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
	db, err := sql.Open("postgres", dbInfo)
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
	db, err := sql.Open("postgres", dbInfo)
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

// main.go ===========================================

func main() {
	//tgBotForText()
	//testDbInsert()
	//tgBotForInlineKb()
	//tgBotForKb()
	//tgSimpleReply()
	//tgInlnineQuery()
	//tgSimpleReplyWithUpdates()

	tgSimpleReply()
}

//func (logger BotLogger) Println(v string) string {
//	fmt.Println("BotLogger Println'")
//	fmt.Println(v)
//}

func tgSimpleReply() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://metateg.site:88/ggbot"))
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	updates := bot.ListenForWebhook("/ggbot")

	go http.ListenAndServeTLS(":88", "ssl.metateg.site.pem", "ssl.metateg.site.key", nil)

	for update := range updates {

		// Перебираем коллбеки от юзера, т.е. нажатия на кнопки (к примеру)
		if update.CallbackQuery != nil {
			callback_with_arguments := update.CallbackQuery.Data
			splitted_callback_with_arguments := strings.Split(callback_with_arguments, " ")
			callback := splitted_callback_with_arguments[0]
			fmt.Println(callback_with_arguments)
			err := saveMessageToHistory(update.CallbackQuery.Message.Chat.ID, -1, "bot", int(update.CallbackQuery.Message.MessageID+1), callback_with_arguments)
			if err != nil {
				log.Println(err)
			}

			switch callback {
			case "/show_more_comments":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				total_pages, err := strconv.Atoi(splitted_callback_with_arguments[2])
				current_page, err := strconv.Atoi(splitted_callback_with_arguments[3])
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				contact, err := findContactById(contact_id)
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				comments, err := getContactCommentsPaginated(contact, current_page)
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				contact_comments_formatted := ""
				if len(comments) > 0 {
					var sb strings.Builder
					for index, comment := range comments {
						if index < 3 {
							comment_value, _ := comment.comment.Value()
							sb.WriteString(fmt.Sprintf("%s", comment_value))
							if index < 2 {
								sb.WriteString("\n---\n")
							}
						}
					}

					contact_comments_formatted = sb.String()
					msg_text := fmt.Sprintf("%s", contact_comments_formatted)
					numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							// /show_more_comments <contact_id> <total_pages> <current_page>
							tgbotapi.NewInlineKeyboardButtonData("Показать больше отзывов", fmt.Sprintf("/show_more_comments %d %d %d", contact_id, int(total_pages), int(current_page)+1)),
							tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/leave_feedback %d", contact_id)),
						),
					)
					sendBotMessageWithKeyboard(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboard)
				} else {
					msg_text := "На этот контакт отзывов больше нет."
					numericKeyboardIfNoFeedback := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/leave_feedback %d", contact_id)),
						),
					)
					sendBotMessageWithKeyboard(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboardIfNoFeedback)
				}

			case "/leave_feedback":
				msg_text := "Напишите и отправьте текст отзыва:"
				sendBotMessageWithForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID)

			case "/create_and_leave_feedback":
				msg_text := "Напишите и отправьте текст отзыва:"
				sendBotMessageWithForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID)

			case "/rate":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				msg_text := "Выберите оценку:"
				numericKeyboardIfNoFeedback := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("1⭐", fmt.Sprintf("/rate_score %d 1", contact_id)),
						tgbotapi.NewInlineKeyboardButtonData("2⭐", fmt.Sprintf("/rate_score %d 2", contact_id)),
						tgbotapi.NewInlineKeyboardButtonData("3⭐", fmt.Sprintf("/rate_score %d 3", contact_id)),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("4⭐", fmt.Sprintf("/rate_score %d 4", contact_id)),
						tgbotapi.NewInlineKeyboardButtonData("5⭐", fmt.Sprintf("/rate_score %d 5", contact_id)),
					),
				)
				sendBotMessageWithKeyboardAndForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboardIfNoFeedback)

			case "/rate_score":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}
				score, err := strconv.Atoi(splitted_callback_with_arguments[2])
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}
				err = createRating(contact_id, int(update.CallbackQuery.Message.Chat.ID), score)
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				msg_text := "Ваша оценка сохранена, спасибо! Теперь ее увидят все, кто будет просматривать этого пользователя/телефон."
				sendBotMessage(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID)

			case "/create_and_rate":
				new_contact_name := splitted_callback_with_arguments[1]
				contact, err := createContact(new_contact_name, update.CallbackQuery.Message.From.ID)
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}
				contact_id, _ := contact.id.Value()
				contact_id_asserted, ok := contact_id.(int64) // type assertion
				if ok == false {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, errors.New("Type of contact_id is not asserted"))
					continue
				}
				contact_id_as_int := int(contact_id_asserted)
				if err != nil {
					sendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				msg_text := "Выберите оценку:"
				numericKeyboardIfNoFeedback := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("1⭐", fmt.Sprintf("/rate_score %d 1", contact_id_as_int)),
						tgbotapi.NewInlineKeyboardButtonData("2⭐", fmt.Sprintf("/rate_score %d 2", contact_id_as_int)),
						tgbotapi.NewInlineKeyboardButtonData("3⭐", fmt.Sprintf("/rate_score %d 3", contact_id_as_int)),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("4⭐", fmt.Sprintf("/rate_score %d 4", contact_id_as_int)),
						tgbotapi.NewInlineKeyboardButtonData("5⭐", fmt.Sprintf("/rate_score %d 5", contact_id_as_int)),
					),
				)
				sendBotMessageWithKeyboardAndForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboardIfNoFeedback)

				continue
			}
		}

		if update.Message == nil {
			continue
		}

		log.Printf("Message not nil, TEXT: %+s \n", update.Message.Text)

		// Перебираем сообщения от юзера, в том числе команды типа /start
		if update.Message.Text == "/start" {
			msg_text := "📩 Перешлите сюда сообщение от пользователя, которого нужно проверить. \n✏️ Если переслать нечего, прямо вручную введите имя или номер телефона пользователя."
			sendBotMessage(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID)

			continue
		}

		log.Print("===========================================================")
		log.Printf("update: %+v \n\n\n", update)
		log.Printf("update.Message: %+v \n\n\n", update.Message)
		log.Printf("update.Message.From: %+v \n\n\n", update.Message.From)
		log.Printf("update.Message.From.ID: %d \n\n\n", update.Message.From.ID)
		log.Printf("update.Message.ForwardFrom: %+v \n\n\n", update.Message.ForwardFrom)
		log.Printf("update.Message.ForwardSenderName: %+s \n\n\n", update.Message.ForwardSenderName)
		log.Printf("update.Message.Chat: %+v \n\n\n", update.Message.Chat)

		// Берем из сообщения имя пользователя (в любом формате)
		message_from_username := ""
		if len(update.Message.From.UserName) > 0 {
			message_from_username = update.Message.From.UserName
		} else {
			message_from_username := update.Message.From.FirstName
			if update.Message.From.LastName != "" {
				message_from_username += " " + update.Message.From.LastName
			}
		}

		// Сохраняем сообщение в историю
		err := saveMessageToHistory(update.Message.Chat.ID, update.Message.From.ID, message_from_username, update.Message.MessageID, update.Message.Text)
		if err != nil {
			log.Println(err)
		}

		// Берем из истории предыдущее сообщение для поиска возможного коллбака
		log.Println("getPreLastMessageFromHistoryToCheckCallback()")
		message, err := getPreLastMessageFromHistoryToCheckCallback(update.Message.Chat.ID)
		if err != nil {
			sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
			continue
		}

		// Перебор коллбаков, если есть
		message_text, _ := message.telegram_message_text.Value()
		message_text_as_string := fmt.Sprintf("%s", message_text)
		first_character := message_text_as_string[0:1]
		if first_character == "/" {
			callback_from_history_with_arguments := message_text_as_string
			splitted_callback_from_history_with_arguments := strings.Split(callback_from_history_with_arguments, " ")
			callback_from_history := splitted_callback_from_history_with_arguments[0]
			callback_from_history_argument := ""
			if len(splitted_callback_from_history_with_arguments) > 1 {
				callback_from_history_argument = splitted_callback_from_history_with_arguments[1]
			}
			switch callback_from_history {
			case "/create_and_leave_feedback":
				if callback_from_history_argument == "" {
					continue
				}

				contact, err := createContact(callback_from_history_argument, update.Message.From.ID)
				if err != nil {
					sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}
				contact_id, _ := contact.id.Value()
				contact_id_asserted, ok := contact_id.(int64) // type assertion
				if ok == false {
					sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, errors.New("Type of contact_id is not asserted"))
					continue
				}
				contact_id_as_int := int(contact_id_asserted)
				err = createComment(contact_id_as_int, update.Message.From.ID, update.Message.Text)
				if err != nil {
					sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}

				msg_text := "Ваш отзыв сохранен, спасибо! Теперь его увидят все, кто будет просматривать это имя/номер телефона."
				sendBotMessage(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID)

				continue
			case "/leave_feedback":
				if callback_from_history_argument == "" {
					continue
				}

				contact_id := callback_from_history_argument
				contact_id_numeric, err := strconv.Atoi(contact_id)
				if err != nil {
					sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}
				err = createComment(contact_id_numeric, update.Message.From.ID, update.Message.Text)
				if err != nil {
					sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}

				msg_text := "Ваш отзыв сохранен, спасибо! Теперь его увидят все, кто будет просматривать этого пользователя."
				sendBotMessage(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID)

				continue
			}
		}

		// Поиск по сообщению среди контактов
		search_query := update.Message.Text
		if update.Message.ForwardFrom != nil {
			log.Print("update.Message.ForwardFrom != nil")

			if len(update.Message.ForwardFrom.UserName) > 0 {
				search_query = update.Message.ForwardFrom.UserName
			} else {
				name := update.Message.ForwardFrom.FirstName
				if update.Message.ForwardFrom.LastName != "" {
					name += " " + update.Message.ForwardFrom.LastName
				}

				search_query = fmt.Sprintf("%v", name)
			}
		} else if update.Message.ForwardSenderName != "" {
			search_query = update.Message.ForwardSenderName
		}

		contacts, err := findContact(search_query)
		if err != nil {
			sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
			continue
		}

		// Если ничего не найдено по сообщению
		if len(contacts) == 0 {
			search_query_escaped := escapeStringForMarkdown(search_query)
			msg_text := fmt.Sprintf("🤷‍ *Не нашли такого имени или номера телефона...* \n📩 Попробуйте переслать сюда сообщение от пользователя, которого нужно проверить. \n✏️ Если переслать нечего, оставьте отзыв или оценку прямо на %s:", search_query_escaped)
			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/create_and_leave_feedback %s", search_query)), // оставить отзыв на текстовый ввод
					tgbotapi.NewInlineKeyboardButtonData("Поставить оценку", fmt.Sprintf("/create_and_rate %s", search_query)),
				),
			)
			if update.Message.ForwardFrom != nil || update.Message.ForwardSenderName != "" {
				contact_name := ""
				if update.Message.ForwardFrom != nil {
					if len(update.Message.ForwardFrom.UserName) > 0 {
						contact_name = update.Message.ForwardFrom.UserName
					} else {
						contact_name = update.Message.ForwardFrom.FirstName
						if update.Message.ForwardFrom.LastName != "" {
							contact_name += " " + update.Message.ForwardFrom.LastName
						}
					}
				} else if update.Message.ForwardSenderName != "" {
					contact_name = update.Message.ForwardSenderName
				}
				contact, err := createContact(contact_name, update.Message.From.ID)
				if err != nil {
					sendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}
				contact_id, _ := contact.id.Value()
				contact_contact, _ := contact.contact.Value()
				contact_contact_escaped := escapeStringForMarkdown(fmt.Sprintf("%s", contact_contact))
				contact_contact_formatted := fmt.Sprintf("🧍 *Имя или телефон*: %s", contact_contact_escaped)
				contact_rating_formatted := "⭐ Оценок пока нет. Вообще, это хорошо - значит, никто не жаловался 😊 \n✏️ Если есть что сказать, оставьте первый отзыв или оценку:"
				msg_text = fmt.Sprintf("%s\n%s", contact_contact_formatted, contact_rating_formatted)

				numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/leave_feedback %d", contact_id)),
						tgbotapi.NewInlineKeyboardButtonData("Поставить оценку", fmt.Sprintf("/rate %d", contact_id)),
					),
				)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
			if update.Message.ForwardFrom == nil && update.Message.ForwardSenderName == "" {
				msg.ReplyToMessageID = update.Message.MessageID
			}
			msg.ReplyMarkup = numericKeyboard
			msg.ParseMode = "markdown"

			bot.Send(msg)
			err := saveMessageToHistory(update.Message.Chat.ID, -1, "bot", int(update.Message.MessageID+1), msg_text)
			if err != nil {
				log.Println(err)
			}
		}

		// Если найдены контакты по сообщению
		for _, contact := range contacts {
			contact_id, _ := contact.id.Value()
			contact_contact, _ := contact.contact.Value()
			contact_contact_escaped := escapeStringForMarkdown(fmt.Sprintf("%s", contact_contact))
			contact_rating, _ := contact.rating.Value()
			contact_contact_formatted := fmt.Sprintf("🧍 *Имя или телефон*: %s", contact_contact_escaped)
			contact_rating_formatted := "⭐ Оценок пока нет. Вообще, это хорошо - значит, никто не жаловался 😊\n✏️ Если есть что сказать, оставьте первый отзыв или оценку:"
			contact_comments_formatted := ""

			if contact_rating != nil {
				contact_rating_formatted = fmt.Sprintf("⭐ *Оценка*: %.1f", contact_rating)
			}
			if contact_rating == nil && len(contact.comments) > 0 {
				contact_rating_formatted = "⭐ Оценок пока нет. Вообще, это хорошо - значит, никто не жаловался 😊\n✏️ Если есть что сказать, оставьте отзыв или оценку:"
			}

			if len(contact.comments) > 0 {
				var sb strings.Builder
				sb.WriteString("\n📢 *Отзывы*:\n")
				for index, comment := range contact.comments {
					if index < 3 {
						comment_value, _ := comment.comment.Value()
						sb.WriteString(fmt.Sprintf("%s", comment_value))
						if index < 2 {
							sb.WriteString("\n---\n")
						}
					}
				}

				contact_comments_formatted = sb.String()
			}

			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/leave_feedback %d", contact_id)),
					tgbotapi.NewInlineKeyboardButtonData("Поставить оценку", fmt.Sprintf("/rate %d", contact_id)),
				),
			)
			if len(contact.comments) > 3 {
				comments_len := len(contact.comments)
				comments_pages_count := math.Ceil(float64(comments_len) / float64(COMMENTS_PAGE_SIZE))

				numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						// /show_more_comments <contact_id> <total_pages> <current_page>
						tgbotapi.NewInlineKeyboardButtonData("Показать больше отзывов", fmt.Sprintf("/show_more_comments %d %d 2", contact_id, int(comments_pages_count))),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/leave_feedback %d", contact_id)),
						tgbotapi.NewInlineKeyboardButtonData("Поставить оценку", fmt.Sprintf("/rate %d", contact_id)),
					),
				)
			}
			msg_text := fmt.Sprintf("%s\n%s\n%s", contact_contact_formatted, contact_rating_formatted, contact_comments_formatted)
			sendBotMessageWithKeyboard(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID, numericKeyboard)
		}
	}
}

func escapeStringForMarkdown(str string) string {
	str_escaped := strings.Replace(str, "_", "\\_", -1)
	str_escaped = strings.Replace(str_escaped, "*", "\\*", -1)
	str_escaped = strings.Replace(str_escaped, "[", "\\[", -1)
	str_escaped = strings.Replace(str_escaped, "`", "\\`", -1)

	return str_escaped
}

func tgSimpleReplyWithUpdates() {
	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.RemoveWebhook()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//log.Printf("%+v", update.Message)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}

func sendBotMessageSendError(bot *tgbotapi.BotAPI, chat_id int64, message_id int, err error) {
	log.Println(err)
	msg_text := fmt.Sprintf("Ошибка: %s,\n Мы уже в курсе, исправляем. Извините за неудобства.", err)
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	_, send_err := bot.Send(msg)
	if send_err != nil {
		log.Fatal(send_err)
	}

	save_err := saveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func sendBotMessage(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := saveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func sendBotMessageWithKeyboard(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := saveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func sendBotMessageWithForceReply(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"

	var force_reply = new(tgbotapi.ForceReply)
	force_reply.ForceReply = true
	msg.ForceReply = force_reply
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := saveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func sendBotMessageWithKeyboardAndForceReply(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = keyboard

	var force_reply = new(tgbotapi.ForceReply)
	force_reply.ForceReply = true
	msg.ForceReply = force_reply

	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := saveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}
