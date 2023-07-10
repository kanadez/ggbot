package core

import (
	"fmt"
	"os"
)

var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var user = os.Getenv("USER")
var password = os.Getenv("GG_CHECK_BOT_PASSWORD")
var dbname = os.Getenv("GG_CHECK_BOT_DBNAME")
var sslmode = os.Getenv("SSLMODE")

func GetDBcreds() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
}
