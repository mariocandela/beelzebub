package TELNET

import (
	"io"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/beelzebub-labs/beelzebub/v3/internal/historystore"
	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
	"github.com/stretchr/testify/assert"
)

type mockTracer struct {
	events []tracer.Event
}

func (m *mockTracer) TraceEvent(event tracer.Event) {
	m.events = append(m.events, event)
}

func newTelnetStrategy() *TelnetStrategy {
	return &TelnetStrategy{Sessions: historystore.NewHistoryStore()}
}

// drain reads from conn until deadline expires, discarding data.
func drain(conn net.Conn, timeout time.Duration) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 512)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
	}
	conn.SetReadDeadline(time.Time{})
}

func TestBuildPrompt(t *testing.T) {
	result := buildPrompt("admin", "server01")
	assert.Equal(t, "admin@server01:~$ ", result)
}

func TestBuildPrompt_EmptyUser(t *testing.T) {
	result := buildPrompt("", "myserver")
	assert.Equal(t, "@myserver:~$ ", result)
}

func TestBuildPrompt_EmptyServer(t *testing.T) {
	result := buildPrompt("user", "")
	assert.Equal(t, "user@:~$ ", result)
}

func TestNegotiateTelnet(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		negotiateTelnet(server)
	}()

	// Send some bytes (client side), negotiateTelnet just drains them
	client.Write([]byte{IAC, WILL, ECHO, IAC, DO, SUPPRESS_GO_AHEAD})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("negotiateTelnet did not complete")
	}
}

func TestNegotiateTelnet_Empty(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		negotiateTelnet(server)
	}()

	// Close client side so negotiateTelnet gets an error after its 100ms deadline
	client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("negotiateTelnet did not complete with empty input")
	}
}

func TestReadLine_Simple(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	var result string
	var err error
	go func() {
		defer close(done)
		result, err = readLine(server)
	}()

	client.Write([]byte("hello\n"))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readLine timeout")
	}

	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestReadLine_WithCarriageReturn(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	var result string
	var err error
	go func() {
		defer close(done)
		result, err = readLine(server)
	}()

	client.Write([]byte("test\r\n"))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readLine timeout")
	}

	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestReadLine_SkipsIACSequences(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	var result string
	var err error
	go func() {
		defer close(done)
		result, err = readLine(server)
	}()

	// IAC DO ECHO then "hi\n" — IAC sequence should be stripped
	go func() {
		client.Write([]byte{IAC, DO, ECHO})
		client.Write([]byte("hi\n"))
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readLine timeout")
	}

	assert.NoError(t, err)
	assert.Equal(t, "hi", result)
}

func TestReadLine_ConnectionClose(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	done := make(chan struct{})
	var err error
	go func() {
		defer close(done)
		_, err = readLine(server)
	}()

	client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readLine timeout on connection close")
	}

	assert.Error(t, err)
}

func TestTelnetStrategy_Init(t *testing.T) {
	strategy := &TelnetStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:                "127.0.0.1:0",
		Description:            "test",
		DeadlineTimeoutSeconds: 2,
		PasswordRegex:          ".*",
	}

	err := strategy.Init(servConf, mt)
	assert.NoError(t, err)
	assert.NotNil(t, strategy.Sessions)
}

// doTelnetAuth performs the username/password exchange over client,
// consuming all server prompts and negotiation bytes.
func doTelnetAuth(client net.Conn, username, password string) {
	// negotiateTelnet on server drains whatever the client sends in 100ms.
	// Give it a moment then start reading.
	time.Sleep(150 * time.Millisecond)

	// Drain the "\r\nlogin: " prompt
	drain(client, 300*time.Millisecond)

	// Send username
	client.Write([]byte(username + "\n"))

	// Drain IAC WILL ECHO + "Password: "
	drain(client, 300*time.Millisecond)

	// Send password
	client.Write([]byte(password + "\n"))

	// Drain IAC WONT ECHO + "\r\n"
	drain(client, 300*time.Millisecond)
}

