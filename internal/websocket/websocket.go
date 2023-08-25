package websocket

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	"github.com/OliveTin/OliveTin/internal/executor"
	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
)

var upgrader = ws.Upgrader{}

type WebsocketClient struct {
	conn *ws.Conn
}

var clients []*WebsocketClient

var marshalOptions = protojson.MarshalOptions{
	UseProtoNames:   false, // eg: canExec for js instead of can_exec from protobuf
	EmitUnpopulated: true,
}

var ExecutionListener WebsocketExecutionListener

type WebsocketExecutionListener struct{}

func (WebsocketExecutionListener) OnExecutionStarted(title string) {
	/*
		broadcast(ExecutionStarted{
			Type: "ExecutionStarted",
			Action: title,
		});
	*/
}

func (WebsocketExecutionListener) OnExecutionFinished(logEntry *executor.InternalLogEntry) {
	le := &pb.LogEntry{
		ActionTitle:      logEntry.ActionTitle,
		ActionIcon:       logEntry.ActionIcon,
		DatetimeStarted:  logEntry.DatetimeStarted,
		DatetimeFinished: logEntry.DatetimeFinished,
		Stdout:           logEntry.Stdout,
		Stderr:           logEntry.Stderr,
		TimedOut:         logEntry.TimedOut,
		ExitCode:         logEntry.ExitCode,
		Tags:             logEntry.Tags,
		Uuid:             logEntry.UUID,
	}

	broadcast("ExecutionFinished", le)
}

func broadcast(messageType string, pbmsg *pb.LogEntry) {
	payload, err := marshalOptions.Marshal(pbmsg)

	// <EVIL>
	// So, the websocket wants to encode messages using the same protomarshaller
	// as the REST API - this gives consistency instead of using encoding/json
	// and allows us to set specific marshalOptions.
	//
	// However, the protomarshaller will marshal the type, but the JavaScript at
	// the other end has no idea what type this object is - as we're just sending
	// it as JSON over the websocket.
	//
	// Therefore, we wrap the nicely marsheled bytes in a hacky JSON string
	// literal and encode that string just with a byte array cast.
	hackyMessageEnvelope := "{\"type\": \"" + messageType + "\", \"payload\": "

	hackyMessage := []byte{}
	hackyMessage = append(hackyMessage, []byte(hackyMessageEnvelope)...)
	hackyMessage = append(hackyMessage, payload...)
	hackyMessage = append(hackyMessage, []byte("}")...)
	// </EVIL>

	if err != nil {
		log.Errorf("websocket marshal error: %v", err)
		return
	}

	for _, client := range clients {
		client.conn.WriteMessage(1, hackyMessage)
	}
}

func (c *WebsocketClient) messageLoop() {
	for {
		mt, message, err := c.conn.ReadMessage()

		if err != nil {
			log.Printf("err: %v", err)
			break
		}

		log.Infof("websocket recv: %s %d", message, mt)
	}
}

func HandleWebsocket(w http.ResponseWriter, r *http.Request) bool {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Warnf("Websocket issue: %v", err)
		return false
	}

	//	defer c.Close()

	wsclient := &WebsocketClient{
		conn: c,
	}

	clients = append(clients, wsclient)

	go wsclient.messageLoop()

	return true
}
