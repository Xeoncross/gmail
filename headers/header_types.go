package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
)

const email = `Date: Fri, 19 Oct 2012 12:22:49 -0700
MIME-Version: 1.0
To: user@example.com
Subject: Hello

Hi John! How are you?
`

// Interfaces can have arbitrary underlying types. In this case,
// both the textproto.MIMEHeader and http.Header interfaces are a
// map[string][]string type. This means we can treat them as
// interchangeable.
//
// https://golang.org/src/net/http/header.go#L18
// https://golang.org/src/net/textproto/header.go#L9
func main() {

	// Read email stream in
	tp := textproto.NewReader(bufio.NewReader(strings.NewReader(email)))

	// Parse the main headers
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}

	// Save headers to stream
	var buf bytes.Buffer

	// Straight dump shows "Headers: map[..."
	fmt.Fprintf(&buf, "Headers: %v\n\n", headers)

	// Convert to http.Header (same type underneath)
	httpHeader := http.Header(headers)
	httpHeader.Write(&buf)
	buf.Write([]byte("\n"))

	// JSON sees the real type (map[string][]string)
	e := json.NewEncoder(&buf)
	e.Encode(headers)

	// Show the result
	fmt.Println(buf.String())

	// Read rest of body
	// io.Copy(os.Stdout, tp.R)
}
