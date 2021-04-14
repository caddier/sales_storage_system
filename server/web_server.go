package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type WebServer struct {
	router     *mux.Router
	handlers   *WebHandlers
	listenPort int
}

func NewWebServer(h *WebHandlers, port int) *WebServer {
	return &WebServer{
		router:     mux.NewRouter(),
		handlers:   h,
		listenPort: port,
	}
}

func (w *WebServer) InitHandlers() {
	w.router.HandleFunc("/api/inputgoods", w.handlers.HandleInputGoods).Methods("POST")
	w.router.HandleFunc("/api/querygoods", w.handlers.HandleQueryGoods).Methods("POST")
	w.router.HandleFunc("/api/sellgoods", w.handlers.HandleSellGoods).Methods("POST")
	w.router.HandleFunc("/api/querylocation", w.handlers.HandleQueryLocation).Methods("POST")
}

func (w *WebServer) Start() {
	w.InitHandlers()
	addr := fmt.Sprintf(":%d", w.listenPort)
	server := http.Server{
		Handler:      w.router,
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		LogError("Start web server failed, %s", err.Error())
	}
}
