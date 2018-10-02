package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/BellerophonMobile/sse"
)

const message = `Just sit right back and you'll hear a tale
a tale of a fateful trip,
that started from this tropic port,
aboard this tiny ship.

[%v]`

func main() {
	shutdownChan := make(chan struct{})

	handler := sse.NewHandler(sse.HandlerConfig{
		RetryTime:    5 * time.Second,
		HistoryLimit: 10,
	})

	server := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 200 * time.Millisecond,
	}

	http.Handle("/events", handler)
	http.HandleFunc("/", viewer)

	go sendLoop(handler)
	go waitSignal(handler, &server, shutdownChan)

	log.Fatal(server.ListenAndServe())

	<-shutdownChan
}

func sendLoop(handler *sse.Handler) {
	for i := 0; true; i++ {
		evt := &sse.Event{
			ID:   fmt.Sprintf("%d", i),
			Data: fmt.Sprintf(message, i),
		}

		if i%10 == 0 {
			evt.Type = "urgentupdate"
		}

		if err := handler.Send(evt); err != nil {
			log.Println("event send error:", err)
			return
		}

		log.Println("generated:", evt.Data)
		time.Sleep(1 * time.Second)
	}
}

func waitSignal(handler *sse.Handler, server *http.Server, shutdownChan chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	// First, shut down the server to get all clients to disconnect.
	server.Shutdown(ctx)
	cancel()

	// Then, close the SSE handler.
	handler.Close()
	close(shutdownChan)
}

func viewer(w http.ResponseWriter, r *http.Request) {
	log.Println("viewer")
	fmt.Fprint(w, html)
}
