package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type client struct {
	ID            string
	MQTTBrokerURL string
	HTTPBackend   string
}

func (c client) httpRequestHandler(client mqtt.Client, msg mqtt.Message) {
	topicSplit := strings.Split(msg.Topic(), "/")
	requestID := topicSplit[3]
	if len(requestID) == 0 {
		log.Println("Invalid topic: " + msg.Topic())
		return
	}
	topic := getHTTPResponseTopic(c.ID, requestID)
	log.Printf("Handling request id: %s\n", requestID)
	conn, err := net.Dial("tcp", c.HTTPBackend)
	if err != nil {
		log.Printf("Unable to connect to URL %s", c.HTTPBackend)
		body := "Bad Gateway\n"
		resp := http.Response{
			StatusCode: 502,
			ProtoMajor: 1,
			ProtoMinor: 0,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		}
		respBuffer := bytes.NewBuffer(nil)
		resp.Write(respBuffer)
		rawResp, _ := ioutil.ReadAll(respBuffer)
		sendMqttMessage(c.MQTTBrokerURL, topic, rawResp)
		return
	}
	defer conn.Close()
	_, err = conn.Write(msg.Payload())
	if err != nil {
		panic(err)
	}
	rawResponse, _ := ioutil.ReadAll(conn)
	sendMqttMessage(c.MQTTBrokerURL, topic, rawResponse)
}

func (c client) run() {
	log.Printf("Client %s listening on %s\n", c.ID, c.MQTTBrokerURL)
	log.Printf("Client %s Using backend HTTP %s\n", c.ID, c.HTTPBackend)
	conn := newMqttClient(c.MQTTBrokerURL, c.ID)
	if token := conn.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	topic := getHTTPRequestTopic(c.ID, "+")
	if token := conn.Subscribe(topic, 0, c.httpRequestHandler); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}
	// Wait for messages
	for {
		time.Sleep(5 * time.Second)
	}
	// c.Disconnect(250)
}
