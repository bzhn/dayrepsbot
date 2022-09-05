package db

import (
	"fmt"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var Mysql_connection string
var DB *sql.DB

func init() {

	Mysql_connection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_IP"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB_NAME"))

	DB, err := sql.Open("mysql", Mysql_connection)
	if err != nil {
		panic(err.Error())
	}
	defer DB.Close()
}

func userExists(userID uint32) bool {

}

func Increment(userID uint32, reps int) (int, error) {

	return 0, nil
}
