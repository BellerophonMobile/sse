package main

import (
	"log"
	"net/http"

	"github.com/BellerophonMobile/sse"
)

const url = "http://localhost:8080/events"

const headerAccept = "Accept"

func main() {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln("Error creating request:", err)
		return
	}

	req.Header.Set(headerAccept, sse.MIMETypeSSE)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	reader := sse.NewReader(resp.Body)
	for {
		evt, err := reader.NextEvent()
		if err != nil {
			log.Fatalln("Error reading event:", err)
			return
		}

		if evt.ID != "" {
			log.Println("ID:", evt.ID)
		}
		if evt.Type != "" {
			log.Println("Type:", evt.Type)
		}
		log.Printf("Data: %s\n\n", evt.Data)
	}
}
