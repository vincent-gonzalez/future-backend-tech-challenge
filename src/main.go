package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

func indexHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}

	switch req.Method {
	case http.MethodGet:
		res.Write([]byte("<h1>Hello World 2!</h1>"))
	}
}

func appointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res.Write([]byte(fmt.Sprintf("<h1>Appointments for trainer: %v </h1>", params.ByName("id"))))
}

func appointmentsPostHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	res.Write([]byte("<h1>Post successful!</h1>"))
}

func initDB() error {
	os.Remove("../db/api-test-sqlite.db")

	file, err := os.Create("../db/api-test-sqlite.db")
	if err != nil {
		return err
	}
	defer file.Close()

	database, err = sql.Open("sqlite3", "../db/api-test-sqlite.db")
	defer database.Close()

	createTableSQL := `CREATE TABLE appointments (
	"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT
	"trainer_id" integer
	"user_id" integer
	"starts_at" datetime
	"ends_at" datetime);`
	statement, err := database.Prepare(createTableSQL)
	if err != nil {
		//
	}
	statement.Exec()

}

func main() {
	fmt.Println("Hello World!")
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", indexHandler)
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/appointments/:id", appointmentsHandler)
	router.POST("/appointments", appointmentsPostHandler)
	log.Fatal(http.ListenAndServe(":8081", router))
}
