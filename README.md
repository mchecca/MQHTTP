# MQHTTP

Make HTTP requests via MQTT

## Building

Building is fairly simple, just get the dependencies then run go build

```bash
$ go get -d
$ go build
```

## Running

The basic use case of MQHTTP is to set up a proxy server that listens for HTTP requests ("the server"), then forwards the request via MQTT to another server ("the client") which executes the request on a predefined host, then returns the HTTP response back to the proxy server which delivers it to a requesting client. All requests must have `MQHTTP-Client-ID` header to specify which server the request will go to. This setup is mostly useful as a way to get around a firewall or make a request between two machines that do not have direct network access.

For example, consider a setup with the following components:

* An embedded device that does not have direct internet access, but is connected to a local network (hostname: `embedded-device.local`)
* A server that is on the same local network as the embedded device and can access the internet (hostname: `server.local`)
* A publicly available server that runs a HTTP REST API and a MQTT broker, and is used to send commands to the embedded device (hostname: `public-server.com`)

### Server (`server.local`)
```bash
$ ./MQHTTP server public-server.com:1883 mqhttp-test-server 8080
```

### Client (`public-server.com`)
```bash
$ ./MQHTTP client public-server.com:1883 mqhttp-test-client public-server.com:80
```

### Embedded Device (`embedded-device.local`)
```bash
http GET http://server.local:8080/v1/available_commands 'MQHTTP-Client-ID: http-test-client'
```

### Sample output

```bash
# From embedded-device.local
$ http GET http://server.local:8080/v1/available_commands 'MQHTTP-Client-ID: http-test-client'

# From server.local
2019/10/13 21:14:38 Listening on :8080
2019/10/13 21:14:45 8b460e18-65d4-43c0-a36a-191f694da421 GET /

# From public-server.com
2019/10/13 21:14:40 Client mqhttp-test-client listening on tcp://public-server.com:1883
2019/10/13 21:14:40 Client mqhttp-test-client Using backend HTTP public-server.com:80
2019/10/13 21:14:45 Handling request id: 8b460e18-65d4-43c0-a36a-191f694da421
```