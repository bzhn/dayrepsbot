package db

import (
	"fmt"
	"os"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Mysql_connection string
var DB *sql.DB
var err error
var TGToken string

// Prepared database Statements
var ps = make(map[string]*sql.Stmt)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	TGToken = os.Getenv("TELEGRAM_API_TOKEN")

	Mysql_connection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_IP"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB_NAME"))

	DB, err = sql.Open("mysql", Mysql_connection)
	if err != nil {
		panic(err.Error())
	}

	inc, err := DB.Prepare(`	
	INSERT INTO reps (reps_amount, reps_date, reps_id)
	VALUES (?, ?, ?)
	ON DUPLICATE KEY UPDATE reps_amount = reps_amount + ?;`)
	if err != nil {
		panic(err)
	}
	ps["increment"] = inc

	amnt, err := DB.Prepare(`SELECT r.reps_amount FROM reps r
	WHERE r.reps_date = ? AND r.reps_id = ?;
	`)
	if err != nil {
		panic(err)
	}
	ps["todayamount"] = amnt

	dltd, err := DB.Prepare(`DELETE FROM reps
	WHERE reps_date = ? AND reps_id = ?;`)
	if err != nil {
		panic(err)
	}
	ps["deletetodayreps"] = dltd

	dltdall, err := DB.Prepare(`DELETE r FROM reps r
	LEFT JOIN dct_reps dr ON r.reps_id = dr.reps_id
	WHERE r.reps_date = ?
	AND dr.telegram_id = ?;`)
	if err != nil {
		panic(err)
	}
	ps["deletetodayrepsall"] = dltdall

	adusr, err := DB.Prepare(`INSERT INTO person (telegram_id, state_id, first_interaction, last_interaction, person_name) VALUES
	(?, 0, NOW(), NOW(), ?)
	ON DUPLICATE KEY UPDATE last_interaction = NOW(), person_name = ?;`)
	if err != nil {
		panic(err)
	}
	ps["adduser"] = adusr

	gtunm, err := DB.Prepare(`SELECT p.person_name
FROM person p
WHERE p.telegram_id = ?;`)
	if err != nil {
		panic(err)
	}
	ps["getusername"] = gtunm

	gtrpid, err := DB.Prepare(`SELECT dr.reps_id
	FROM dct_reps dr
	WHERE dr.telegram_id = ?
	AND dr.reps_name = ?
	LIMIT 1;`)
	if err != nil {
		panic(err)
	}
	ps["getrepsid"] = gtrpid

	addex, err := DB.Prepare(`INSERT INTO dct_reps(telegram_id, reps_name) VALUES (?, ?);`)
	if err != nil {
		panic(err)
	}
	ps["addnewexercise"] = addex

	exlst, err := DB.Prepare(`SELECT dr.reps_name
	FROM dct_reps dr
	WHERE dr.telegram_id = ?
	ORDER BY dr.order_value, -dr.reps_id;`)
	if err != nil {
		panic(err)
	}
	ps["getexercises"] = exlst

}

func userExists(userID int64) bool {
	return false
}

func Increment(userID int64, reps int, repsName string) (int, error) {

	res, err := ps["increment"].Exec(reps, GetUserDate(userID), GetRepsID(userID, repsName), reps)
	if err != nil {
		return 0, err
	}
	fmt.Println(res)

	var curReps int
	ps["todayamount"].QueryRow(GetUserDate(userID), GetRepsID(userID, repsName)).Scan(&curReps) // TODO: Provide the reps name

	return curReps, nil
}

func AddUser(userID int64, userName string) error {

	res, err := ps["adduser"].Exec(userID, userName, userName)
	if err != nil {
		return err
	}
	fmt.Println(res)

	return nil
}

func ClearTodayProgress(userID int64, repsName string) error {

	res, err := ps["deletetodayreps"].Exec(GetUserDate(userID), GetRepsID(userID, repsName))
	if err != nil {
		return err
	}
	fmt.Println(res)

	return nil
}

func ClearTodayProgressAll(userID int64) error {

	res, err := ps["deletetodayrepsall"].Exec(GetUserDate(userID), userID)
	if err != nil {
		return err
	}
	fmt.Println(res)

	return nil
}

func UserRepsAmount(userID int64, repsName string) (int, error) {
	panic("")
}

func GetUserDate(userID int64) string {
	// Check when user's day starts at and change the date if it starts later or sooner

	return time.Now().Format("20060102")
}

func GetUserName(userID int64) string {
	var username string

	if err := ps["getusername"].QueryRow(userID).Scan(&username); err != nil {
		return "Anonymous"
	}

	return username
}

func GetRepsID(userID int64, repsName string) int {
	var rpid int
	if err := ps["getrepsid"].QueryRow(userID, repsName).Scan(&rpid); err != nil {
		panic(err)
	}
	return rpid
}

func AddNewExercise(userID int64, repsName string) error {
	if err := verifyExerciseName(repsName); err != nil {
		return err
	}

	_, err := ps["addnewexercise"].Exec(userID, repsName)
	if err != nil {
		return err
	}

	return nil
}

func verifyExerciseName(exerciseName string) error {
	return nil // TODO
}

// UserKeyboard generates the telegram keyboard according to
// exercises he has added. Also there are default buttons
func UserKeyboard(userID int64) tgbotapi.ReplyKeyboardMarkup {
	var nextExercise string
	var allExercises []string
	var ButtonRows [][]tgbotapi.KeyboardButton

	res, err := ps["getexercises"].Query(userID)
	if err != nil {
		panic(err)
	}

	for res.Next() {
		res.Scan(&nextExercise)
		allExercises = append(allExercises, nextExercise)
	}

	for i := len(allExercises); i > 0; {
		if i >= 2 {
			ButtonRows = append(ButtonRows, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(allExercises[len(allExercises)-i]),
				tgbotapi.NewKeyboardButton(allExercises[len(allExercises)-i+1]),
			))
			i -= 2
		} else {

			ButtonRows = append(ButtonRows, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(allExercises[len(allExercises)-i]),
			))
			i--
		}
	}

	return tgbotapi.NewReplyKeyboard(ButtonRows...)
}

// IfButton checks if the message of user is
// one of the buttons or not. If not, false will be returned
func IfButton(userID int64, textToCheck string) bool {
	var smthng int
	if err := ps["getrepsid"].QueryRow(userID, textToCheck).Scan(&smthng); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

var del = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Розклад"),
		tgbotapi.NewKeyboardButton("Яка зараз пара"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Мої викладачі"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Налаштування акаунту"),
	),
)
