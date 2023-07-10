package core

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"unicode/utf8"
)

type MessagesHistory struct {
	Id                  sql.NullInt64
	UserId              sql.NullInt64
	TelegramChatId      sql.NullString
	TelegramMessageId   sql.NullInt64
	TelegramMessageText sql.NullString
}

func SaveMessageToHistory(chat_id int64, telegram_id int, telegram_name string, message_id int, message_text string) error {
	if utf8.RuneCountInString(message_text) > MAX_MESSAGE_LEN {
		return errors.New("Message is too big!")
	}

	if utf8.RuneCountInString(message_text) == 0 {
		return errors.New("Message is empty!")
	}

	log.Printf("saveMessageToHistory: %d, %d, %s, %d, %s", chat_id, telegram_id, telegram_name, message_id, message_text)
	//Подключаемся к БД
	db, err := sql.Open("postgres", GetDBcreds())
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

func GetMessagesHistoryByChatId(TelegramChatId int64) ([]*MessagesHistory, error) {
	db, err := sql.Open("postgres", GetDBcreds())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sql_query := fmt.Sprintf(
		`
		SELECT id, user_id, telegram_chat_id, telegram_message_id, telegram_message_text
		FROM messages_history
		WHERE telegram_chat_id = %d;
		`, TelegramChatId)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	messages_history := make([]*MessagesHistory, 0)
	for rows.Next() {
		messages_history_record := new(MessagesHistory)
		err := rows.Scan(&messages_history_record.Id, &messages_history_record.UserId, &messages_history_record.TelegramChatId, &messages_history_record.TelegramMessageId, &messages_history_record.TelegramMessageText)
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

func GetPreLastMessageFromHistoryToCheckCallback(TelegramChatId int64) (*MessagesHistory, error) {
	db, err := sql.Open("postgres", GetDBcreds())
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
		`, TelegramChatId)
	rows, err := db.Query(sql_query)
	if err != nil {
		return nil, err
	}

	messages_history_record := new(MessagesHistory)
	for rows.Next() {
		err := rows.Scan(&messages_history_record.Id, &messages_history_record.UserId, &messages_history_record.TelegramChatId, &messages_history_record.TelegramMessageId, &messages_history_record.TelegramMessageText)
		if err != nil {
			return nil, err
		}
	}

	return messages_history_record, nil
}

func SendBotMessageSendError(bot *tgbotapi.BotAPI, chat_id int64, message_id int, err error) {
	log.Println(err)
	msg_text := fmt.Sprintf("Ошибка: %s,\n Мы уже в курсе, исправляем. Извините за неудобства.", err)
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	_, send_err := bot.Send(msg)
	if send_err != nil {
		log.Fatal(send_err)
	}

	save_err := SaveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func SendBotMessage(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := SaveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func SendBotMessageWithKeyboard(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := SaveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func SendBotMessageWithForceReply(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int) {
	msg := tgbotapi.NewMessage(chat_id, msg_text)
	msg.ParseMode = "markdown"

	var force_reply = new(tgbotapi.ForceReply)
	force_reply.ForceReply = true
	msg.ForceReply = force_reply
	_, err := bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	save_err := SaveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}

func SendBotMessageWithKeyboardAndForceReply(bot *tgbotapi.BotAPI, chat_id int64, msg_text string, message_id int, keyboard tgbotapi.InlineKeyboardMarkup) {
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

	save_err := SaveMessageToHistory(chat_id, -1, "bot", int(message_id+1), msg_text)
	if save_err != nil {
		log.Println(save_err)
	}
}
