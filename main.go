package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type wallet map[string]float64

type binanceResp struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("XXXXXXXXXXXXXXXXXXXXS")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil { // If we got a message
			msgArr := strings.Split(update.Message.Text, " ")
			switch msgArr[0] {
			case "ADD":
				sum, err := strconv.ParseFloat(msgArr[2], 64)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Warning1"))
					continue
				}
				if _, ok := db[update.Message.Chat.ID]; !ok {
					db[update.Message.Chat.ID] = wallet{}
				}
				db[update.Message.Chat.ID][msgArr[1]] += sum
				msg := fmt.Sprintf("Balanse: %s %f", msgArr[1], sum)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			case "SUB":
				sum, err := strconv.ParseFloat(msgArr[2], 64)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Warning1"))
					continue
				}
				if _, ok := db[update.Message.Chat.ID]; !ok {
					db[update.Message.Chat.ID] = wallet{}
				}
				db[update.Message.Chat.ID][msgArr[1]] -= sum
				msg := fmt.Sprintf("Balanse: %s %f", msgArr[1], sum)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			case "DEL":
				delete(db[update.Message.Chat.ID], msgArr[1])
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "For delete"))
			case "SHOW":
				msg := "Balance:\n"
				var usdSum float64
				for key, val := range db[update.Message.Chat.ID] {
					coinPrice, err := getPrice(key)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Err"))
					}
					usdSum += val * coinPrice
					msg += fmt.Sprintf(" %s: %.2f [%.2f]\n", key, val, val*coinPrice)
				}
				msg += fmt.Sprintf("Cумма: %.2f\n", usdSum)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unnown command"))
			}
		}
	}
}

func getPrice(c string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sCUSDT", c))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	if jsonResp.Code != 0 {
		msg := "Invalid value"
		fmt.Println(msg)
	}
	price = jsonResp.Price
	return
}
