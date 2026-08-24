package main

import (
	"fmt"
	"net/http"
	"os"
	// Boshqa barcha kutubxona va importlaringiz shu yerda bo'lsin...
)

func main() {
	// 1. Render talab qiladigan HTTP server qismi
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Bot is running!")
		})
		http.ListenAndServe(":"+port, nil)
	}()

	// 2. Shu yerdan pastga o'zingizning asl Telegram bot kodingizni joylang...
}
go func() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Bot is running!")
    })
    http.ListenAndServe(":"+port, nil)
}()
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Render port talab qilgani uchun soxta HTTP server
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Bot is running!")
		})
		log.Println("HTTP server listening on port", port)
		http.ListenAndServe(":"+port, nil)
	}()

	log.Println("Bot ishga tushdi...")

	// Bu yerda botingizning asosiy kodi davom etadi
	// Cheksiz kutish jarayoni
	select {}
}
