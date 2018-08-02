package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/satori/go.uuid"
)

type server struct {
	ID            string
	MQTTBrokerURL string
	Port          int
}

type requestInfo struct {
	RequestTime time.Time
	Channel     chan []byte
}

var pendingRequests = make(map[string]requestInfo)

func (s server) httpResponseHandler(client mqtt.Client, msg mqtt.Message) {
	requestID, err := getRequestIDFromTopic(msg.Topic())
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	ri := pendingRequests[requestID]
	ri.Channel <- msg.Payload()
}

func (s server) handler(w http.ResponseWriter, h *http.Request) {
	clientID := h.Header.Get("MQHTTP-Client-ID")
	if len(clientID) == 0 {
		log.Println("MQHTTP-Client-ID header not present")
		http.Error(w, "MQHTTP-Client-ID header not present", http.StatusBadRequest)
		return
	}
	requestID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "Unable to generate UUID", http.StatusInternalServerError)
		return
	}
	log.Printf("%s %s %s\n", requestID, h.Method, h.URL.Path)
	rawRequest, err := httputil.DumpRequest(h, true)
	if err != nil {
		panic(err)
	}
	// Send MQTT Request
	ri := requestInfo{
		RequestTime: time.Now(),
		Channel:     make(chan []byte),
	}
	pendingRequests[requestID.String()] = ri
	defer delete(pendingRequests, requestID.String())
	topic := getHTTPRequestTopic(clientID, requestID.String())
	sendMqttMessage(s.MQTTBrokerURL, topic, rawRequest)
	response := <-ri.Channel
	if len(response) > 0 {
		hj, ok := w.(http.Hijacker)
		if !ok {
			panic("Webserver does not support hijacking")
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		bufrw.Write(response)
		bufrw.Flush()
	} else {
		http.Error(w, "Unable to contact the server", http.StatusInternalServerError)
	}
}

func cleanupRequests() {
	for {
		time.Sleep(requestTimeout / 2)
		now := time.Now()
		for requestID, ri := range pendingRequests {
			timeout := ri.RequestTime.Add(requestTimeout)
			log.Printf("Checking request %s\n", requestID)
			if now.After(timeout) {
				ri.Channel <- []byte("")
				delete(pendingRequests, requestID)
			}
		}
	}
}

func (s server) run() {
	c := newMqttClient(s.MQTTBrokerURL, s.ID)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	topic := getHTTPResponseTopic("+", "+")
	if token := c.Subscribe(topic, 0, s.httpResponseHandler); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}
	go cleanupRequests()
	portStr := fmt.Sprintf(":%d", s.Port)
	log.Println("Listening on " + portStr)
	http.HandleFunc("/", s.handler)
	err := http.ListenAndServe(portStr, nil)
	if err != nil {
		panic(err)
	}
}
