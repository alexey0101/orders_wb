package main

import (
	"app/db/impl"
	"app/nats"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func userInput(cancelChan chan<- struct{}) {

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Type 'stop' to close application: ")
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) == "stop" {
			cancelChan <- struct{}{}
			return
		}
	}
}

func main() {
	cache, err := impl.NewCache()
	if err != nil {
		log.Fatal(err)
	}
	err = cache.LoadOrders()
	if err != nil {
		log.Fatal(err)
	}

	cancelChan := make(chan struct{})
	go nats.Subscribe(cancelChan, cache, "test-cluster", "sub", "orders", "orders")

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		emptyStruct := struct{}{}
		uid := r.URL.Query().Get("uid")
		if uid == "" {
			json.NewEncoder(w).Encode(emptyStruct)
			return
		}

		order, err := cache.FindByUid(uid)
		if err != nil {
			json.NewEncoder(w).Encode(emptyStruct)
			return
		}

		json.NewEncoder(w).Encode(order)
	})

	http.Handle("/", http.FileServer(http.Dir("templates/")))

	go http.ListenAndServe(":3000", nil)

	go userInput(cancelChan)

	select {
	case <-cancelChan:
		log.Default().Println("App closed!")
	}
}
