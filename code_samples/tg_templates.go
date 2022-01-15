

// получение текста и отправка ответа
func tgBotForText() {

	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))

	if err != nil {
		panic(err)
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//Проверяем что от пользователья пришло именно текстовое сообщение
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			switch update.Message.Text {
			case "/start":
				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a wikipedia bot, i can search information in a wikipedia, send me something what you want find in Wikipedia.")
				bot.Send(msg)
			}
		} else {
			//Отправлем сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Use the words for search.")
			bot.Send(msg)
		}
	}
}

// выдача встроенной (внутри сообщения) клавиатуры и реакция на нажатие ее кнопок
func tgBotForInlineKb() {

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
			tgbotapi.NewInlineKeyboardButtonSwitch("2sw", "open 2"),
			tgbotapi.NewInlineKeyboardButtonData("3", "3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("4", "4"),
			tgbotapi.NewInlineKeyboardButtonData("5", "5"),
			tgbotapi.NewInlineKeyboardButtonData("6", "6"),
		),
	)

	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	fmt.Print(".")
	for update := range updates {
		if update.CallbackQuery != nil {
			fmt.Print(update)

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

			bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
		}
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Text {
			case "open":
				msg.ReplyMarkup = numericKeyboard

			}

			bot.Send(msg)
		}
	}
}

// выдача НЕвстроенной клавиатуры (внизу экрана) и реакция на ее нажатия
func tgBotForKb() {
	var numericKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1"),
			tgbotapi.NewKeyboardButton("2"),
			tgbotapi.NewKeyboardButton("3"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("4"),
			tgbotapi.NewKeyboardButton("5"),
			tgbotapi.NewKeyboardButton("6"),
		),
	)

	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		switch update.Message.Text {
		case "open":
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}

		bot.Send(msg)
	}
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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}

// получение inline-запроса типа @tizer_spb_bot 123 и ответ списком результатов с одним пунктом Echo 123
func tgInlnineQuery() {
	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("GG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.InlineQuery.Query == "" {

		}

		article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, "Echo", update.InlineQuery.Query)
		article.Description = update.InlineQuery.Query

		inlineConf := tgbotapi.InlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{article},
		}

		if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
			log.Println(err)
		}
	}
}

// запрос к базе с результатом одной строкой
func getNumberOfUsers() (int64, error) {

	var count int64

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// команда для запроса с результатом одной строки
	row := db.QueryRow("SELECT COUNT(id) FROM users;")
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// запрос к базе с результатом множеством строк
func getAllUsers() (string, error) {

	var response string

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// команда для запроса с результатом множества строк
	rows, err := db.Query("SELECT * FROM users;")
	if err != nil {
		log.Fatal(err)
	}

	type User struct { // сюда будем класть результаты запроса
		id        int
		timestamp string
		username  string
		chat_id   int
		message   string
		answer    string
	}

	users := make([]*User, 0) // создаем пустой массив с динамическим размером из указателей на структуру User и возвращаем его срез (slice)
	for rows.Next() {
		user := new(User)
		// .Scan кладет результаты запроса только в указатели, поэтому кладем их в указатели на поля структуры User
		// указатель это не переменная, это адрес переменной в памяти. т е мы меняем не переменную напрямую, а по ее адресу в памяти. Это дает возможность
		// , например, не думать в пределах области видимости, а думать в пределах памяти
		err := rows.Scan(&user.id, &user.timestamp, &user.username, &user.chat_id, &user.message, &user.answer)
		if err != nil {
			panic(err)
		}
		//fmt.Println(user)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	for _, user := range users {
		fmt.Println(user.id)
	}

	return response, nil
}

// вставка строки в базу
func testDbInsert() error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаем SQL запрос
	query := `INSERT INTO users(username, chat_id, message, answer) VALUES($1, $2, $3, $4);`

	for i := 0; i <= 10; i++ {
		//Выполняем наш SQL запрос
		result, err := db.Exec(query, `@test`, i, "test-message %d"+strconv.Itoa(i), "test-answer")
		fmt.Println(result)
		fmt.Println(time.Now())

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}