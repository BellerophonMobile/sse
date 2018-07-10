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
	handler := sse.NewHandler(sse.HandlerConfig{
		RetryTime:    5 * time.Second,
		WriteTimeout: 1 * time.Second,
		BufferSize:   20,
		HistoryLimit: 10,
	})

	server := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 200 * time.Millisecond,
	}
	shutdown := make(chan struct{})

	go func() {
		c := 0
		for {
			evt := &sse.Event{
				ID:   fmt.Sprintf("%d", c),
				Data: fmt.Sprintf(message, c),
			}

			if c%10 == 0 {
				evt.Type = "urgentupdate"
			}

			handler.Send(evt)

			fmt.Println("Generated:", evt.Data)
			time.Sleep(1 * time.Second)
			c++
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

		// First, shut down the server to get all clients to disconnect.
		server.Shutdown(ctx)
		cancel()

		// Then, close the SSE handler.
		handler.Close()
		close(shutdown)
	}()

	http.Handle("/events", handler)
	http.HandleFunc("/", viewer)

	log.Fatal(server.ListenAndServe())

	<-shutdown
}

const html = `
<!doctype html>
<html>
<body>
<p>Events:</p>
<div id="events"></div>

<script>
	var events = document.getElementById('events');

	var source = new EventSource('/events');
	source.addEventListener('open', function (e) {
		console.log('open:', e);
	});
	source.addEventListener('error', function (e) {
		console.log('error:', e);
	});
	source.addEventListener('message', function (e) {
		console.log('message:', e.data);
		var p = document.createElement('p');
		p.textContent = e.data;
		events.insertBefore(p, events.firstChild);
	});
	source.addEventListener('urgentupdate', function (e) {
		console.log('urgent update:', e.data);
		var p = document.createElement('p');
		p.textContent = e.data;
		p.style.color = 'red';
		events.insertBefore(p, events.firstChild);
	});
</script>
</body>
</html>
`

func viewer(w http.ResponseWriter, r *http.Request) {
	log.Println("Viewer")
	fmt.Fprint(w, html)
}
