package main

import (
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("mqhttp", "A HTTP request broker run over MQTT")

	clientCmd         = app.Command("client", "Run the client (listens for requests and responds over MQTT)")
	clientMqtt        = clientCmd.Arg("mqtt-host", "IP:port of the MQTT broker").Required().TCP()
	clientID          = clientCmd.Arg("client-id", "Unique client ID").Required().String()
	clientHTTPBackend = clientCmd.Arg("http-backend", "Address of the backend HTTP server to send requests to").Required().TCP()

	serverCmd      = app.Command("server", "Run the server (listens publicly for requests)")
	serverMqtt     = serverCmd.Arg("mqtt-host", "IP:port of the MQTT broker").Required().TCP()
	serverID       = serverCmd.Arg("server-id", "Unique server ID").Required().String()
	serverHTTPPort = serverCmd.Arg("port", "Port to listen for HTTP requests on").Required().Int()
)

func main() {
	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case clientCmd.FullCommand():
		c := client{
			ID:            *clientID,
			MQTTBrokerURL: "tcp://" + (*clientMqtt).String(),
			HTTPBackend:   (*clientHTTPBackend).String(),
		}
		c.run()
	case serverCmd.FullCommand():
		s := server{
			ID:            *serverID,
			MQTTBrokerURL: "tcp://" + (*serverMqtt).String(),
			Port:          *serverHTTPPort,
		}
		s.run()
	}
}
