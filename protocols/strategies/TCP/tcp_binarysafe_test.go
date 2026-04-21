package TCP

import (
	"net"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/stretchr/testify/assert"
)

// ─── decodeProtocolCommand ────────────────────────────────────────────

func TestDecodeProtocolCommand_RESP2Ping(t *testing.T) {
	frame := []byte("*1\r\n$4\r\nPING\r\n")
	assert.Equal(t, "PING", decodeProtocolCommand(frame))
}

func TestDecodeProtocolCommand_RESP2SetKeyValue(t *testing.T) {
	frame := []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	assert.Equal(t, "SET foo bar", decodeProtocolCommand(frame))
}

func TestDecodeProtocolCommand_RESP2Info(t *testing.T) {
	frame := []byte("*1\r\n$4\r\nINFO\r\n")
	assert.Equal(t, "INFO", decodeProtocolCommand(frame))
}

func TestDecodeProtocolCommand_RESP2ClientList(t *testing.T) {
	frame := []byte("*2\r\n$6\r\nCLIENT\r\n$4\r\nLIST\r\n")
	assert.Equal(t, "CLIENT LIST", decodeProtocolCommand(frame))
}

func TestDecodeProtocolCommand_RESP2Lowercase(t *testing.T) {
	frame := []byte("*1\r\n$4\r\nping\r\n")
	assert.Equal(t, "ping", decodeProtocolCommand(frame))
}

func TestDecodeProtocolCommand_RESP2TruncatedReturnsPrefix(t *testing.T) {
	// SET foo bar with `bar` cut off after one byte.
	frame := []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nb")
	got := decodeProtocolCommand(frame)
	assert.NotEmpty(t, got, "truncated RESP should return a parseable prefix, not nothing")
	assert.True(t, strings.HasPrefix(got, "SET foo"), "expected prefix 'SET foo', got %q", got)
}

func TestDecodeProtocolCommand_RESP2MalformedFallsToHex(t *testing.T) {
	// `*` followed by non-digit.
	frame := []byte("*z\r\nGarbage")
	got := decodeProtocolCommand(frame)
	assert.NotEmpty(t, got, "malformed RESP should still return the hex-escaped form, not empty")
}

func TestDecodeProtocolCommand_RESP2SimpleTypes(t *testing.T) {
	// RESP2 simple string / error / integer / bulk-header — returned CRLF-stripped.
	cases := map[string]string{
		"+OK\r\n":                  "+OK",
		"-ERR unknown command\r\n": "-ERR unknown command",
		":42\r\n":                  ":42",
		"$5\r\nhello\r\n":          "$5\r\nhello", // bulk header + payload; trimmed trailing CRLF
	}
	for in, want := range cases {
		assert.Equal(t, want, decodeProtocolCommand([]byte(in)), "input %q", in)
	}
}

func TestDecodeProtocolCommand_PlainASCII(t *testing.T) {
	assert.Equal(t, "hello world", decodeProtocolCommand([]byte("hello world\r\n")))
}

func TestDecodeProtocolCommand_BinaryFallsToHex(t *testing.T) {
	frame := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	got := decodeProtocolCommand(frame)
	assert.Contains(t, got, `\x00`, "binary input should be hex-escaped")
}

func TestDecodeProtocolCommand_Empty(t *testing.T) {
	assert.Equal(t, "", decodeProtocolCommand([]byte{}))
}

// RESP array with argc=0 is treated as malformed by the RESP decoder.
// The outer decoder falls through to plaintext trimming — that's fine;
// the requirement is no panic and a non-empty capture.
func TestDecodeProtocolCommand_RESP2ZeroArgCount(t *testing.T) {
	got := decodeProtocolCommand([]byte("*0\r\n"))
	assert.NotEmpty(t, got)
}

// RESP array with an absurd argc must not allocate wildly.
func TestDecodeProtocolCommand_RESP2HugeArgCountRejected(t *testing.T) {
	frame := []byte("*999999\r\n$4\r\nPING\r\n")
	got := decodeProtocolCommand(frame)
	// sanity cap rejects the count; decoder falls through to hex.
	assert.NotEqual(t, "PING", got)
}

// RESP bulk-length overrun must not read past len(b). The RESP decoder
// rejects the frame; the outer decoder falls through to plaintext
// trimming, which is fine — we just need no crash and a non-empty capture.
func TestDecodeProtocolCommand_RESP2BulkLenOverrun(t *testing.T) {
	frame := []byte("*1\r\n$9999999\r\n")
	got := decodeProtocolCommand(frame)
	assert.NotEmpty(t, got, "malformed bulk-len should still yield some capture, not empty")
	assert.NotContains(t, got, "PING", "should not spuriously decode as a RESP command")
}

// ─── hexEscapeNonPrintable ────────────────────────────────────────────

