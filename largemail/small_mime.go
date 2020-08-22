package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
)

// https://stackoverflow.com/questions/3902455/mail-multipart-alternative-vs-multipart-mixed
// https://github.com/jhillyerd/enmime/blob/master/builder.go#L225
// Fully loaded structure; the presence of text, html, inlines, and attachments will determine
// how much is necessary:
//
//  multipart/mixed
//  |- multipart/related
//  |  |- multipart/alternative
//  |  |  |- text/plain
//  |  |  `- text/html
//  |  `- inlines..
//  `- attachments..

/*
// Start our multipart/mixed part
	mixed := multipart.NewWriter(&buf)
	if err := mixed.SetBoundary(mb); err != nil {
		return nil, err
	}
	defer mixed.Close()

	fmt.Fprintf(&buf, "Content-Type: multipart/mixed;\r\n\tboundary=\"%s\"; charset=UTF-8\r\n\r\n", mixed.Boundary())

	ctype := fmt.Sprintf("multipart/alternative;\r\n\tboundary=\"%s\"", ab)

	altPart, err := mixed.CreatePart(textproto.MIMEHeader{"Content-Type": {ctype}})
	if err != nil {
		return nil, err
}
*/

func main() {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	var part io.Writer
	var err error

	// Text Content
	part, err = writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"multipart/alternative"}})
	if err != nil {
		log.Fatal(err)
	}

	childWriter := multipart.NewWriter(part)

	var subpart io.Writer
	for _, contentType := range []string{"text/plain", "text/html"} {
		subpart, err = CreateQuoteTypePart(childWriter, contentType)
		if err != nil {
			log.Fatal(err)
		}
		_, err := subpart.Write([]byte("This is a line of text that needs to be wrapped by quoted-printable before it goes to far.\r\n\r\n"))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Attachments
	filename := fmt.Sprintf("File_%d.jpg", rand.Int31())
	part, err = writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"attachment; filename=" + filename}})
	if err != nil {
		log.Fatal(err)
	}
	part.Write([]byte("AABBCCDDEEFF"))

	writer.Close()

	fmt.Print(`From: Bob <bob@example.com>
To: Alice <alias@example.com>
Subject: Formatted text mail
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=`)
	fmt.Println(writer.Boundary())
	fmt.Println(body.String())

}

// https://github.com/domodwyer/mailyak/blob/master/attachments.go#L142
func CreateQuoteTypePart(writer *multipart.Writer, contentType string) (part io.Writer, err error) {
	header := textproto.MIMEHeader{
		"Content-Type":              []string{contentType},
		"Content-Transfer-Encoding": []string{"quoted-printable"},
	}

	part, err = writer.CreatePart(header)
	if err != nil {
		return
	}
	part = quotedprintable.NewWriter(part)
	return
}
