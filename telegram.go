package sopromocao

import (
	"fmt"
	"log"
	"time"

	"regexp"

	"github.com/tucnak/telebot"
)

var regComando *regexp.Regexp
var regMensagem *regexp.Regexp
var telegramBot *telebot.Bot

//dev 278478343:AAGXj08TDDnM6_fv7yiZNK78ZkE-UH9Zges
//prod 233668925:AAFFwfhIY292fmTxKZgFjcFDsCF9FpwZVQ0
func ShowTelegram(token string, urls chan string) {
	bot, err := telebot.NewBot(token)
	if err != nil {
		log.Fatalln(err)
	}
	telegramBot = bot

	regComando, err = regexp.Compile("(/[\\w-_]+) ?")
	if err != nil {
		panic(err)
	}
	regMensagem, err = regexp.Compile("(/[\\w-_]+) (.*?)$")
	if err != nil {
		panic(err)
	}

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, 1*time.Second)
	for message := range messages {
		log.Println("new message ", message.Sender.FirstName, "ID", message.Sender.ID)
		fmt.Println(message.Text)

		command := TelegramGetCommand(message.Text)
		switch command {
		case "/novo":
			msg := TelegramCommandGetMessage(message.Text)
			if msg == "" {
				bot.SendMessage(message.Chat, "Tente novamente, nao foi possivel identificar o link enviado", nil)
			} else {
				TelegramCommandNovo(message, msg, urls)
			}
		case "/monitore":
			bot.SendMessage(message.Chat, "Comando liberado mas ainda nao disponivel, aguarde", nil)
		case "/ajuda":
			bot.SendMessage(message.Chat, "/novo Link com o produto para compartilhar \n/monitore Monitore ofertas a partir de palavras chaves exemplo: motorola smartphone 32gb", nil)
		default:
			bot.SendMessage(message.Chat, "não identifiquei o comando, digite /ajuda para ajuda", nil)
		}

	}
}

func TelegramCommandNovo(telegramMessage telebot.Message, m string, c chan string) {
	var msg string
	_, loja := IdentifyNomeLoja(m)
	if loja == "" {
		msg = telegramMessage.Sender.FirstName + " o link enviado não é suportado pelo sistema."
	} else {
		msg = telegramMessage.Sender.FirstName + " muito obrigado por compartilhar, irei processar e posteriormente publicar no site http://www.radardaoferta.com.br/"
		c <- m
	}
	telegramBot.SendMessage(telegramMessage.Chat, msg, nil)
}

func TelegramGetCommand(s string) string {
	if regComando.MatchString(s) {
		match := regComando.FindStringSubmatch(s)
		return match[1]
	}
	return ""
}

func TelegramCommandGetMessage(s string) string {
	if regMensagem.MatchString(s) {
		match := regMensagem.FindStringSubmatch(s)
		return match[2]
	}
	return ""
}
