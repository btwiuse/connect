package connect

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

var Handler = http.HandlerFunc(Connect)

func Connect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	connect(w, r)
}

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func connect(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, "Failed to connect to the destination host", http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	switch r.ProtoMajor {
	case 1:
		serveHijack(w, destConn)
	default:
		// Write 200 OK to the client to establish the connection
		w.WriteHeader(http.StatusOK)
		http.NewResponseController(w).Flush()

		// Start copying data between client and destination host
		dualStream(destConn, r.Body, w)
	}
}

func serveHijack(w http.ResponseWriter, targetConn net.Conn) error {
	clientConn, bufReader, err := http.NewResponseController(w).Hijack()
	if err != nil {
		return fmt.Errorf("hijack failed: %v", err)
	}
	defer clientConn.Close()
	// bufReader may contain unprocessed buffered data from the client.
	if bufReader != nil {
		// snippet borrowed from `proxy` plugin
		if n := bufReader.Reader.Buffered(); n > 0 {
			rbuf, err := bufReader.Reader.Peek(n)
			if err != nil {
				return err
			}
			_, _ = targetConn.Write(rbuf)

		}
	}
	// Since we hijacked the connection, we lost the ability to write and flush headers via w.
	// Let's handcraft the response and send it manually.
	res := &http.Response{
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	res.Header.Set("Server", "Caddy")

	buf := bufio.NewWriter(clientConn)
	err = res.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to write response: %v", err)
	}
	err = buf.Flush()
	if err != nil {
		return fmt.Errorf("failed to send response to client: %v", err)
	}

	return dualStream(targetConn, clientConn, clientConn)
}

// Copies data target->clientReader and clientWriter->target, and flushes as needed
// Returns when clientWriter-> target stream is done.
// Caddy should finish writing target -> clientReader.
func dualStream(target net.Conn, clientReader io.ReadCloser, clientWriter io.Writer) error {
	stream := func(w io.Writer, r io.Reader) error {
		// copy bytes from r to w
		bufPtr := bufferPool.Get().(*[]byte)
		buf := *bufPtr
		buf = buf[0:cap(buf)]
		_, _err := flushingIoCopy(w, r, buf)
		bufferPool.Put(bufPtr)

		if cw, ok := w.(closeWriter); ok {
			_ = cw.CloseWrite()
		}
		return _err
	}
	go stream(target, clientReader) //nolint: errcheck
	return stream(clientWriter, target)
}

type closeWriter interface {
	CloseWrite() error
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, 0, 32*1024)
		return &buffer
	},
}

// flushingIoCopy is analogous to buffering io.Copy(), but also attempts to flush on each iteration.
// If dst does not implement http.Flusher(e.g. net.TCPConn), it will do a simple io.CopyBuffer().
// Reasoning: http2ResponseWriter will not flush on its own, so we have to do it manually.
func flushingIoCopy(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	rw, ok := dst.(http.ResponseWriter)
	if !ok {
		return io.CopyBuffer(dst, src, buf)
	}
	rc := http.NewResponseController(rw)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			ef := rc.Flush()
			if ef != nil {
				err = ef
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}
