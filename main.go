package main

import (
	modelPkg "MenuAPI/model"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"runtime/debug"
)

func main() {
	router := mux.NewRouter()
	model := modelPkg.NewModel("menu.db")
	defer model.Close()

	router.Use(panicRecovery)

	router.HandleFunc("/list", model.EnumerateIds).Methods("GET")
	router.HandleFunc("/item/{id:[0-9]+}", model.GetItem).Methods("GET")
	router.HandleFunc("/item/{id:[0-9]+}/image", model.GetItemImage).Methods("GET")
	router.HandleFunc("/purchase", model.Purchase).Methods("POST")

	http.Handle("/", router)
	fmt.Println(http.ListenAndServe("localhost:8082", nil))
}

func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				fmt.Println(err)
				fmt.Println(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, req)
	})
}
