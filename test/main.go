package main

import (
	"time"
	"net/http"
	"fmt"
	"log"
	"github.com/BellerophonMobile/sse"
)

func main() {

	http.HandleFunc("/events", Generator)
	http.HandleFunc("/view", Viewer)
	log.Fatal(http.ListenAndServe(":8080", nil))
	
}

func Viewer(w http.ResponseWriter, r *http.Request) {
	log.Println("Viewer")
	
	fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<body>
Events:

	<script type="text/javascript">
	    var source = new EventSource('/events');
      source.onopen = function(e) {
          console.log("OnOpen:" + e)
      }
      source.onerror = function(e) {
          console.log("OnError: " + e)
      }
	    source.onmessage = function(e) {
          console.log("OnMessage:" + e)
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

func Generator(w http.ResponseWriter, r *http.Request) {

	log.Println("Generator")
	
	writer,err := sse.NewWriter(w, r)
	if err != nil {
		fmt.Fprint(w, "SSE unsupported")
		return
	}

	c := 0
	for {
		writer.Event("urgentupdate", []byte("OK"))
		time.Sleep(1 * time.Second)
		c++
	}

}
