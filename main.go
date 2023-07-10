package main

import (
	"errors"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/kanadez/ggbot/core"
	_ "github.com/lib/pq"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	tgSimpleReply()
}

//func (logger BotLogger) Println(v string) string {
//	fmt.Println("BotLogger Println'")
//	fmt.Println(v)
//}

func tgSimpleReply() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_CHECK_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	/**
	// Updates with Webhook start
	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://metateg.site:88/ggbot-check"))
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
	updates := bot.ListenForWebhook("/ggbot-check")

	go http.ListenAndServeTLS(":88", "ssl.pem", "ssl.key", nil)
	// Updates with Webhook end
	**/

	_, err = bot.RemoveWebhook()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		// Перебираем коллбеки от юзера, т.е. нажатия на кнопки (к примеру)
		if update.CallbackQuery != nil {
			callback_with_arguments := update.CallbackQuery.Data
			splitted_callback_with_arguments := strings.Split(callback_with_arguments, " ")
			callback := splitted_callback_with_arguments[0]
			fmt.Println(callback_with_arguments)
			err := core.SaveMessageToHistory(update.CallbackQuery.Message.Chat.ID, -1, "bot", int(update.CallbackQuery.Message.MessageID+1), callback_with_arguments)
			if err != nil {
				log.Println(err)
			}

			switch callback {
			case "/show_more_comments":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				total_pages, err := strconv.Atoi(splitted_callback_with_arguments[2])
				current_page, err := strconv.Atoi(splitted_callback_with_arguments[3])
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				contact, err := core.FindContactById(contact_id)
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				comments, err := core.GetContactCommentsPaginated(contact, current_page)
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				contact_comments_formatted := ""
				if len(comments) > 0 {
					var sb strings.Builder
					for index, comment := range comments {
						if index < 3 {
							comment_value, _ := comment.Comment.Value()
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
					core.SendBotMessageWithKeyboard(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboard)
				} else {
					msg_text := "На этот контакт отзывов больше нет."
					numericKeyboardIfNoFeedback := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Оставить отзыв", fmt.Sprintf("/leave_feedback %d", contact_id)),
						),
					)
					core.SendBotMessageWithKeyboard(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboardIfNoFeedback)
				}

			case "/leave_feedback":
				msg_text := "Напишите и отправьте текст отзыва:"
				core.SendBotMessageWithForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID)

			case "/create_and_leave_feedback":
				msg_text := "Напишите и отправьте текст отзыва:"
				core.SendBotMessageWithForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID)

			case "/rate":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
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
				core.SendBotMessageWithKeyboardAndForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboardIfNoFeedback)

			case "/rate_score":
				contact_id, err := strconv.Atoi(splitted_callback_with_arguments[1])
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}
				score, err := strconv.Atoi(splitted_callback_with_arguments[2])
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}
				err = core.CreateRating(contact_id, int(update.CallbackQuery.Message.Chat.ID), score)
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}

				msg_text := "Ваша оценка сохранена, спасибо! Теперь ее увидят все, кто будет просматривать этого пользователя/телефон."
				core.SendBotMessage(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID)

			case "/create_and_rate":
				new_contact_name := splitted_callback_with_arguments[1]
				contact, err := core.CreateContact(new_contact_name, update.CallbackQuery.Message.From.ID)
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
					continue
				}
				contact_id, _ := contact.Id.Value()
				contact_id_asserted, ok := contact_id.(int64) // type assertion
				if ok == false {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, errors.New("Type of contact_id is not asserted"))
					continue
				}
				contact_id_as_int := int(contact_id_asserted)
				if err != nil {
					core.SendBotMessageSendError(bot, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, err)
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
				core.SendBotMessageWithKeyboardAndForceReply(bot, update.CallbackQuery.Message.Chat.ID, msg_text, update.CallbackQuery.Message.MessageID, numericKeyboardIfNoFeedback)

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
			core.SendBotMessage(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID)

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
		err := core.SaveMessageToHistory(update.Message.Chat.ID, update.Message.From.ID, message_from_username, update.Message.MessageID, update.Message.Text)
		if err != nil {
			log.Println(err)
		}

		// Берем из истории предыдущее сообщение для поиска возможного коллбака
		log.Println("getPreLastMessageFromHistoryToCheckCallback()")
		message, err := core.GetPreLastMessageFromHistoryToCheckCallback(update.Message.Chat.ID)
		if err != nil {
			core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
			continue
		}

		// Перебор коллбаков, если есть
		message_text, _ := message.TelegramMessageText.Value()
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

				contact, err := core.CreateContact(callback_from_history_argument, update.Message.From.ID)
				if err != nil {
					core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}
				contact_id, _ := contact.Id.Value()
				contact_id_asserted, ok := contact_id.(int64) // type assertion
				if ok == false {
					core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, errors.New("Type of contact_id is not asserted"))
					continue
				}
				contact_id_as_int := int(contact_id_asserted)
				err = core.CreateComment(contact_id_as_int, update.Message.From.ID, update.Message.Text)
				if err != nil {
					core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}

				msg_text := "Ваш отзыв сохранен, спасибо! Теперь его увидят все, кто будет просматривать это имя/номер телефона."
				core.SendBotMessage(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID)

				continue
			case "/leave_feedback":
				if callback_from_history_argument == "" {
					continue
				}

				contact_id := callback_from_history_argument
				contact_id_numeric, err := strconv.Atoi(contact_id)
				if err != nil {
					core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}
				err = core.CreateComment(contact_id_numeric, update.Message.From.ID, update.Message.Text)
				if err != nil {
					core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}

				msg_text := "Ваш отзыв сохранен, спасибо! Теперь его увидят все, кто будет просматривать этого пользователя."
				core.SendBotMessage(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID)

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

		contacts, err := core.FindContact(search_query)
		if err != nil {
			core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
			continue
		}

		// Если ничего не найдено по сообщению
		if len(contacts) == 0 {
			search_query_escaped := core.EscapeStringForMarkdown(search_query)
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
				contact, err := core.CreateContact(contact_name, update.Message.From.ID)
				if err != nil {
					core.SendBotMessageSendError(bot, update.Message.Chat.ID, update.Message.MessageID, err)
					continue
				}
				contact_id, _ := contact.Id.Value()
				contact_contact, _ := contact.Contact.Value()
				contact_contact_escaped := core.EscapeStringForMarkdown(fmt.Sprintf("%s", contact_contact))
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
			err := core.SaveMessageToHistory(update.Message.Chat.ID, -1, "bot", int(update.Message.MessageID+1), msg_text)
			if err != nil {
				log.Println(err)
			}
		}

		// Если найдены контакты по сообщению
		for _, contact := range contacts {
			contact_id, _ := contact.Id.Value()
			contact_contact, _ := contact.Contact.Value()
			contact_contact_escaped := core.EscapeStringForMarkdown(fmt.Sprintf("%s", contact_contact))
			contact_rating, _ := contact.Rating.Value()
			contact_contact_formatted := fmt.Sprintf("🧍 *Имя или телефон*: %s", contact_contact_escaped)
			contact_rating_formatted := "⭐ Оценок пока нет. Вообще, это хорошо - значит, никто не жаловался 😊\n✏️ Если есть что сказать, оставьте первый отзыв или оценку:"
			contact_comments_formatted := ""

			if contact_rating != nil {
				contact_rating_formatted = fmt.Sprintf("⭐ *Оценка*: %.1f", contact_rating)
			}
			if contact_rating == nil && len(contact.Comments) > 0 {
				contact_rating_formatted = "⭐ Оценок пока нет. Вообще, это хорошо - значит, никто не жаловался 😊\n✏️ Если есть что сказать, оставьте отзыв или оценку:"
			}

			if len(contact.Comments) > 0 {
				var sb strings.Builder
				sb.WriteString("\n📢 *Отзывы*:\n")
				for index, comment := range contact.Comments {
					if index < 3 {
						comment_value, _ := comment.Comment.Value()
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
			if len(contact.Comments) > 3 {
				comments_len := len(contact.Comments)
				comments_pages_count := math.Ceil(float64(comments_len) / float64(core.COMMENTS_PAGE_SIZE))

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
			core.SendBotMessageWithKeyboard(bot, update.Message.Chat.ID, msg_text, update.Message.MessageID, numericKeyboard)
		}
	}
}
