package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

var (
	bot      *tgbotapi.BotAPI
	botToken = os.Getenv("GG_BOT_TOKEN")
)

func initTelegram() {
	var err error

	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://metateg.online:88/ggbot"))
	if err != nil {
		log.Println(err)
	}
}

func webhookHandler(c *gin.Context) {
	defer c.Request.Body.Close()

	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
	if err != nil {
		log.Println(err)
		return
	}

	// to monitor changes run: heroku logs --tail
	log.Printf("From: %+v Text: %+v\n", update.Message.From, update.Message.Text)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// gin router
	router := gin.New()
	router.Use(gin.Logger())

	// telegram
	initTelegram()
	router.POST("/ggbot", webhookHandler)

	err := router.Run("https://metateg.online:88")
	if err != nil {
		log.Println(err)
	}
}
