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

func findContact(search_query string) ([]*Contact, error) {

	//log.Printf("findContact() search_query: %s", search_query)

	if utf8.RuneCountInString(search_query) == 0 {
		return nil, errors.New("search_query length is equals 0!")
	}

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sql_query := fmt.Sprintf(
		`
		SELECT contacts.id as id, contacts.contact as contact, AVG(ratings.rating) as rating
		FROM contacts
		LEFT JOIN ratings ON contacts.id = ratings.contact_id
		WHERE contact LIKE '%%%s%%'
		GROUP BY contacts.id;
		`, search_query)
	rows, err := db.Query(sql_query)
	if err != nil {
		log.Fatal(err)
	}

	contacts := make([]*Contact, 0)
	for rows.Next() {
		contact := new(Contact)
		err := rows.Scan(&contact.id, &contact.contact, &contact.rating)
		if err != nil {
			panic(err)
		}

		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	if len(contacts) > 0 {
		for _, contact := range contacts {
			comments, err := getContactComments(contact)

			/*fmt.Println("Comments: ")
			for _, comment := range comments {
				fmt.Println(comment.comment)
			}*/

			if err != nil {
				panic(err)
			}

			contact.comments = comments
		}
	}

	/*for _, contact := range contacts {
		fmt.Println(contact)
	}*/

	return contacts, nil
}

func findContactById(contact_id int) (*Contact, error) {

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	contacts := make([]*Contact, 0)
	for rows.Next() {
		contact := new(Contact)
		err := rows.Scan(&contact.id, &contact.contact, &contact.rating)
		if err != nil {
			panic(err)
		}

		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		panic(err)
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
		log.Fatal(err)
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
		log.Fatal(err)
	}

	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := new(Comment)
		err := rows.Scan(&comment.id, &comment.comment)
		if err != nil {
			panic(err)
		}

		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	//for _, contact := range contacts {
	//	fmt.Println(contact)
	//}

	return comments, nil
}

func getContactCommentsPaginated(contact *Contact, current_page int) ([]*Comment, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := new(Comment)
		err := rows.Scan(&comment.id, &comment.comment)
		if err != nil {
			panic(err)
		}

		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	//for _, contact := range contacts {
	//	fmt.Println(contact)
	//}

	return comments, nil
}

// main.go ===========================================

func main() {

	//Вызываем бота
	//tgBotForText()
	//testDbInsert()
	//tgBotForInlineKb()
	//tgBotForKb()
	//tgSimpleReply()
	//tgInlnineQuery()

	tgSimpleReply()
	//tgSimpleReplyWithUpdates()

	/*response, err := getAllUsers()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)*/
}

// получение отправленного текста и просто дублирование его в ответе
func tgSimpleReply() {
	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	//https://metateg.site/sandbox/go/ggbot/main.go - это видимо надо поменять на НЕ файл, и далее запустить на этот маршрут сервер ниже как-то
	//здесь видимо надо использовать NewWebhookWithCert и свой серт
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
	updates := bot.ListenForWebhook("/ggbot") // здесь непонятно, надо ли менять на свой маршрут. подробнее смотреть пример https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api#NewWebhook

	// здесь мы запускаем веб-сервер, слушаюший запросы от телеграма. видимо, должен быть маршрут, который я передам в тг в качестве вебхука
	go http.ListenAndServeTLS(":88", "ssl.metateg.site.pem", "ssl.metateg.site.key", nil)

	for update := range updates {
		if update.CallbackQuery != nil {
			callback_with_arguments := update.CallbackQuery.Data
			splitted_callback_with_arguments := strings.Split(callback_with_arguments, " ")
			callback := splitted_callback_with_arguments[0]
			fmt.Println(callback_with_arguments)

			switch callback {
			case "/show_more_comments":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				total_pages, err := strconv.Atoi(splitted_callback_with_arguments[2])
				current_page, err := strconv.Atoi(splitted_callback_with_arguments[3])
				if err != nil {
					log.Fatal(err)
				}

				contact, err := findContactById(contact_id)
				if err != nil {
					log.Fatal(err)
				}

				comments, err := getContactCommentsPaginated(contact, current_page)
				if err != nil {
					log.Fatal(err)
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
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg_text)
					msg.ParseMode = "markdown"

					numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							// /show_more_comments <contact_id> <total_pages> <current_page>
							tgbotapi.NewInlineKeyboardButtonData("Показать больше отзывов", fmt.Sprintf("/show_more_comments %d %d %d", contact_id, int(total_pages), int(current_page)+1)),
							tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", "/leave_feedback"),
						),
					)
					msg.ReplyMarkup = numericKeyboard

					bot.Send(msg)
				} else {
					msg_text := "На этот контакт отзывов больше нет."
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msg_text)
					msg.ParseMode = "markdown"

					numericKeyboardIfNoFeedback := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", "/leave_feedback"),
						),
					)
					msg.ReplyMarkup = numericKeyboardIfNoFeedback

					bot.Send(msg)
				}
			}

			continue
		}

		if update.Message == nil {
			continue
		}

		log.Print("===========================================================")
		//log.Printf("update.Message.From.UserName: [%s], update.Message.Text: %s", update.Message.From.UserName, update.Message.Text)
		log.Printf("update: %+v", update)
		log.Printf("update.Message: %+v", update.Message)
		log.Printf("update.Message.Chat: %+v", update.Message.Chat)

		//log.Print("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		search_query := update.Message.Text
		if update.Message.ForwardFrom != nil {
			log.Print("update.Message.ForwardFrom != nil")

			if len(update.Message.ForwardFrom.UserName) > 0 {
				//log.Printf("update.Message.ForwardFrom.UserName: %s\n", update.Message.ForwardFrom.UserName)
				search_query = update.Message.ForwardFrom.UserName
			} else {
				//log.Printf("update.Message.ForwardFrom: %s\n", update.Message.ForwardFrom)
				search_query = fmt.Sprintf("%v", update.Message.ForwardFrom)
			}
		} else if update.Message.ForwardSenderName != "" {
			//log.Print("update.Message.ForwardSenderName != ''")
			//log.Printf("update.Message.ForwardSenderName: %s\n", update.Message.ForwardSenderName)
			search_query = update.Message.ForwardSenderName
		} else {
			//log.Print("no ForwardFrom, no ForwardSenderName")
			//fmt.Printf("%+v\n", (update.Message))
		}

		contacts, err := findContact(search_query)
		if err != nil {
			log.Panic(err)
		}

		if len(contacts) == 0 {
			msg_text := "🤷‍ *Ничего не найдено* \n☝️ Попробуйте ввести вручную имя, ник или номер телефона."
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = "markdown"

			bot.Send(msg)
		}

		for _, contact := range contacts {
			contact_id, _ := contact.id.Value()
			contact_contact, _ := contact.contact.Value()
			contact_rating, _ := contact.rating.Value()

			contact_contact_formatted := fmt.Sprintf("🧍 *Контакт*: %s", contact_contact)
			contact_rating_formatted := "⭐ Оценок пока нет. Это может быть хорошим знаком 😊"
			contact_comments_formatted := ""

			if contact_rating != nil {
				contact_rating_formatted = fmt.Sprintf("⭐ *Оценка*: %.1f", contact_rating)
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

			//fmt.Printf("id: %d\n", contact_id)
			//fmt.Printf("contact: %s\n", contact_contact)
			//fmt.Printf("rating: %d\n", contact_rating)

			msg_text := fmt.Sprintf("%s\n%s\n%s", contact_contact_formatted, contact_rating_formatted, contact_comments_formatted)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msg_text)
			//msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = "markdown"

			var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", "/leave_feedback"),
					tgbotapi.NewInlineKeyboardButtonData("Поставить оценку", "/rate"),
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
						tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", "/leave_feedback"),
						tgbotapi.NewInlineKeyboardButtonData("Поставить оценку", "/rate"),
					),
				)
			}
			msg.ReplyMarkup = numericKeyboard

			bot.Send(msg)
		}
	}
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
