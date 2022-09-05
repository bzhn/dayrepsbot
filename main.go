package main

import (
	"fmt"
	"log"
	"strconv"

	db "github.com/bzhn/dayrepsbot/pkg/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var IKBresetConfirm = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Reset all today's progress", "reset_today"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("No, don't remove my progress", "reset_cancel"),
	),
)

func main() {
	bot, err := tgbotapi.NewBotAPI(db.TGToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			// msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "")

			switch update.CallbackQuery.Data {
			case "reset_today":
				err := db.ClearTodayProgress(update.CallbackQuery.From.ID)
				if err != nil {
					panic(err)
				}
				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID, "Successfully removed your reps")
				bot.Send(msg)

			case "reset_cancel":
				err := db.ClearTodayProgress(update.CallbackQuery.From.ID)

				if err != nil {
					panic(err)
				}
				msg := tgbotapi.NewDeleteMessage(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID)
				bot.Send(msg)

			}

			continue
		}

		if update.Message == nil || update.Message.Text == "" { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() { // if Message was a command

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			// Extract the command from the Message.
			switch update.Message.Command() {
			case "start":
				fmt.Println("Got start command")
				err := db.AddUser(update.Message.From.ID)

				if err != nil {
					panic(err)
				}

				msg.Text = "Successfully added you to database."

			case "help":
				msg.Text = "I understand /sayhi and /status."
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			case "reset":
				msg.Text = "Are you sure you want to remove your progress?"
				msg.ReplyMarkup = IKBresetConfirm
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue

		}

		if n, err := strconv.Atoi(update.Message.Text); err == nil {
			if n > 2147483647 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, amount of reps is too big!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			}

			curReps, err := db.Increment(update.Message.From.ID, n)
			if err != nil {
				panic(err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You have done %d reps today!", curReps))

			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, I don't understand")
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}

	}
}
