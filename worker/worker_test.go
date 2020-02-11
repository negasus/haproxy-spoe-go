package worker

import (
	"fmt"
	"github.com/negasus/haproxy-spoe-go/client"
	"github.com/negasus/haproxy-spoe-go/request"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"testing"
	"time"
)

type MockedHandler struct {
	m mock.Mock
}

func (h *MockedHandler) Handle(r *request.Request) {
	h.m.MethodCalled("handle", r)
	fmt.Println("test")
}

func (h *MockedHandler) Finish() {
	h.m.MethodCalled("Finished")
}

func TestWorker(t *testing.T) {
	clientConn, server := net.Pipe()
	spoe := client.NewClient(clientConn)
	var m MockedHandler
	m.m.On("handle", mock.Anything)
	m.m.On("Finished")

	go func() {
		Handle(server, m.Handle)
		m.Finish()
	}()
	assert.NoError(t, spoe.Init())
	assert.NoError(t, spoe.Notify())
	assert.NoError(t, spoe.Stop())

	// Lets wait a bit to have everything finished
	<-time.After(time.Millisecond * 100)
	clientConn.Close()

	m.m.AssertExpectations(t)

}
