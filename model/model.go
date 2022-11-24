package model

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
)

type Model struct {
	db *sql.DB
}

func NewModel(dbPath string) Model {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	return Model{db}
}

func (model Model) Close() {
	model.db.Close()
}

func (model Model) EnumerateIds(writer http.ResponseWriter, request *http.Request) {
	ids := make([]int, 16)
	rows, err := model.db.Query("SELECT Id FROM Menu")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var item int
		err = rows.Scan(&item)
		if err != nil {
			panic(err)
		}
	}
	response, err := json.Marshal(ids)
	if err != nil {
		panic(err)
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(response)
	if err != nil {
		panic(err)
	}
}

func (model Model) GetItem(writer http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(request)["id"])
	row := model.db.QueryRow("SELECT Name, (Count - Locked) AS Free FROM Menu WHERE Id = $1", id)
	if row.Err() != nil {
		writer.WriteHeader(404)
		return
	}
	var free int
	var name string
	err := row.Scan(&name, &free)
	if err != nil {
		panic(err)
	}
	response := make(map[string]interface{}, 3)
	response["id"] = id
	response["name"] = name
	response["available"] = free
	js, err := json.Marshal(response)
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(js)
	if err != nil {
		panic(err)
	}
}

func (model Model) GetItemImage(writer http.ResponseWriter, request *http.Request) {
	//TODO read blob from db
}

func (model Model) Purchase(writer http.ResponseWriter, request *http.Request) {
	//TODO
}
