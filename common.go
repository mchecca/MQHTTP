package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const httpRequestTopicFmt string = "http/%s/request/%s"
const httpResponseTopicFmt string = "http/%s/response/%s"
const requestTimeout = time.Second * 30

func getHTTPRequestTopic(id string, requestID string) string {
	return fmt.Sprintf(httpRequestTopicFmt, id, requestID)
}

func getHTTPResponseTopic(id string, requestID string) string {
	return fmt.Sprintf(httpResponseTopicFmt, id, requestID)
}

func sendMqttMessage(mqttBrokerURL string, topic string, msg []byte) {
	opts := mqtt.NewClientOptions().AddBroker(mqttBrokerURL)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	token := c.Publish(topic, 0, false, msg)
	token.Wait()
}

func newMqttClient(mqttBrokerURL string, clientID string) mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker(mqttBrokerURL).SetClientID(clientID)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	c := mqtt.NewClient(opts)
	return c
}
