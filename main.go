package main

import (
	modelPkg "MenuAPI/model"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"runtime/debug"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "localhost:8082", "Address with port")
	flag.Parse()
	if addr == "" {
		flag.Usage()
		os.Exit(1)
	}

	router := mux.NewRouter()
	model := modelPkg.NewModel("menu.db")
	defer model.Close()

	router.Use(panicRecovery)

	router.HandleFunc("/list", model.EnumerateIds).Methods("GET")
	router.HandleFunc("/item/{id:[0-9]+}", model.GetItem).Methods("GET")
	router.HandleFunc("/item/{id:[0-9]+}/image", model.GetItemImage).Methods("GET")
	router.HandleFunc("/purchase", model.Purchase).Methods("POST")

	http.Handle("/", router)
	fmt.Println(http.ListenAndServe(addr, nil))
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
