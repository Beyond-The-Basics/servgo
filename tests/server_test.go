package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Mohamed-Moumni/servgo/internal/server"
)

// getFreePort asks the OS for an available TCP port, releases it, and returns
// it.  There is a small TOCTOU window, which is acceptable for test use.
func getFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("getFreePort: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// dialWithRetry attempts to connect to addr up to maxAttempts times, waiting
// delay between each attempt.  It returns the established connection or fails
// the test.
func dialWithRetry(t *testing.T, addr string, maxAttempts int, delay time.Duration) net.Conn {
	t.Helper()
	var (
		conn net.Conn
		err  error
	)
	for i := 0; i < maxAttempts; i++ {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			return conn
		}
		time.Sleep(delay)
	}
	t.Fatalf("dialWithRetry: could not connect to %s after %d attempts: %v", addr, maxAttempts, err)
	return nil
}

// startServer launches CreateServer in a goroutine and returns the error
// channel.  The caller is responsible for triggering shutdown (e.g. by
// occupying the port first to produce an error, or by terminating the process).
func startServer(s *server.Server) <-chan error {
	errCh := make(chan error, 1)
	go func() { errCh <- s.CreateServer() }()
	return errCh
}

// ---- New -----------------------------------------------------------------

func TestNew_ReturnsNonNilServer(t *testing.T) {
	s := server.New(9000)
	if s == nil {
		t.Fatal("New returned nil")
	}
}

// ---- CreateServer --------------------------------------------------------

func TestCreateServer_ReturnsErrorForInvalidPort(t *testing.T) {
	s := server.New(-1)
	err := s.CreateServer()
	if err == nil {
		t.Fatal("expected an error for port -1, got nil")
	}
}

func TestCreateServer_ReturnsErrorWhenPortAlreadyInUse(t *testing.T) {
	port := getFreePort(t)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Skipf("could not bind port %d for setup: %v", port, err)
	}
	defer l.Close()

	s := server.New(port)
	err = s.CreateServer()
	if err == nil {
		t.Fatal("expected an error when the port is already in use, got nil")
	}
}

func TestCreateServer_ListensOnAssignedPort(t *testing.T) {
	port := getFreePort(t)
	s := server.New(port)

	startServer(s)

	addr := fmt.Sprintf(":%d", port)
	conn := dialWithRetry(t, addr, 20, 10*time.Millisecond)
	conn.Close()
}

func TestCreateServer_AcceptsMultipleSequentialConnections(t *testing.T) {
	port := getFreePort(t)
	s := server.New(port)

	startServer(s)

	addr := fmt.Sprintf(":%d", port)

	for i := 0; i < 3; i++ {
		conn := dialWithRetry(t, addr, 20, 10*time.Millisecond)
		conn.Close()
		// Brief pause to let the server recycle the handler goroutine.
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCreateServer_AcceptsConcurrentConnections(t *testing.T) {
	port := getFreePort(t)
	s := server.New(port)

	startServer(s)

	addr := fmt.Sprintf(":%d", port)
	// Wait for the listener to be ready before launching goroutines.
	probe := dialWithRetry(t, addr, 20, 10*time.Millisecond)
	probe.Close()

	const numClients = 5
	errCh := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		go func() {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				errCh <- err
				return
			}
			conn.Close()
			errCh <- nil
		}()
	}

	for i := 0; i < numClients; i++ {
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("concurrent dial failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for concurrent connection")
		}
	}
}

// ---- handleConnection (tested indirectly through CreateServer) -----------

// TestServer_DoesNotCrashOnDataReceived verifies that data sent by a connected
// client is processed without the server goroutine panicking or closing the
// connection prematurely.
func TestServer_DoesNotCrashOnDataReceived(t *testing.T) {
	port := getFreePort(t)
	s := server.New(port)

	startServer(s)

	addr := fmt.Sprintf(":%d", port)
	conn := dialWithRetry(t, addr, 20, 10*time.Millisecond)
	defer conn.Close()

	payloads := []string{"PING", "GET key\r\n", "SET key value\r\n"}
	for _, p := range payloads {
		if _, err := fmt.Fprint(conn, p); err != nil {
			t.Fatalf("Write(%q): %v", p, err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Server must still be reachable — a second connection confirms it did not crash.
	probe := dialWithRetry(t, addr, 5, 10*time.Millisecond)
	probe.Close()
}

// TestServer_HandlesClientDisconnect verifies that the server remains
// operational after a client disconnects abruptly.
func TestServer_HandlesClientDisconnect(t *testing.T) {
	port := getFreePort(t)
	s := server.New(port)

	startServer(s)

	addr := fmt.Sprintf(":%d", port)

	// Connect, send a payload, then abruptly close.
	conn := dialWithRetry(t, addr, 20, 10*time.Millisecond)
	fmt.Fprint(conn, "HELLO")
	conn.Close()

	time.Sleep(50 * time.Millisecond)

	// The server must still accept new connections after the disconnect.
	probe := dialWithRetry(t, addr, 10, 10*time.Millisecond)
	probe.Close()
}
