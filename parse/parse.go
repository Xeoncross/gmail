package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
)

// https://github.com/veqryn/go-email/

const myMessage = `Content-Type: multipart/alternative;
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
Content-Type: text/html; charset="utf-8"
MIME-Version: 1.0
Content-Transfer-Encoding: base64

0K3RgtC+INC80L7RkSDRgdC+0L7QsdGJ0LXQvdC40LUu

--===============5769616449556512256==--`

func main() {
	// msg, err := mail.ReadMessage(bytes.NewBufferString(myMessage))
	// if err != nil {
	// 	log.Fatal("Cannot parse myMessage.")
	// }

	msg, err := ParseMessage(bytes.NewBufferString(myMessage))

	if err != nil {
		log.Fatal(err)
	}

	// _ = msg
	fmt.Println(msg)

	// to, _ := (&mail.AddressParser{}).ParseList(msg.Header.Get("To"))
	// fmt.Println("To", to)
	//
	// mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if strings.HasPrefix(mediaType, "multipart/") {
	// 	mr := multipart.NewReader(msg.Body, params["boundary"])
	// 	for {
	// 		p, err := mr.NextPart()
	// 		if err == io.EOF {
	// 			return
	// 		}
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	//
	// 		// decode any Q-encoded values
	// 		for name, values := range p.Header {
	// 			for idx, val := range values {
	// 				fmt.Printf("%d: %s: %s\n", idx, name, decodeRFC2047(val))
	// 			}
	// 		}
	//
	// 		slurp, err := ioutil.ReadAll(p)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Printf("Content-type: %s\n%s\n", p.Header.Get("Content-Type"), slurp)
	// 	}
	// }
}

// ParseMessage parses and returns a Message from an io.Reader
// containing the raw text of an email message.
// (If the raw email is a string or []byte, use strings.NewReader()
// or bytes.NewReader() to create a reader.)
// Any "quoted-printable" or "base64" encoded bodies will be decoded.
func ParseMessage(r io.Reader) (io.Reader, error) {

	p := textproto.NewReader(bufio.NewReader(s))

	msg, err := mail.ReadMessage(bufioReader(r))
	if err != nil {
		return nil, err
	}
	// mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("ParseMessage mediatype:", mediaType, params)

	// decode any Q-encoded values
	for _, values := range msg.Header {
		for idx, val := range values {
			values[idx] = decodeRFC2047(val)
		}
	}
	return parseMessageWithHeader(msg)
}

// parseMessageWithHeader parses and returns a Message from an already filled
// Header, and an io.Reader containing the raw text of the body/payload.
// (If the raw body is a string or []byte, use strings.NewReader()
// or bytes.NewReader() to create a reader.)
// Any "quoted-printable" or "base64" encoded bodies will be decoded.
// func parseMessageWithHeader(msg *mail.Message) (io.Reader, error) {
func parseMessageWithHeader(header textproto.MIMEHeader, msg io.Reader) (io.Reader, error) {

	// fmt.Println("parseMessageWithHeader", msg.Header)

	// bufferedReader := contentReader(msg.Header, msg.Body)

	var err error
	var mediaType string
	var mediaTypeParams map[string]string

	if contentType := msg.Header.Get("Content-Type"); len(contentType) > 0 {
		mediaType, mediaTypeParams, err = mime.ParseMediaType(contentType)
		fmt.Println("parseMessageWithHeader", mediaType)
		if err != nil {
			return nil, err
		}
	} // Lack of contentType is not a problem

	// Can only have one of the following: Parts, SubMessage, or Body
	if strings.HasPrefix(mediaType, "multipart") {
		boundary := mediaTypeParams["boundary"]

		parts, err := readParts(bufferedReader, boundary)
		if err == nil {
			if parts != nil {
				return parts, nil
			}
		}

	} else if strings.HasPrefix(mediaType, "message") {
		return ParseMessage(bufferedReader)
	}

	if mediaType == "text/plain" {
		return bufferedReader, nil
	}

	return nil, nil
}

// readParts parses out the parts of a multipart body, including the preamble and epilogue.
func readParts(bodyReader io.Reader, boundary string) (io.Reader, error) {

	multipartReader := multipart.NewReader(bodyReader, boundary)

	for part, partErr := multipartReader.NextPart(); partErr != io.EOF; part, partErr = multipartReader.NextPart() {
		if partErr != nil && partErr != io.EOF {
			return nil, partErr
		}

		fmt.Println("readParts", part.Header.Get("Content-Type"))

		// m, err := mail.ReadMessage(part) // fails to keep headers
		// if err != nil {
		// 	return nil, err
		// }

		m := &mail.Message{
			Header: part.Header,
			Body:   part,
		}

		newEmailPart, msgErr := parseMessageWithHeader(m)
		// part.Close()

		if msgErr != nil {
			return nil, msgErr
		}

		if newEmailPart != nil {
			return newEmailPart, nil
		}
	}

	return nil, errors.New("not found")
}

// contentReader ...
func contentReader(headers mail.Header, bodyReader io.Reader) *bufio.Reader {
	if headers.Get("Content-Transfer-Encoding") == "quoted-printable" {
		// headers.Del("Content-Transfer-Encoding")
		return bufioReader(quotedprintable.NewReader(bodyReader))
	}
	if headers.Get("Content-Transfer-Encoding") == "base64" {
		// headers.Del("Content-Transfer-Encoding")
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

// https://github.com/golang/go/issues/4687
// decodeRFC2047 ...
func decodeRFC2047(s string) string {
	// GO 1.5 does not decode headers, but this may change in future releases...
	decoded, err := (&mime.WordDecoder{}).DecodeHeader(s)
	if err != nil || len(decoded) == 0 {
		return s
	}
	return decoded
}
