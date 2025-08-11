package worker

import (
	"bufio"
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/negasus/haproxy-spoe-go/client"
	"github.com/negasus/haproxy-spoe-go/frame"
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

// --- New test to reproduce in-flight Notify dropped on Disconnect ---

func sendFrame(t *testing.T, c net.Conn, f *frame.Frame) {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	if _, err := f.Encode(buf); err != nil {
		t.Fatalf("encode frame: %v", err)
	}
	if _, err := c.Write(buf.Bytes()); err != nil {
		t.Fatalf("write frame: %v", err)
	}
}

func TestWorkerNotifyInFlightOnDisconnect(t *testing.T) {
	clientConn, server := net.Pipe()
	defer clientConn.Close()

	processedCh := make(chan struct{}, 1)

	// Slow handler keeps notify goroutine busy while we send Disconnect.
	handler := func(r *request.Request) {
		time.Sleep(200 * time.Millisecond)
		processedCh <- struct{}{}
	}

	go Handle(server, handler, logger.NewNop())

	reader := bufio.NewReader(clientConn)

	// 1) HaproxyHello -> expect AgentHello
	hello := frame.AcquireFrame()
	hello.Type = frame.TypeHaproxyHello
	hello.StreamID = 0
	hello.FrameID = 0
	hello.KV.Add("supported-versions", "2")
	hello.KV.Add("max-frame-size", uint32(16*1024))
	hello.KV.Add("capabilities", "")
	sendFrame(t, clientConn, hello)
	frame.ReleaseFrame(hello)

	resp := frame.AcquireFrame()
	// frame.Read does not parse AgentHello payload and returns an error; ignore it.
	_ = resp.Read(reader)
	if resp.Type != frame.TypeAgentHello {
		t.Fatalf("unexpected response type: got %v, want AgentHello", resp.Type)
	}
	frame.ReleaseFrame(resp)

	// 2) Send Notify that will be slow to process.
	notify := frame.AcquireFrame()
	notify.Type = frame.TypeNotify
	notify.StreamID = 1
	notify.FrameID = 1
	sendFrame(t, clientConn, notify)
	frame.ReleaseFrame(notify)

	// 3) Immediately send HaproxyDisconnect while notify is in-flight.
	disc := frame.AcquireFrame()
	disc.Type = frame.TypeHaproxyDisconnect
	disc.StreamID = 0
	disc.FrameID = 0
	disc.KV.Add("status-code", uint32(0))
	disc.KV.Add("message", "normal")
	sendFrame(t, clientConn, disc)
	frame.ReleaseFrame(disc)

	// 4) Read AgentAck (for notify) and AgentDisconnect (order may vary).
	gotAck := false
	gotDisconnect := false

	_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for i := 0; i < 2; i++ {
		f := frame.AcquireFrame()
		// Ignore errors: frame.Read returns an error for Agent* frames after consuming the payload.
		_ = f.Read(reader)
		switch f.Type {
		case frame.TypeAgentAck:
			gotAck = true
		case frame.TypeAgentDisconnect:
			gotDisconnect = true
		}
		frame.ReleaseFrame(f)
		if gotAck && gotDisconnect {
			break
		}
	}

	// Ensure the handler actually ran.
	select {
	case <-processedCh:
	case <-time.After(1 * time.Second):
		t.Fatal("notify handler did not run")
	}

	// With current code, this likely fails because run() closes conn
	// before the notify goroutine writes the ACK.
	if !gotAck {
		t.Fatalf("expected AgentAck for in-flight Notify, but none was received (Disconnect received: %v)", gotDisconnect)
	}
}
