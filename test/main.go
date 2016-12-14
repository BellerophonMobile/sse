package main

import (
	"time"
	"net/http"
	"fmt"
	"log"
	"github.com/BellerophonMobile/sse"
)

func main() {

	out := sse.NewEventServer(nil)

	go func() {
	c := 0
	for {
		msg := fmt.Sprintf(
`Just sit right back and you'll hear a tale
a tale of a fateful trip,
that started from this tropic port,
aboard this tiny ship.

[%v]`, c)
		out.Message(msg)
		fmt.Println("Generated " + msg)
		time.Sleep(1 * time.Second)
		c++
	}
	}()
	
	http.HandleFunc("/events", out.Handle)
	http.HandleFunc("/view", Viewer)
	log.Fatal(http.ListenAndServe(":8080", nil))

	
}

func Viewer(w http.ResponseWriter, r *http.Request) {
	log.Println("Viewer")
	
	fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<body>
Events:<br/>

	<script type="text/javascript">
	    var source = new EventSource('/events');
      source.onopen = function(e) {
          console.log("OnOpen:" + e)
      }
      source.onerror = function(e) {
          console.log("OnError: " + e)
      }
	    source.onmessage = function(e) {
          console.log("OnMessage:" + e.data)
	        document.body.innerHTML += e.data + '<br>';
	    }
      source.addEventListener("urgentupdate", function(e) {
          console.log("Update:" + e)
	        document.body.innerHTML += e.data + '<br>';
      });
	</script>
</body>
</html>
`)
	
}