func TestHexEscape_PreservesPrintable(t *testing.T) {
	assert.Equal(t, "Hello, World!", hexEscapeNonPrintable([]byte("Hello, World!")))
}

func TestHexEscape_EscapesCRLF(t *testing.T) {
	assert.Equal(t, `a\x0db\x0ac`, hexEscapeNonPrintable([]byte("a\rb\nc")))
}

func TestHexEscape_EscapesNull(t *testing.T) {
	assert.Equal(t, `\x00A\x00`, hexEscapeNonPrintable([]byte{0x00, 0x41, 0x00}))
}

func TestHexEscape_EscapesBackslash(t *testing.T) {
	// Backslash escapes itself so output is unambiguously decodable.
	assert.Equal(t, `a\x5cb`, hexEscapeNonPrintable([]byte(`a\b`)))
}

func TestHexEscape_HighByte(t *testing.T) {
	// Critical case: LDAP BindRequest CONTEXT 0 IMPLICIT tag = 0x80.
	// Must survive as `\x80`, not a UTF-8 replacement character.
	assert.Equal(t, `\x80`, hexEscapeNonPrintable([]byte{0x80}))
}

// ─── isMostlyPrintable ────────────────────────────────────────────────

func TestIsMostlyPrintable(t *testing.T) {
	cases := map[string]bool{
		"hello":                           true,
		"hello\r\nworld":                  true,
		"":                                false, // empty is not "text"
		string([]byte{0, 0, 0, 0}):        false,
		"abcdefghi" + string([]byte{0}):   true,  // 9/10 printable
		"abcdefgh" + string([]byte{0, 0}): false, // 8/10 printable
	}
	for input, want := range cases {
		assert.Equal(t, want, isMostlyPrintable([]byte(input)), "input %q", input)
	}
}

// ─── encodeReply — RESP2 format coverage ─────────────────────────────

func TestEncodeReply_LegacyDefault(t *testing.T) {
	cmd := parser.Command{ReplyFormat: ""}
	assert.Equal(t, []byte("hello\r\n"), encodeReply(cmd, "hello"))
}

func TestEncodeReply_LegacyEmptyValue(t *testing.T) {
	cmd := parser.Command{ReplyFormat: ""}
	assert.Nil(t, encodeReply(cmd, ""), "legacy empty value should suppress reply entirely")
}

func TestEncodeReply_RedisSimple(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-simple"}
	assert.Equal(t, []byte("+PONG\r\n"), encodeReply(cmd, "PONG"))
}

func TestEncodeReply_RedisInteger(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-integer"}
	assert.Equal(t, []byte(":42\r\n"), encodeReply(cmd, "42"))
}

func TestEncodeReply_RedisError(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-error"}
	assert.Equal(t, []byte("-ERR unknown command\r\n"), encodeReply(cmd, "ERR unknown command"))
}

func TestEncodeReply_RedisBulk(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-bulk"}
	value := "redis_version:7.2.4"
	got := encodeReply(cmd, value)
	assert.Equal(t, []byte("$19\r\nredis_version:7.2.4\r\n"), got)
}

func TestEncodeReply_RedisBulkWithEmbeddedCRLF(t *testing.T) {
	// RESP bulk strings are byte-exact: CRLF inside the value is preserved
	// and the length prefix counts every byte.
	cmd := parser.Command{ReplyFormat: "redis-bulk"}
	value := "line1\r\nline2"
	got := encodeReply(cmd, value)
	assert.Equal(t, []byte("$12\r\nline1\r\nline2\r\n"), got, "bulk length must equal byte count of value, CRLF included")
}

func TestEncodeReply_RedisNilBulk(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-nil-bulk"}
	// handler value is intentionally ignored.
	assert.Equal(t, []byte("$-1\r\n"), encodeReply(cmd, "ignored"))
}

func TestEncodeReply_RedisRawVerbatim(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-raw"}
	raw := "*3\r\n:1\r\n:2\r\n:3\r\n"
	assert.Equal(t, []byte(raw), encodeReply(cmd, raw))
}

func TestEncodeReply_RedisArray(t *testing.T) {
	cmd := parser.Command{
		ReplyFormat: "redis-array",
		ReplyBulks:  []string{"get", "set", "ping"},
	}
	got := encodeReply(cmd, "")
	want := []byte("*3\r\n$3\r\nget\r\n$3\r\nset\r\n$4\r\nping\r\n")
	assert.Equal(t, want, got)
}

func TestEncodeReply_RedisArrayEmpty(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-array", ReplyBulks: nil}
	assert.Equal(t, []byte("*0\r\n"), encodeReply(cmd, ""))
}

