package worker

import (
	"net"
	"testing"
	"time"

	"github.com/negasus/haproxy-spoe-go/client"
	"github.com/negasus/haproxy-spoe-go/logger"
	"github.com/negasus/haproxy-spoe-go/request"
)

type MockedHandler struct {
	handleFunc func(r *request.Request)
	finishFunc func()
}

func (h *MockedHandler) Handle(r *request.Request) {
	h.handleFunc(r)
}

func (h *MockedHandler) Finish() {
	h.finishFunc()
}

func TestWorker(t *testing.T) {
	clientConn, server := net.Pipe()
	spoe := client.NewClient(clientConn)
	m := MockedHandler{
		handleFunc: func(r *request.Request) {

		},
		finishFunc: func() {

		},
	}

	go func() {
		Handle(server, m.Handle, logger.NewNop())
		m.Finish()
	}()
	if spoe.Init() != nil {
		t.Fatal("unexpected error on Init")
	}
	if spoe.Notify() != nil {
		t.Fatal("unexpected error on Notify")
	}
	if spoe.Stop() != nil {
		t.Fatal("unexpected error on Stop")
	}

	// Let's wait a bit to have everything finished
	<-time.After(time.Millisecond * 100)
	clientConn.Close()
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
	m := MockedHandler{
		handleFunc: func(r *request.Request) {

		},
		finishFunc: func() {

		},
	}

	go func() {
		Handle(server, m.Handle, logger.NewNop())
	}()
	go func() {
		Handle(server2, m.Handle, logger.NewNop())
	}()
	duration := time.Second
	loop := func(s client.Client) {
		if s.Init() != nil {
			t.Fatal("unexpected error on Init")
		}
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

	// Let's wait a bit to have everything finished
	<-time.After(duration)
}

/*
 * Simple bench to compare memory usage
 *
 */
func BenchmarkWorker(b *testing.B) {
	clientConn, server := net.Pipe()
	spoe := client.NewClient(clientConn)
	m := MockedHandler{
		handleFunc: func(r *request.Request) {

		},
		finishFunc: func() {

		},
	}

	go func() {
		Handle(server, m.Handle, logger.NewNop())
		m.Finish()
	}()

	spoe.Init()
	for n := 0; n < b.N; n++ {
		spoe.Notify()
	}
	spoe.Stop()

	// Let's wait a bit to have everything finished
	<-time.After(time.Millisecond * 100)
	clientConn.Close()
}
