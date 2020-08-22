package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"

	"github.com/skillian/mparthelp"
)

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

type TextPart struct {
	Text string
}

// Add implements the Source interface.
func (p TextPart) Add(name string, w *multipart.Writer) error {
	part, err := CreateQuoteTypePart(w, name)
	if err != nil {
		return err
	}
	_, err = part.Write([]byte(p.Text))
	return err
}

func main() {

	parts := mparthelp.Parts{
		mparthelp.Part{
			Name: "textpart",
			Source: TextPart{
				Text: "This is the text that goes in the plain part. It will need to be wrapped to 76 characters and quoted.",
			},
		},

		mparthelp.Part{
			Name: "filepart",
			Source: mparthelp.File{
				Name:   "filename.txt",
				Reader: strings.NewReader("Filename text content"),
			},
		},
		mparthelp.Part{
			Name: "jsonpart",
			Source: mparthelp.JSON{
				Value: map[string]int{"one": 1, "two": 2},
			},
		},
	}

	buf := &bytes.Buffer{}

	header, _ := parts.Into(buf)

	fmt.Println(header)
	fmt.Println(buf)
}
