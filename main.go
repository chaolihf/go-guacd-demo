package main

import (
	"errors"
	"fmt"
	"github.com/anhk/guacamole/guacd"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

func main() {
	router := gin.Default()
	router.GET("/ws", apiConnect)
	if err := router.Run(":9528"); err != nil {
		panic(err)
	}
}

var upgrade = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func guacParameters(ctx *gin.Context) (string, map[string]string) {
	parameters := make(map[string]string)
	parameters["width"] = ctx.Query("width")
	parameters["height"] = ctx.Query("height")
	parameters["dpi"] = "300"
	parameters["hostname"] = ctx.Query("remote")
	parameters["port"] = ctx.Query("port")

	// driver
	if false {
		parameters["enable-drive"] = "true"
		parameters["drive-name"] = "Remote Terminal Mounted Driver"
		parameters["drive-path"] = "/export/FileTransfer/"
	}

	switch parameters["hostname"][:3] { // TODO: 先使用前3个字母判断下协议
	case "rdp":
		parameters["username"] = "ubuntu"
		parameters["password"] = "ubuntu"

	case "ssh":
		parameters["username"] = "root"
		if true { // use password
			parameters["password"] = "linuxserver"
		} else { // use privateKey
			parameters["private-key"] = "----"
			parameters["passphrase"] = "----"
		}
		parameters["font-name"] = "Courier New"
		parameters["font-size"] = "12"
		parameters["color-scheme"] = "gray-black"
	case "vnc":
	}

	// TODO: record
	return parameters["hostname"][:3], parameters
}

func apiConnect(ctx *gin.Context) {
	protocol := ctx.Request.Header.Get("Sec-Websocket-Protocol")
	fmt.Println(protocol) // guacamole

	ws, err := upgrade.Upgrade(ctx.Writer, ctx.Request, http.Header{
		"Sec-Websocket-Protocol": {protocol},
	})
	if err != nil {
		return
	}
	defer ws.Close()
	fmt.Println("## #")

	configuration := guacd.NewConfiguration()
	configuration.Protocol, configuration.Parameters = guacParameters(ctx)

	fmt.Println(configuration)

	guacdTunnel, err := guacd.NewTunnel("127.0.0.1:4822", configuration)
	if err != nil {
		return
	}
	done := make(chan error, 2)
	go proxyTunnelToWebsocket(guacdTunnel, ws, done)

loop:
	for {
		select {
		case <-done:
			break loop
		default:
			if _, msg, err := ws.ReadMessage(); err != nil {
				done <- errors.New("disconnected")
				break loop
			} else if _, err := guacdTunnel.WriteAndFlush(msg); err != nil {
				done <- errors.New("disconnected")
				break loop
			}
		}
	}

	return
}

func proxyTunnelToWebsocket(tunnel *guacd.Tunnel, ws *websocket.Conn, done chan error) {
loop:
	for {
		select {
		case <-done:
			break loop
		default:
			if data, err := tunnel.Read(); err != nil {
				done <- errors.New("disconnected")
				break loop
			} else if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				done <- errors.New("disconnected")
				break loop
			}
		}
	}
}
