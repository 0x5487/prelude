package websocket

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jasonsoft/prelude"
	"github.com/jasonsoft/prelude/hub/nats"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestGateway(t *testing.T) {
	opts := nats.HubOptions{
		URL: "nats://127.0.0.1:4222",
	}
	hub, err := nats.NewNatsHub(opts)
	require.Nil(t, err)

	helloQueue1 := make(chan *prelude.Command)
	err = hub.QueueSubscribe("/hello", "gateway", helloQueue1)
	require.Nil(t, err)

	go func() {
		websocketGateway := NewGateway()
		websocketGateway.ListenAndServe(":10080", hub)
	}()

	time.Sleep(1 * time.Second)

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial("http://localhost:10080", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	cmd := prelude.Command{
		Path: "/hello",
		Data: []byte("hello world"),
	}

	err = ws.WriteJSON(cmd)
	require.Nil(t, err)

	revCmd := <-helloQueue1

	assert.Equal(t, "/hello", revCmd.Path)
	assert.Equal(t, "Hello World", string(revCmd.Data))

}