func TestEncodeReply_UnknownFormatFallsBackToPlaintext(t *testing.T) {
	cmd := parser.Command{ReplyFormat: "redis-unknown-xyz"}
	got := encodeReply(cmd, "hello")
	assert.Equal(t, []byte("hello\r\n"), got)
}

// ─── readBinaryFrame — uses net.Pipe for deterministic I/O ───────────

// pipePair wraps a net.Pipe for server-side / client-side halves with
// deadline support.
func pipePair(t *testing.T) (server, client net.Conn) {
	t.Helper()
	s, c := net.Pipe()
	t.Cleanup(func() { _ = s.Close(); _ = c.Close() })
	return s, c
}

func TestReadBinaryFrame_CollectsSingleWrite(t *testing.T) {
	server, client := pipePair(t)
	frame := []byte("*1\r\n$4\r\nPING\r\n")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = client.Write(frame)
		// Leave the connection open; readBinaryFrame should end via settle window.
	}()

	buf, err := readBinaryFrame(server, 2*time.Second)
	wg.Wait()
	assert.NoError(t, err)
	assert.Equal(t, frame, buf, "single write should round-trip verbatim")
}

func TestReadBinaryFrame_CollectsBurstAcrossWrites(t *testing.T) {
	server, client := pipePair(t)
	part1 := []byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo")
	part2 := []byte("\r\n$3\r\nbar\r\n")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = client.Write(part1)
		// Tight enough that we stay inside the settle window.
		time.Sleep(5 * time.Millisecond)
		_, _ = client.Write(part2)
	}()

	buf, err := readBinaryFrame(server, 2*time.Second)
	wg.Wait()
	assert.NoError(t, err)
	assert.Equal(t, append(part1, part2...), buf, "multi-write burst should merge into one frame")
}

func TestReadBinaryFrame_StopsOnSettleWindow(t *testing.T) {
	// Write one frame, then stop writing. The reader should return promptly
	// (after the settle window), not wait for the session deadline.
	server, client := pipePair(t)
	frame := []byte("*1\r\n$4\r\nPING\r\n")

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		_, _ = client.Write(frame)
	}()

	start := time.Now()
	buf, err := readBinaryFrame(server, 5*time.Second)
	elapsed := time.Since(start)

	<-writerDone
	assert.NoError(t, err)
	assert.Equal(t, frame, buf)
	// 5-second session deadline but settle window should return in ~50 ms.
	// Generous 500 ms upper bound to avoid flakes on loaded CI.
	assert.Less(t, elapsed, 500*time.Millisecond,
		"reader should return via settle window, not session deadline")
}

func TestReadBinaryFrame_ClosedConnectionReturnsError(t *testing.T) {
	// net.Pipe returns its own "closed pipe" error; real net.TCPConn
	// returns io.EOF. Accept either — we just need "no data + error".
	server, client := pipePair(t)
	_ = client.Close()
	buf, err := readBinaryFrame(server, 500*time.Millisecond)
	assert.Error(t, err)
	assert.Empty(t, buf)
}

// Mock tracer for handler test.
type captureTracer struct {
	mu     sync.Mutex
	events []tracer.Event
}

func (c *captureTracer) TraceEvent(e tracer.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
}

func (c *captureTracer) snapshot() []tracer.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]tracer.Event, len(c.events))
	copy(out, c.events)
	return out
}

// ─── handleBinarySafeConnection — full round-trip ────────────────────

func TestHandleBinarySafe_RESPPingReturnsPong(t *testing.T) {
	server, client := pipePair(t)
	cap := &captureTracer{}

	pingRegex := regexp.MustCompile("^(?i)PING$")
	cfg := parser.BeelzebubServiceConfiguration{
		BinarySafe:             true,
		DeadlineTimeoutSeconds: 2,
		Commands: []parser.Command{
			{
				RegexStr:    "^(?i)PING$",
				Regex:       pingRegex,
				Handler:     "PONG",
				ReplyFormat: "redis-simple",
				Name:        "ping",
			},
		},
	}

	done := make(chan struct{})
	go func() {
		handleBinarySafeConnection(server, cfg, cap)
		close(done)
	}()

	// Send one RESP2 PING.
	_, err := client.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	assert.NoError(t, err)

	// Read the reply.
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 64)
	n, _ := client.Read(reply)
	assert.Equal(t, "+PONG\r\n", string(reply[:n]))

	_ = client.Close()
	<-done

	events := cap.snapshot()
	if assert.Len(t, events, 1, "expect one Interaction event per matched command") {
		e := events[0]
		assert.Equal(t, "TCP binary-safe interaction", e.Msg)
		assert.Equal(t, "PING", e.Command)
		assert.Equal(t, `*1\x0d\x0a$4\x0d\x0aPING\x0d\x0a`, e.CommandRaw)
		assert.Equal(t, "PONG", e.CommandOutput)
		assert.Equal(t, "ping", e.Handler)
		assert.Equal(t, "TCP", e.Protocol)
		assert.Equal(t, tracer.Interaction.String(), e.Status)
	}
}

