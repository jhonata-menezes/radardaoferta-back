package sopromocao

import (
	"log"
	"net/url"
	"time"

	"github.com/tucnak/telebot"
)

func ShowTelegram(urls chan string) {
	bot, err := telebot.NewBot("233668925:AAFFwfhIY292fmTxKZgFjcFDsCF9FpwZVQ0")
	if err != nil {
		log.Fatalln(err)
	}

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, 1*time.Second)
	var msg string
	for message := range messages {
		log.Println("new message ", message.Sender.FirstName, "ID", message.Sender.ID)
		validator, err := url.Parse(message.Text)
		if err != nil {
			panic(err)
		}
		if validator.Host == "" {
			msg = message.Sender.FirstName + " desculpa, mas nao foi possivel identifica link do produto, para que possa ser identificado envie apenas o link."
		} else {
			msg = message.Sender.FirstName + " muito obrigado por compartilhar, irei processar e publicar no site http://www.radardaoferta.com.br/"
			urls <- message.Text
		}
		bot.SendMessage(message.Chat, msg, nil)
	}
}
