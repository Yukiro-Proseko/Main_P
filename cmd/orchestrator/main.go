package main

import (
	"log"
	"net/http"
	"paral/internal/orchestrator"

	"github.com/gorilla/mux"
)

func main() {
	app := orchestrator.New()

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/expressions/{id}", app.GetExpressionByIDHandler).Methods("GET")
	router.HandleFunc("/internal/task", app.GetPendingTaskHandler).Methods("GET")
	router.HandleFunc("/internal/task/result", app.SubmitTaskResultHandler).Methods("POST", "GET")
	router.HandleFunc("/api/v1/calculate", app.AddExpressionHandler).Methods("POST")
	router.HandleFunc("/api/v1/expressions", app.GetAllExpressionsHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
