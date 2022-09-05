package db

import (
	"fmt"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var Mysql_connection string
var DB *sql.DB
var err error
var TGToken string

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
}

func userExists(userID int64) bool {
	return false
}

func Increment(userID int64, reps int) (int, error) {

	stmt, err := DB.Prepare(`INSERT INTO reps (telegram_id, reps_amount, reps_date) VALUES 
	(?, ?, date(NOW()))
	ON DUPLICATE KEY UPDATE reps_amount = reps_amount + ?;`)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(userID, reps, reps)
	if err != nil {
		return 0, err
	}
	fmt.Println(res)

	stmt, err = DB.Prepare(`SELECT (r.reps_amount) FROM reps r
	WHERE r.telegram_id = ? AND r.reps_date = date(NOW());
	`)
	if err != nil {
		return 0, err
	}

	var curReps int
	stmt.QueryRow(userID).Scan(&curReps)

	return curReps, nil
}

func AddUser(userID int64) error {
	stmt, err := DB.Prepare(`INSERT INTO person (telegram_id, state_id, first_interaction, last_interaction) VALUES
	(?, 0, NOW(), NOW())
	ON DUPLICATE KEY UPDATE telegram_id = telegram_id, last_interaction = NOW();`)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(userID)
	if err != nil {
		return err
	}
	fmt.Println(res)

	return nil
}

func ClearTodayProgress(userID int64) error {

	stmt, err := DB.Prepare(`DELETE FROM reps
	WHERE telegram_id = ? AND reps_date = date(NOW());`)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(userID)
	if err != nil {
		return err
	}
	fmt.Println(res)

	return nil
}
