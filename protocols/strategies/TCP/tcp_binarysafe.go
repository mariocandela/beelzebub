// Binary-safe interactive TCP handler. Opts in via `binarySafe: true` on
// the service config. Preserves every byte on the wire so Redis RESP2
// (`*1\r\n$4\r\nPING\r\n`) and similar multi-line binary protocols can
// be matched and answered without going through a line-mangling read.
//
// RESP2 encodings follow https://redis.io/docs/latest/develop/reference/protocol-spec/.

package TCP

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
)

// Upper bound on a single captured frame, in bytes. Protects against
// unbounded reads from misbehaving clients.
const binarySafeMaxFrame = 8192

// Short read deadline applied after the first byte arrives. A pause
// longer than this ends the current frame. 50 ms cleanly separates
// back-to-back RESP2 requests on one connection while remaining
// imperceptible to legitimate clients.
const binarySafeSettleWindow = 50 * time.Millisecond

// handleBinarySafeConnection runs the binary-safe request loop on one
// accepted connection. The caller sets the initial deadline and owns
// the conn lifecycle (close).
func handleBinarySafeConnection(conn net.Conn, servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) {
	// Byte-exact banner. No appended newline: operators include one if they want one.
	if servConf.Banner != "" {
		_, _ = conn.Write([]byte(servConf.Banner))
	}

	sessionDeadline := time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second
	if sessionDeadline <= 0 {
		sessionDeadline = 30 * time.Second
	}

	host, port, _ := net.SplitHostPort(conn.RemoteAddr().String())

	for {
		rawBytes, err := readBinaryFrame(conn, sessionDeadline)
		if err != nil && len(rawBytes) == 0 {
			return
		}
		if len(rawBytes) == 0 {
			return
		}

		// Refresh the session-level deadline so long interactive sessions aren't cut mid-stream.
		_ = conn.SetDeadline(time.Now().Add(sessionDeadline))

		commandRawHex := hexEscapeNonPrintable(rawBytes)
		commandInput := decodeProtocolCommand(rawBytes)
		if commandInput == "" {
			continue
		}

		matchedCommand, commandOutput, handlerName := matchCommand(servConf, commandInput)

		wireBytes := encodeReply(matchedCommand, commandOutput)
		if len(wireBytes) > 0 {
			if _, err := conn.Write(wireBytes); err != nil {
				return
			}
		}

		tr.TraceEvent(tracer.Event{
			Msg:           "TCP binary-safe interaction",
			Protocol:      tracer.TCP.String(),
			Status:        tracer.Interaction.String(),
			Command:       commandInput,
			CommandRaw:    commandRawHex,
			CommandOutput: commandOutput,
			RemoteAddr:    conn.RemoteAddr().String(),
			SourceIp:      host,
			SourcePort:    port,
			ID:            uuid.New().String(),
			Description:   servConf.Description,
			Handler:       handlerName,
		})
	}
}

// matchCommand finds the first Command whose Regex matches commandInput,
// falling back to servConf.FallbackCommand when nothing matches. Returns
// the selected Command, its handler string, and a human-readable name
// for the trace event.
func matchCommand(servConf parser.BeelzebubServiceConfiguration, commandInput string) (parser.Command, string, string) {
	for _, command := range servConf.Commands {
		if command.Regex == nil || !command.Regex.MatchString(commandInput) {
			continue
		}
		name := command.Name
		if name == "" {
			name = "configured_regex"
		}
		return command, command.Handler, name
	}

	fb := servConf.FallbackCommand
	name := "not_found"
	if fb.Name != "" {
		name = fb.Name
	} else if fb.Handler != "" || fb.ReplyFormat != "" {
		name = "fallback"
	}
	return fb, fb.Handler, name
}

