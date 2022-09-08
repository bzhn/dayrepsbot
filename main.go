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
		tgbotapi.NewInlineKeyboardButtonData("Reset one exercise", "reset_today"),
		tgbotapi.NewInlineKeyboardButtonData("Reset for all", "reset_all_today"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("No, don't remove my progress", "reset_cancel"),
	),
)

var lastButton = make(map[int64]string)
var lastCommand = make(map[int64]string)

func main() {
	bot, err := tgbotapi.NewBotAPI(db.TGToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "")

			switch update.CallbackQuery.Data {
			case "reset_today":
				msg.Text = "Click the exercise to reset progress\n\nI'm going to reset the progress you've made today for one exercise. Please click the button on the keyboard with your exercises to remove it's progress. If you want to cancel this action, simply type /cancel"
				lastCommand[update.CallbackQuery.From.ID] = update.CallbackQuery.Data
				bot.Send(msg)

			case "reset_all_today":
				msg.Text = "⚠️ Are you sure you want to reset all progress you made today for all exercises?\n\nI'm going to reset the progress you made today for all exercises. If you want to cancel this action, simply type /cancel"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Yes, remove all my progress for today", "reset_all_today_sure"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("No, don't remove my progress", "reset_cancel"),
					),
				)
				bot.Send(msg)

			case "reset_all_today_sure":
				db.ClearTodayProgressAll(update.CallbackQuery.From.ID)
				deleteLast(update.CallbackQuery.From.ID)
				msgdlt := tgbotapi.NewDeleteMessage(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID)
				bot.Send(msgdlt)
				msg.Text = "Your today's progress was successfully removed!"
				bot.Send(msg)

			case "reset_cancel":

				deleteLast(update.CallbackQuery.From.ID)
				msg := tgbotapi.NewDeleteMessage(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID)
				bot.Send(msg)

			}

			continue
		}

		if update.Message == nil || update.Message.Text == "" { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() { // if Message was a command

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "/")
			lastCommand[update.Message.Chat.ID] = update.Message.Command()
			// Extract the command from the Message.
			switch update.Message.Command() {
			case "start":
				msg.Text = "Successfully added you to database."
				err := db.AddUser(update.Message.From.ID, update.Message.From.FirstName)

				if err != nil {
					if err == db.UserExists {
						msg.Text = "Welcome back! What about workout?"
						fmt.Println("User already exist")
					} else {
						panic(err)
					}
				}

				deleteLast(update.Message.Chat.ID)

			case "name":
				msg.Text = "Please write how to address you"
			case "newexercise":
				msg.Text = "Okay, what's the name of your exercise?"
			case "stats":
				msg.Text = "I'm ok."
				deleteLast(update.Message.Chat.ID)
			case "reset":
				msg.Text = "⚠️ Are you sure you want to remove your progress?"
				msg.ReplyMarkup = IKBresetConfirm
			case "cancel":
				msg.Text = "Canceled your pending actions! You're ready to go!"
				deleteLast(update.Message.Chat.ID)

			default:
				msg.Text = "I don't know that command"
				deleteLast(update.Message.Chat.ID)
				// Clear last action
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue

		}

		// If user sent some message after running a command
		if lc, ok := lastCommand[update.Message.Chat.ID]; ok {

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "/")
			msg.ReplyMarkup = db.UserKeyboard(update.Message.Chat.ID)
			switch lc {
			case "name":
				db.AddUser(update.Message.Chat.ID, update.Message.Text)
				msg.Text = "Nice to meet you, " + db.GetUserName(update.Message.Chat.ID)
				deleteLast(update.Message.Chat.ID)
			case "newexercise":
				err := db.AddNewExercise(update.Message.Chat.ID, update.Message.Text)
				if err != nil {
					panic(err)
				}
				msg.Text = "Your exercise have been added!"
				msg.ReplyMarkup = db.UserKeyboard(update.Message.Chat.ID)
				deleteLast(update.Message.Chat.ID)
			case "reset_today":
				if db.IfButton(update.Message.Chat.ID, update.Message.Text) {
					db.ClearTodayProgress(update.Message.Chat.ID, update.Message.Text)
					msg.Text = "✅ Progress removed\n\nYour progress for the exercise with the name " + update.Message.Text + " was successfully removed!"
					deleteLast(update.Message.Chat.ID)
				} else {
					msg.Text = "Sorry, you don't have the exercise with name " + update.Message.Text + ", so I can't remove it's today progress.\n\nPlease click one of the buttons with exercises to delete today's progress or type /cancel to cancel this action."

				}
			default:
				if db.IfButton(update.Message.Chat.ID, update.Message.Text) {
					lastButton[update.Message.Chat.ID] = update.Message.Text
					msg.Text = "Please type the amount of reps for exercise " + update.Message.Text + ", or type /cancel to cancel."
				} else {
					msg.Text = "Sorry, I can't understand. Use buttons and commands to communicate with the bot."
				}

				deleteLast(update.Message.Chat.ID)
			}

			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
			continue
		}

		// If the user sent a number after click on the button
		_, wasButton := lastButton[update.Message.From.ID]
		if n, err := strconv.Atoi(update.Message.Text); err == nil && wasButton {
			if n > 2147483647 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, amount of reps is too big!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			}

			curReps, err := db.Increment(update.Message.From.ID, n, lastButton[update.Message.From.ID])
			if err != nil {
				panic(err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You have done %d reps today!", curReps))

			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		} else { // If user sent text, but not a number
			if db.IfButton(update.Message.Chat.ID, update.Message.Text) {
				lastButton[update.Message.Chat.ID] = update.Message.Text
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Reps of "+update.Message.Text+". \nYou already done "+fmt.Sprintf("%d", db.GetTodaysAmount(update.Message.Chat.ID, lastButton[update.Message.From.ID]))+" reps. \n\nOkay, please type the amount of reps you made:")

				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
				lastButton[update.Message.Chat.ID] = update.Message.Text

				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, I don't understand")
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}

	}
}

// Validators

// ValidateRepsAmount checks if the provided string contains
// only positive number and that a sum of this number and of
// existing one is less than maximal number for
// unsigned smallint: 655335
func ValidateRepsAmount(string) {}

func deleteLast(userID int64) {
	if _, ok := lastCommand[userID]; ok {
		delete(lastCommand, userID)
	}

	if _, ok := lastButton[userID]; ok {
		delete(lastButton, userID)
	}
}
