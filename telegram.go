package sopromocao

import (
	"log"
	"net/url"
	"time"

	"github.com/tucnak/telebot"
)

//dev
//prod 233668925:AAFFwfhIY292fmTxKZgFjcFDsCF9FpwZVQ0
func ShowTelegram(urls chan string) {
	bot, err := telebot.NewBot("278478343:AAGXj08TDDnM6_fv7yiZNK78ZkE-UH9Zges")
	if err != nil {
		log.Fatalln(err)
	}

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, 1*time.Second)
	var msg string
	for message := range messages {
		log.Println("new message ", message.Sender.FirstName, "ID", message.Sender.ID)
		validator, err := url.Parse(CleanUrl(message.Text))
		if err != nil {
			panic(err)
		}
		if validator.Host == "" {
			msg = message.Sender.FirstName + " desculpa, mas nao foi possivel identifica link do produto, para que possa ser identificado envie apenas o link."
		} else {
			msg = message.Sender.FirstName + " muito obrigado por compartilhar, irei processar e posteriormente publicar no site http://www.radardaoferta.com.br/"
			urls <- message.Text
		}
		bot.SendMessage(message.Chat, msg, nil)
	}
}
