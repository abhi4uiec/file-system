package main

import (
	"log"
	"net/http"

	controller "filesystem/controller"

	"github.com/gorilla/mux"
)

func handleRequests() {
	// creates a new instance of a mux router
	router := mux.NewRouter().StrictSlash(true)

	// register rest endpoints
	router.HandleFunc("/files", controller.FetchFilesBasedOnCriteria).Methods("POST")

	// pass in newly created mux router as the second argument
	log.Fatal(http.ListenAndServe(":8000", router), "Failed to bootstrap the server")
}

func main() {

	handleRequests()
}
