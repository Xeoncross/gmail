package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"
	"unicode"
)

const myMessage = `Content-Type: multipart/mixed;
 boundary="===============5769616449556512256=="
MIME-Version: 1.0
To: test@test.com
From: test@gmail.com
Cc:
Subject: =?utf-8?b?0J/RgNC40LLQtdGC?=
Date: Mon, 30 Jun 2014 18:29:38 -0000

--===============5769616449556512256==
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0
Content-Transfer-Encoding: base64
X-Data: =?utf-8?b?AxfhfujropadladnggnfjgwsaiubvnmkadiuhterqHJSFfuAjkfhrqpeorLA?=
 =?utf-8?b?kFnjNfhgt7Fjd9dfkliodQ==?=

0K3RgtC+INC80L7RkSDRgdC+0L7QsdGJ0LXQvdC40LUu

--===============5769616449556512256==
Content-Type: multipart/alternative; boundary="Enmime-Test-100"

--Enmime-Test-100
Content-Transfer-Encoding: 7bit
Content-Type: text/plain; charset=us-ascii

Section one

--Enmime-Test-100
Content-Type: application/pgp-signature; name="signature.asc"
Content-Description: Digital signature

iQIVAwUBUvPn/Tk1h9l9hlALAQisew
--Enmime-Test-100--

--===============5769616449556512256==--`

func main() {

	dir := "./tmp"
	// dir, err := ioutil.TempDir("", "example")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// defer os.RemoveAll(dir) // clean up

	mw, err := NewEmailFromReader(bytes.NewBufferString(myMessage), dir)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(mw.Parts), "parts found")

	fmt.Println(mw.Close())

}

// MailWrapper around headers and parts
type MailWrapper struct {
	// https://golang.org/src/net/http/header.go?s=1542:1582#L48
	Header textproto.MIMEHeader
	Parts  []*Part
}

func (m MailWrapper) Close() (err error) {
	for _, p := range m.Parts {
		err = p.Close()
		if err != nil {
			return
		}
	}
	return nil
}

// Part is a copyable representation of a multipart.Part
type Part struct {
	Header textproto.MIMEHeader
	Body   io.Reader
	Closer io.ReadCloser
}

func (p Part) Close() error {
	return p.Closer.Close()
}

// trimReader is a custom io.Reader that will trim any leading
// whitespace, as this can cause email imports to fail.
type trimReader struct {
	rd io.Reader
}

// Read trims off any unicode whitespace from the originating reader
func (tr trimReader) Read(buf []byte) (int, error) {
	n, err := tr.rd.Read(buf)
	t := bytes.TrimLeftFunc(buf[:n], unicode.IsSpace)
	n = copy(buf, t)
	return n, err
}

// NewEmailFromReader reads a stream of bytes from an io.Reader, r,
// and returns an email struct containing the parsed data.
// This function expects the data in RFC 5322 format.
func NewEmailFromReader(r io.Reader, dir string) (mw MailWrapper, err error) {
	s := trimReader{rd: r}
	tp := textproto.NewReader(bufio.NewReader(s))

	mw.Header, err = tp.ReadMIMEHeader()
	if err != nil {
		return
	}

	// Recursively parse the MIME parts
	mw.Parts, err = parseMIMEParts(mw.Header, tp.R, dir)
	return
}

// func readAll(r io.Reader) []byte {
// 	b, err := ioutil.ReadAll(r)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return b
// }

// parseMIMEParts will recursively walk a MIME entity and return a []mime.Part containing
// each (flattened) mime.Part found.
// It is important to note that there are no limits to the number of recursions, so be
// careful when parsing unknown MIME structures!
func parseMIMEParts(hs textproto.MIMEHeader, b io.Reader, dir string) (parts []*Part, err error) {

	ct, params, err := mime.ParseMediaType(hs.Get("Content-Type"))
	if err != nil {
		return
	}

	// If it's a multipart email, recursively parse the parts
	if strings.HasPrefix(ct, "multipart/") {

		if _, ok := params["boundary"]; !ok {
			return parts, errors.New("Missing boundary")
		}

		// Readers are buffered https://golang.org/src/mime/multipart/multipart.go#L99
		mr := multipart.NewReader(b, params["boundary"])

		var p *multipart.Part
		for {

			// Decodes quotedprintable: https://golang.org/src/mime/multipart/multipart.go#L128
			// Closes last part reader: https://golang.org/src/mime/multipart/multipart.go#L302
			p, err = mr.NextPart()
			if err == io.EOF {
				break
			}

			if err != nil {
				return
			}

			// Correctly decode the body bytes
			body := contentDecoderReader(p.Header, p)

			// https://golang.org/ref/spec#Type_assertions
			// http.Header and textproto.MIMEHeader are both just a map[string][]string
			// httpHeader := http.Header(p.Header)
			// httpHeader := p.Header.(map[string][]string)
			// httpHeader := (*map[string][]string).(p.Header)
			// fmt.Fprintf(tmpFile, "%#v\n\n\n", httpHeader)

			var subct string
			subct, _, err = mime.ParseMediaType(p.Header.Get("Content-Type"))

			if strings.HasPrefix(subct, "multipart/") {
				// fmt.Println("\tparsing multipart?", subct)

				var subparts []*Part
				subparts, err = parseMIMEParts(p.Header, body, dir)
				if err != nil {
					return
				}
				parts = append(parts, subparts...)

			} else {
				// fmt.Println("\tparsing plain?", subct)

				var tmpFile *os.File
				tmpFile, err = ioutil.TempFile(dir, "mime")
				if err != nil {
					return
				}
				tmpFile.Close()

				_, err = io.Copy(tmpFile, body) // Save body disk
				if err != nil {
					return
				}

				// Rewind for reading
				tmpFile.Seek(0, 0)

				parts = append(parts, &Part{Body: tmpFile, Closer: tmpFile, Header: p.Header})
			}
		}
	} else {
		// If it is not a multipart email, parse the body content as a single "part"
		parts = append(parts, &Part{Body: contentDecoderReader(hs, b), Header: hs})
	}

	return parts, nil
}

// func newTempFile() (os.File, err error) {
//   tmpfile, err = ioutil.TempFile("", "example")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// defer os.Remove(tmpfile.Name()) // clean up
//
// }

// func headersToReader(headers map[string][]string) io.Reader {
// 	var buffer bytes.Buffer // Go 1.10+ can use strings.Builder
// 	for name, values := range headers {
// 		for _, value := range values {
// 			buffer.WriteString(fmt.Sprintf("%s: %s\n", name, value))
// 		}
// 	}
// 	return &buffer
// }

// func headerToReader(header http.Header) io.Reader {
// 	header.Write
// }

// contentDecoderReader
func contentDecoderReader(headers textproto.MIMEHeader, bodyReader io.Reader) *bufio.Reader {
	// Already handled by textproto
	// if headers.Get("Content-Transfer-Encoding") == "quoted-printable" {
	// 	return bufioReader(quotedprintable.NewReader(bodyReader))
	// }
	if headers.Get("Content-Transfer-Encoding") == "base64" {
		return bufioReader(base64.NewDecoder(base64.StdEncoding, bodyReader))
	}
	return bufioReader(bodyReader)
}

// bufioReader ...
func bufioReader(r io.Reader) *bufio.Reader {
	if bufferedReader, ok := r.(*bufio.Reader); ok {
		return bufferedReader
	}
	return bufio.NewReader(r)
}