// readBinaryFrame reads one logical request frame from conn. It returns
// when any of the following is true, whichever comes first:
//
//   - the client stops sending for binarySafeSettleWindow (end of frame);
//   - binarySafeMaxFrame bytes have been collected;
//   - the connection closes;
//   - sessionDeadline expires before any byte arrives.
//
// Bytes are never converted through string — CRLF and non-printable
// bytes survive intact.
func readBinaryFrame(conn net.Conn, sessionDeadline time.Duration) ([]byte, error) {
	buf := make([]byte, 0, 256)
	chunk := make([]byte, 256)

	// Wait the full session deadline for the first byte.
	if err := conn.SetReadDeadline(time.Now().Add(sessionDeadline)); err != nil {
		return buf, err
	}
	n, err := conn.Read(chunk)
	if n > 0 {
		buf = append(buf, chunk[:n]...)
	}
	if err != nil {
		if err == io.EOF {
			return buf, err
		}
		if len(buf) == 0 {
			return buf, err
		}
	}

	// Short reads until the client pauses or the frame fills.
	for len(buf) < binarySafeMaxFrame {
		if err := conn.SetReadDeadline(time.Now().Add(binarySafeSettleWindow)); err != nil {
			break
		}
		n, err := conn.Read(chunk)
		if n > 0 {
			remaining := binarySafeMaxFrame - len(buf)
			if n > remaining {
				n = remaining
			}
			buf = append(buf, chunk[:n]...)
		}
		if err != nil {
			// Settle timeout ends the frame; not a propagated error.
			break
		}
	}

	// Restore the session deadline for subsequent I/O.
	_ = conn.SetReadDeadline(time.Now().Add(sessionDeadline))
	return buf, nil
}

// decodeProtocolCommand returns a regex-friendly logical command string
// extracted from raw frame bytes. It auto-detects the wire format from
// the first bytes:
//
//   - `*N\r\n$L\r\nCMD\r\n...`                          RESP2 array        -> "CMD ARG1 ARG2"
//   - `+OK\r\n` / `-ERR x\r\n` / `:42\r\n` / `$N...`    RESP2 simple types -> returned CRLF-stripped
//   - Mostly-printable ASCII                            telnet-like        -> CRLF-stripped
//   - Binary / unknown                                  fallthrough        -> hex-escaped
//
// Never returns an empty string for non-empty input; never panics.
func decodeProtocolCommand(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	if b[0] == '*' {
		if cmd := decodeRESPArray(b); cmd != "" {
			return cmd
		}
		// Malformed RESP array — fall through to hex.
	}

	if b[0] == '+' || b[0] == '-' || b[0] == ':' || b[0] == '$' {
		return strings.TrimRight(string(b), "\r\n\x00")
	}

	if isMostlyPrintable(b) {
		return strings.TrimRight(string(b), "\r\n\x00 \t")
	}

	return hexEscapeNonPrintable(b)
}

// decodeRESPArray parses a RESP2 array frame per
// https://redis.io/docs/latest/develop/reference/protocol-spec/ and
// returns a space-joined "CMD ARG1 ARG2 ...". An empty string means
// malformed or truncated with no arguments recovered; a partial input
// that parsed at least one argument returns that prefix.
func decodeRESPArray(b []byte) string {
	if len(b) < 4 || b[0] != '*' {
		return ""
	}
	end := bytesIndexCRLF(b)
	if end < 2 {
		return ""
	}

	count := 0
	for _, c := range b[1:end] {
		if c < '0' || c > '9' {
			return ""
		}
		count = count*10 + int(c-'0')
	}
	// Sanity cap. Real Redis commands rarely exceed a handful of arguments.
	if count <= 0 || count > 64 {
		return ""
	}

	pos := end + 2
	parts := make([]string, 0, count)

	for i := 0; i < count; i++ {
		if pos >= len(b) || b[pos] != '$' {
			return ""
		}
		hdrEnd := bytesIndexCRLFFrom(b, pos)
		if hdrEnd < 0 {
			return ""
		}

		bulkLen := 0
		for _, c := range b[pos+1 : hdrEnd] {
			if c < '0' || c > '9' {
				return ""
			}
			bulkLen = bulkLen*10 + int(c-'0')
		}
		if bulkLen < 0 || bulkLen > binarySafeMaxFrame {
			return ""
		}

		valStart := hdrEnd + 2
		valEnd := valStart + bulkLen
		if valEnd > len(b) {
			if len(parts) > 0 {
				return strings.Join(parts, " ")
			}
			return ""
		}

		parts = append(parts, string(b[valStart:valEnd]))
		pos = valEnd + 2
	}

	return strings.Join(parts, " ")
}

