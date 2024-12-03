package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "leonid"
	password = "qwerty1234"
	dbname   = "sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Обработчики HTTP-запросов
func (h *Handlers) GetCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	msg, err := h.dbProvider.GetCount()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
	}
}

func (h *Handlers) ChangeCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	input := struct {
		Count string `json:"count"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)

	countValue, err2 := strconv.ParseInt(input.Count, 10, 32)

	if err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("count - not a number!"))
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = h.dbProvider.IncrementCount(int(countValue))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("OK"))
		w.WriteHeader(http.StatusCreated)
	}
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) GetCount() (string, error) {
	var name string
	row := dp.db.QueryRow("SELECT count FROM count LIMIT 1")
	err := row.Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}
func (dp *DatabaseProvider) IncrementCount(value int) error {
	fmt.Println(value)
	_, err := dp.db.Exec("UPDATE count SET count = count + ($1)", value)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Формирование строки подключения для postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Создание соединения с сервером postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер для БД с набором методов
	dp := DatabaseProvider{db: db}
	// Создаем экземпляр структуры с набором обработчиков
	h := Handlers{dbProvider: dp}

	// Регистрируем обработчики
	http.HandleFunc("/count/update", h.ChangeCount)
	http.HandleFunc("/count/get", h.GetCount)

	// Запускаем веб-сервер на указанном адресе
	err = http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatal(err)
	}
}
