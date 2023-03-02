# HTTP connector to PrivatBank POS-terminals via ECR protocol (JSON based) by TCP connection

This connector is a service that, on the one hand, detects a terminal and exchanges commands with it according to the
protocol PrivatBank, using a TCP connection for this. On the other hand, it is implemented HTTP server that can receive
/ send JSON requests and broadcast them to the terminal

## Run

`docker run -d -e TERMINAL_IP='192.168.0.111' -p 3333:3333 --rm subsan/http2privat_json:latest`

## Endpoints

### `GET => /check`

Check is TCP-connection up, send PING command to terminal and broadcast response

### `GET => /command` (request json in body)

Send input json (in body) to terminal, waiting next response and broadcast in to response.
On timeout (`TIMEOUT_TRANSACTION`) send to terminal `interrupt` command and response `Transaction timeout` error.
Only one command at a time is supported. For other requests during the execution of the command, the response
is error `Another transaction active`.

## Environments

#### Terminal connection

- TERMINAL_IP - ip address of POS-terminal
- TERMINAL_PORT - port for connect to POS-terminal (default 2000)

#### Web-server

- WEBSERVER_PORT - web listener port (default 3333)

#### Timeouts:

- TIMEOUT_CONNECTION - timeout connection to TCP port to terminal (default 15s)
- TIMEOUT_WRITE - timeout to write to TCP port of terminal (default 15s)
- TIMEOUT_RECONNECT - timeout connection to TCP port to terminal when reconnect (default 5s)
- TIMEOUT_TRANSACTION - timeout of waiting next command (response) at `/command` endpoint (default 10m)

## JSON-protocol POS-terminals capability

- Ingenico (sw version TA7E/TE7E 136 and higher)
- Verifone (sw version 03P and higher)
- PAX S800
- Android terminals A930 (with FacePay24)

## JSON request/response structure

```
Method           string            `json:"method"`
Step             int               `json:"step"`
Params           map[string]string `json:"params"`
Error            bool              `json:"error"`
ErrorDescription string            `json:"errorDescription"`
```

## JSON-protocol specification

https://drive.google.com/drive/u/0/folders/1ySRZ_UVCsy77iSFLf1IrR9zpvEC7SoDW?fbclid=IwAR0s6MUVX_ioMLI36--Y4FwoYXMDe_C-HZl3W9CTNv76pB4AKdN4uinxdaY