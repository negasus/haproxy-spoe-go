package worker

import (
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

/*
 * simple test that check for race condition
 * tests need to be run with -race
 */
func TestWorkerConcurrent(t *testing.T) {
	clientConn, server := net.Pipe()
	clientConn2, server2 := net.Pipe()
	spoe := client.NewClient(clientConn)
	spoe2 := client.NewClient(clientConn2)
	var m MockedHandler
	m.m.On("handle", mock.Anything)

	go func() {
		Handle(server, m.Handle)
	}()
	go func() {
		Handle(server2, m.Handle)
	}()
	duration := time.Second
	loop := func(s client.Client) {
		assert.NoError(t, s.Init())
		for {
			select {
			case <-time.After(duration):
				s.Stop()
			default:
				s.Notify()
			}
		}
	}
	go loop(spoe)
	go loop(spoe2)

	// Lets wait a bit to have everything finished
	<-time.After(duration)

	m.m.AssertExpectations(t)

}

/*
 * Simple bench to compare memory usage
 *
 */
func BenchmarkWorker(b *testing.B) {
	clientConn, server := net.Pipe()
	spoe := client.NewClient(clientConn)
	var m MockedHandler
	m.m.On("handle", mock.Anything)
	m.m.On("Finished")

	go func() {
		Handle(server, m.Handle)
		m.Finish()
	}()

	spoe.Init()
	for n := 0; n < b.N; n++ {
		spoe.Notify()
	}
	spoe.Stop()

	// Lets wait a bit to have everything finished
	<-time.After(time.Millisecond * 100)
	clientConn.Close()

}
