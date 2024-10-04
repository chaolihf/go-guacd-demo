package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/anhk/guacamole/guacd"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	parameters["dpi"] = "96"
	parameters["hostname"] = ctx.Query("remote")
	parameters["port"] = ctx.Query("port")
	parameters["resize-method"] = "display-update"
	parameters["disable-audio"] = "true"
	parameters[guacd.ColorDepth] = "24"

	// driver
	if false {
		parameters["enable-drive"] = "true"
		parameters["drive-name"] = "Remote Terminal Mounted Driver"
		parameters["drive-path"] = "/export/FileTransfer/"
	}

	parameters["scheme"] = "rdp" //parameters["hostname"][:3] // TODO: 先使用前3个字母判断下协议
	switch parameters["scheme"] {
	case "rdp":
		parameters["username"] = "lichao"
		parameters["password"] = "lc2013!"
		parameters["ignore-cert"] = "true"
		parameters["security"] = "any"

	case "ssh":
		parameters["username"] = "root"
		if true { // use password
			parameters["password"] = "linuxserver"
		} else { // use privateKey
			parameters["private-key"] = "----"
			parameters["passphrase"] = "----"
		}
		parameters[guacd.FontName] = "Courier New"
		parameters[guacd.FontSize] = "12"
		parameters[guacd.ColorScheme] = "white-black"
	case "vnc":
	}

	// TODO: record
	//parameters["recording-file"]
	//parameters["recording-path"]
	//parameters["recording-name"]

	fmt.Println("parameters:", parameters)
	return parameters["scheme"], parameters
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
	defer func() { _ = ws.Close() }()
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
			} else {
				fmt.Println("msg ws->guacd:", string(msg))
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
			} else if len(data) == 0 {
				// do nothing
			} else if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				done <- errors.New("disconnected")
				break loop
			} else {
				fmt.Println("msg guacd->ws: ", len(data), "==", string(data))
			}
		}
	}
}
