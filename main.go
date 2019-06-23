package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"
  "time"

  . "subscribe/backend-protobuf/go"

  "github.com/joho/godotenv"
  "github.com/julienschmidt/httprouter"
  "github.com/nats-io/go-nats"
  "github.com/golang/protobuf/proto"
)

type RawClient struct {
  UserId string `json:"userid"`
  ClientId string `json:"clientid"`
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
  // Load .env
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
  listen = os.Getenv("LISTEN")
  natsHost = os.Getenv("NATS")

  connections = make(map[RawClient]chan []byte)

  // Routes
	router := httprouter.New()
  router.GET("/subscribe", Subscribe)

  // NATS
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

  ua := r.Header.Get("X-User-Claim")
  if ua == "" {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
  }

  var client RawClient
  err := json.Unmarshal([]byte(ua), &client)

  if err != nil {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
  }

  w.Header().Set("Content-Type", "text/event-stream")
  w.Header().Set("Cache-Control", "no-cache")
  w.Header().Set("Connection", "keep-alive")

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