// bytesIndexCRLF returns the index of the first `\r\n` in b, or -1.
func bytesIndexCRLF(b []byte) int {
	for i := 0; i < len(b)-1; i++ {
		if b[i] == '\r' && b[i+1] == '\n' {
			return i
		}
	}
	return -1
}

// bytesIndexCRLFFrom returns the index of the first `\r\n` at or after
// from, or -1.
func bytesIndexCRLFFrom(b []byte, from int) int {
	if from < 0 {
		from = 0
	}
	for i := from; i < len(b)-1; i++ {
		if b[i] == '\r' && b[i+1] == '\n' {
			return i
		}
	}
	return -1
}

// isMostlyPrintable reports whether at least 90% of b is printable
// ASCII or whitespace. Empty input returns false.
func isMostlyPrintable(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	printable := 0
	for _, c := range b {
		if (c >= 32 && c <= 126) || c == '\t' || c == '\n' || c == '\r' {
			printable++
		}
	}
	return printable*10 >= len(b)*9
}

// hexEscapeNonPrintable returns a printable-ASCII rendering of b. Bytes
// outside the printable ASCII range are emitted as `\xNN`. The
// backslash itself is escaped for unambiguous decoding. Safe for JSON,
// database TEXT columns, and log lines.
func hexEscapeNonPrintable(b []byte) string {
	var sb strings.Builder
	sb.Grow(len(b))
	for _, c := range b {
		if c >= 32 && c <= 126 && c != '\\' {
			sb.WriteByte(c)
		} else {
			fmt.Fprintf(&sb, "\\x%02x", c)
		}
	}
	return sb.String()
}

// encodeReply returns the wire bytes for one reply, honoring
// cmd.ReplyFormat. When ReplyFormat is empty, the legacy behavior is
// preserved: `value + "\r\n"`. Unrecognized values log a warning and
// fall back to plaintext so a YAML typo does not silently drop a
// reply.
//
// RESP2 encodings (per redis.io protocol spec):
//
//	redis-simple   -> "+<value>\r\n"
//	redis-integer  -> ":<value>\r\n"
//	redis-error    -> "-<value>\r\n"
//	redis-bulk     -> "$<len>\r\n<value>\r\n"
//	redis-nil-bulk -> "$-1\r\n"
//	redis-raw      -> value written verbatim
//	redis-array    -> "*<n>\r\n" followed by per-entry bulk encoding of cmd.ReplyBulks
func encodeReply(cmd parser.Command, value string) []byte {
	switch cmd.ReplyFormat {
	case "":
		if value == "" {
			return nil
		}
		return []byte(value + "\r\n")
	case "redis-simple":
		return []byte("+" + value + "\r\n")
	case "redis-integer":
		return []byte(":" + value + "\r\n")
	case "redis-error":
		return []byte("-" + value + "\r\n")
	case "redis-bulk":
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
	case "redis-nil-bulk":
		return []byte("$-1\r\n")
	case "redis-raw":
		return []byte(value)
	case "redis-array":
		var buf strings.Builder
		fmt.Fprintf(&buf, "*%d\r\n", len(cmd.ReplyBulks))
		for _, entry := range cmd.ReplyBulks {
			fmt.Fprintf(&buf, "$%d\r\n%s\r\n", len(entry), entry)
		}
		return []byte(buf.String())
	default:
		log.Warnf("tcp.encodeReply: unknown replyFormat %q, falling back to plaintext", cmd.ReplyFormat)
		if value == "" {
			return nil
		}
		return []byte(value + "\r\n")
	}
}
