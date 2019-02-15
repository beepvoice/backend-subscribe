package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "log"
  "net/http"
  "time"

  "github.com/julienschmidt/httprouter"
  "github.com/nats-io/go-nats"
  "github.com/golang/protobuf/proto"
)

type RawClient struct {
  UserId string
  ClientId string
}

type JsonResponse struct {
  Code uint32 `json:"code"`
  Message string `json:"message"`
}

var listen string
var natsHost string

var nc *nats.Conn
var connections map[RawClient]chan []byte

func main() {
  // Parse flags
  flag.StringVar(&listen, "listen", ":8080", "host and port to listen on")
  flag.StringVar(&natsHost, "nats", "nats://localhost:4222", "host and port of NATS")
  flag.Parse()

  connections = make(map[RawClient]chan []byte)

  // Routes
	router := httprouter.New()
  router.GET("/subscribe/:userid/client/:clientid", Subscribe)

  // NATS
  var err error
  nc, err = nats.Connect(natsHost)
  if err != nil {
		log.Fatal(err)
	}
  nc.Subscribe("res", ResponseHandler)

  // Start server
  log.Printf("starting server on %s", listen)
	log.Fatal(http.ListenAndServe(listen, router))
}

func Subscribe(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  flusher, ok := w.(http.Flusher)
  if !ok {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "text/event-stream")
  w.Header().Set("Cache-Control", "no-cache")
  w.Header().Set("Connection", "keep-alive")

  client := RawClient {
    UserId: p.ByName("userid"),
    ClientId: p.ByName("clintid"),
  }
  recv := make(chan []byte)
  connections[client] = recv;

  // Refresh connection periodically
  resClosed := w.(http.CloseNotifier).CloseNotify()
  ticker := time.NewTicker(25 * time.Second)

  for {
    select {
      case msg := <- recv:
        fmt.Fprintf(w, "data: %s\n\n", msg)
        flusher.Flush()
      case <- ticker.C:
        w.Write([]byte(":\n\n"))
      case <- resClosed:
        ticker.Stop()
        delete(connections, client)
        return
    }
  }
}

func ResponseHandler(msg *nats.Msg) {
  response := Response {}
  if err := proto.Unmarshal(msg.Data, &response); err != nil { // Fail quietly
    log.Println(err)
    return
  }

  client := RawClient {
    UserId: response.Client.Key,
    ClientId: response.Client.Client,
  }
  connection, present := connections[client]
  if present {
    res := JsonResponse {
      Code: response.Code,
      Message: string(response.Message),
    }
    json_string, err := json.Marshal(res)
    if err != nil {
      log.Println(err)
      return
    }
    connection <- []byte(json_string)
  }
}
