package main

import (
	"encoding/base64"
	"io"
	"os"
	"strings"
)

// MimeWrapWriter adds a newline pair (CRLF) every 76 bytes
type MimeWrapWriter struct {
	Out io.Writer
}

func (w MimeWrapWriter) Write(b []byte) (written int, err error) {
	// https://tools.ietf.org/html/rfc2822#section-2.1.1
	stride := 76
	var n int

	for left := 0; left < len(b); left += stride {
		right := left + stride
		if right > len(b) {
			right = len(b)
		}

		n, err = w.Out.Write(b[left:right])
		if err != nil {
			return
		}
		written += n

		// The newlines are not a part of the provide slice. Do not count.
		_, err = w.Out.Write([]byte("\r\n"))
		if err != nil {
			return
		}
	}

	return
}

func main() {
	input := strings.NewReader(strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100))

	writer := MimeWrapWriter{os.Stdout}

	encoder := base64.NewEncoder(base64.StdEncoding, writer)

	io.Copy(encoder, input)

	encoder.Close()
}