func TestHandleBinarySafe_FallbackOnNoMatch(t *testing.T) {
	server, client := pipePair(t)
	cap := &captureTracer{}

	cfg := parser.BeelzebubServiceConfiguration{
		BinarySafe:             true,
		DeadlineTimeoutSeconds: 2,
		FallbackCommand: parser.Command{
			Handler:     "ERR unknown command",
			ReplyFormat: "redis-error",
			Name:        "catch_all",
		},
	}

	done := make(chan struct{})
	go func() {
		handleBinarySafeConnection(server, cfg, cap)
		close(done)
	}()

	_, _ = client.Write([]byte("*1\r\n$6\r\nFOOBAR\r\n"))

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 64)
	n, _ := client.Read(reply)
	assert.Equal(t, "-ERR unknown command\r\n", string(reply[:n]))

	_ = client.Close()
	<-done

	events := cap.snapshot()
	if assert.Len(t, events, 1) {
		assert.Equal(t, "catch_all", events[0].Handler)
		assert.Equal(t, "FOOBAR", events[0].Command)
	}
}

func TestHandleBinarySafe_MultipleCommandsOnOneConnection(t *testing.T) {
	// Redis clients reuse connections; each frame must be answered in turn.
	server, client := pipePair(t)
	cap := &captureTracer{}

	pingRegex := regexp.MustCompile("^(?i)PING$")
	infoRegex := regexp.MustCompile("^(?i)INFO$")
	cfg := parser.BeelzebubServiceConfiguration{
		BinarySafe:             true,
		DeadlineTimeoutSeconds: 2,
		Commands: []parser.Command{
			{RegexStr: "^(?i)PING$", Regex: pingRegex, Handler: "PONG", ReplyFormat: "redis-simple", Name: "ping"},
			{RegexStr: "^(?i)INFO$", Regex: infoRegex, Handler: "redis_version:7.2.4", ReplyFormat: "redis-bulk", Name: "info"},
		},
	}

	done := make(chan struct{})
	go func() {
		handleBinarySafeConnection(server, cfg, cap)
		close(done)
	}()

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))

	_, _ = client.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	reply1 := make([]byte, 64)
	n1, _ := client.Read(reply1)
	assert.Equal(t, "+PONG\r\n", string(reply1[:n1]))

	_, _ = client.Write([]byte("*1\r\n$4\r\nINFO\r\n"))
	reply2 := make([]byte, 64)
	n2, _ := client.Read(reply2)
	assert.Equal(t, "$19\r\nredis_version:7.2.4\r\n", string(reply2[:n2]))

	_ = client.Close()
	<-done

	events := cap.snapshot()
	if assert.Len(t, events, 2) {
		assert.Equal(t, "PING", events[0].Command)
		assert.Equal(t, "INFO", events[1].Command)
	}
}

func TestHandleBinarySafe_BannerWrittenByteExact(t *testing.T) {
	// Byte-exact banner: no appended newline, no mangling.
	server, client := pipePair(t)
	cap := &captureTracer{}

	cfg := parser.BeelzebubServiceConfiguration{
		BinarySafe:             true,
		DeadlineTimeoutSeconds: 1,
		Banner:                 "\x00\x11\x22\x33raw-banner",
	}

	done := make(chan struct{})
	go func() {
		handleBinarySafeConnection(server, cfg, cap)
		close(done)
	}()

	_ = client.SetReadDeadline(time.Now().Add(1 * time.Second))
	got := make([]byte, 64)
	n, _ := client.Read(got)
	assert.Equal(t, []byte("\x00\x11\x22\x33raw-banner"), got[:n])

	_ = client.Close()
	<-done
}

// ─── bytesIndexCRLF helpers ──────────────────────────────────────────

func TestBytesIndexCRLF(t *testing.T) {
	assert.Equal(t, 3, bytesIndexCRLF([]byte("abc\r\ndef")))
	assert.Equal(t, -1, bytesIndexCRLF([]byte("no-crlf-here")))
	assert.Equal(t, -1, bytesIndexCRLF([]byte("")))
	assert.Equal(t, 0, bytesIndexCRLF([]byte("\r\nabc")))
}

func TestBytesIndexCRLFFrom(t *testing.T) {
	b := []byte("abc\r\ndef\r\nghi")
	assert.Equal(t, 3, bytesIndexCRLFFrom(b, 0))
	assert.Equal(t, 8, bytesIndexCRLFFrom(b, 5))
	assert.Equal(t, -1, bytesIndexCRLFFrom(b, 10))
	assert.Equal(t, 3, bytesIndexCRLFFrom(b, -1), "negative offset should be treated as 0")
}
