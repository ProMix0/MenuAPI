package model

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"os"
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
	var ids []int
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
		ids = append(ids, item)
	}
	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(ids)
	if err != nil {
		panic(err)
	}
}

func (model Model) GetItem(writer http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(request)["id"])
	row := model.db.QueryRow("SELECT Name, (Count - Locked) AS Free, Price FROM Menu WHERE Id = $1", id)
	if row.Err() != nil {
		panic(row.Err())
	}
	var free, price int
	var name string
	err := row.Scan(&name, &free, &price)
	if err != nil {
		writer.WriteHeader(404)
		return
	}
	response := make(map[string]interface{}, 4)
	response["id"] = id
	response["name"] = name
	response["available"] = free
	response["price"] = price
	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		panic(err)
	}
}

func (model Model) GetItemImage(writer http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(request)["id"])
	row := model.db.QueryRow("SELECT Image FROM Menu WHERE Id = $1", id)
	if row.Err() != nil {
		panic(row.Err())
	}
	image := make([]byte, 256*1024)
	err := row.Scan(&image)
	if err != nil {
		writer.WriteHeader(404)
		return
	}
	writer.Header().Set("Content-Type", "image/jpeg")
	_, err = writer.Write(image)
	if err != nil {
		panic(err)
	}
}

type purchaseItem struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

func (model Model) Purchase(writer http.ResponseWriter, request *http.Request) {
	var requestBody []purchaseItem
	err := json.NewDecoder(request.Body).Decode(&requestBody)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	state := validate(model, requestBody)
	if !state.valid() {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(writer).Encode(state)
		if err != nil {
			panic(err)
		}
		return
	}

	for _, item := range requestBody {
		_, err = model.db.Exec("UPDATE Menu SET Locked = Locked + $1 WHERE Id = $2", item.Count, item.Id)
		if err != nil {
			panic(err)
		}
	}

	//TODO get QR-code
	writer.Header().Set("Content-Type", "image/jpeg")
	file, _ := os.Open("QR-code.jpg")
	_, err = io.Copy(writer, file)
	file.Close()
	if err != nil {
		panic(err)
	}

	//TODO get paying confirmation
	for _, item := range requestBody {
		_, err = model.db.Exec("UPDATE Menu SET Locked = Locked - $1, Count = Count - $1 WHERE Id = $2", item.Count, item.Id)
		if err != nil {
			panic(err)
		}
	}
}

type purchaseState struct {
	InvalidIds    []int          `json:"invalidIds"`
	InvalidCounts []purchaseItem `json:"invalidCounts"`
}

func (state purchaseState) valid() bool {
	return len(state.InvalidCounts) == 0 && len(state.InvalidIds) == 0
}

func validate(model Model, request []purchaseItem) (result purchaseState) {
	for _, item := range request {
		row := model.db.QueryRow("SELECT (Count - Locked) AS Free FROM Menu WHERE Id = $1", item.Id)
		if row.Err() != nil {
			panic(row.Err())
		}
		var count int
		err := row.Scan(&count)
		if err != nil {
			result.InvalidIds = append(result.InvalidIds, item.Id)
			continue
		}
		if item.Count <= 0 || item.Count > count {
			result.InvalidCounts = append(result.InvalidCounts, item)
		}
	}
	return
}