func TestHandleTelnetConnection_InvalidPassword(t *testing.T) {
	client, server := net.Pipe()

	mt := &mockTracer{}
	servConf := parser.BeelzebubServiceConfiguration{
		Description:            "test",
		DeadlineTimeoutSeconds: 10,
		PasswordRegex:          "^correct$",
	}
	strategy := newTelnetStrategy()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handleTelnetConnection(server, servConf, mt, strategy)
	}()

	doTelnetAuth(client, "admin", "wrongpass")

	// After wrong password the server closes connection
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	io.ReadAll(client) // drain remaining bytes
	client.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	assert.GreaterOrEqual(t, len(mt.events), 1)
	found := false
	for _, e := range mt.events {
		if e.Status == tracer.Stateless.String() && e.User == "admin" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected stateless auth trace event")
}

func TestHandleTelnetConnection_ExitCommand(t *testing.T) {
	client, server := net.Pipe()

	mt := &mockTracer{}
	servConf := parser.BeelzebubServiceConfiguration{
		Description:            "test",
		DeadlineTimeoutSeconds: 10,
		PasswordRegex:          ".*",
		ServerName:             "testserver",
		Commands:               []parser.Command{},
	}
	strategy := newTelnetStrategy()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handleTelnetConnection(server, servConf, mt, strategy)
	}()

	doTelnetAuth(client, "user", "pass")

	// Drain shell prompt
	drain(client, 300*time.Millisecond)

	// Send exit command
	client.Write([]byte("exit\n"))

	// Drain any remaining output
	client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	io.ReadAll(client)
	client.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for session to end")
	}

	hasStart := false
	hasEnd := false
	for _, e := range mt.events {
		if e.Status == tracer.Start.String() {
			hasStart = true
		}
		if e.Status == tracer.End.String() {
			hasEnd = true
		}
	}
	assert.True(t, hasStart, "expected session start event")
	assert.True(t, hasEnd, "expected session end event")
}

func TestHandleTelnetConnection_MatchingCommand(t *testing.T) {
	client, server := net.Pipe()

	mt := &mockTracer{}
	servConf := parser.BeelzebubServiceConfiguration{
		Description:            "test",
		DeadlineTimeoutSeconds: 10,
		PasswordRegex:          ".*",
		ServerName:             "testserver",
		Commands: []parser.Command{
			{
				Name:    "ls-handler",
				Regex:   regexp.MustCompile(`^ls$`),
				Handler: "file.txt\nfolder/",
			},
		},
	}
	strategy := newTelnetStrategy()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handleTelnetConnection(server, servConf, mt, strategy)
	}()

	doTelnetAuth(client, "user", "pass")

	// Drain shell prompt
	drain(client, 300*time.Millisecond)

	// Send matching command
	client.Write([]byte("ls\n"))

	// Read the response
	buf := make([]byte, 512)
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := client.Read(buf)
	assert.Contains(t, string(buf[:n]), "file.txt")

	client.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	found := false
	for _, e := range mt.events {
		if e.Status == tracer.Interaction.String() && e.Handler == "ls-handler" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected interaction event with ls-handler")
}

func TestHandleTelnetConnection_UnmatchedCommand(t *testing.T) {
	client, server := net.Pipe()

	mt := &mockTracer{}
	servConf := parser.BeelzebubServiceConfiguration{
		Description:            "test",
		DeadlineTimeoutSeconds: 10,
		PasswordRegex:          ".*",
		ServerName:             "testserver",
		Commands: []parser.Command{
			{
				Regex:   regexp.MustCompile(`^ls$`),
				Handler: "file.txt",
			},
		},
	}
	strategy := newTelnetStrategy()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handleTelnetConnection(server, servConf, mt, strategy)
	}()

	doTelnetAuth(client, "user", "pass")
	drain(client, 300*time.Millisecond)

	// Send unmatched command
	client.Write([]byte("unknown_command\n"))

	// Read "command not found" response
	buf := make([]byte, 512)
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := client.Read(buf)
	assert.Contains(t, string(buf[:n]), "command not found")

	client.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	found := false
	for _, e := range mt.events {
		if e.Handler == "not_found" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected not_found handler event")
}
