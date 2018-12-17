package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"

	"github.com/pkg/errors"
)

func main() {

	parts := Parts{
		Part{
			Name: "textpart1",
			Source: TextPart{
				ContentType: "text/plain/part",
				Text:        "This is the text that goes in the plain part. It will need to be wrapped to 76 characters and quoted.",
			},
		},
		Part{
			Name: "filepart1",
			Source: File{
				Name:   "filename.txt",
				Reader: strings.NewReader("Filename text content"),
			},
		},
		// Part{
		// 	Name: "jsonpart1",
		// 	Source: JSON{
		// 		Value: map[string]int{"one": 1, "two": 2},
		// 	},
		// },
	}

	buf := &bytes.Buffer{}

	header, _ := parts.Into(buf)

	fmt.Println(header)
	fmt.Println(buf)
}

// Based on https://github.com/skillian/mparthelp/
// with help from https://github.com/philippfranke/multipart-related/

// Parts is a collection of parts of a multipart message.
type Parts []Part

// Into creates a multipart message into the given target from the provided
// parts.
func (p Parts) Into(target io.Writer) (formDataContentType string, err error) {
	w := multipart.NewWriter(target)
	for _, part := range p {
		err = part.Source.Add(part.Name, w)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to add %T part %v", part, part))
			return
		}
	}
	formDataContentType = w.FormDataContentType()
	return formDataContentType, w.Close()
}

// Part defines a named part inside of a multipart message.
type Part struct {
	Name string
	Source
}

// Source is a data source that can add itself to a mime/multipart.Writer.
type Source interface {
	Add(name string, w *multipart.Writer) error
}

// JSON is a Source implementation that handles marshaling a value to JSON
type JSON struct {
	Value interface{}
}

// Add implements the Source interface.
func (j JSON) Add(name string, w *multipart.Writer) error {
	jsonBytes, err := json.Marshal(j.Value)
	if err != nil {
		return err
	}
	part, err := w.CreateFormField(name)
	if err != nil {
		return err
	}
	jsonBuffer := bytes.NewBuffer(jsonBytes)
	_, err = io.Copy(part, jsonBuffer)
	return err
}

// File is a Source implementation for files read from an io.Reader.
type File struct {
	// Name is the name of the file, not to be confused with the name of the
	// Part.
	Name string

	// Reader is the data source that the part is populated from.
	io.Reader

	// Closer is an optional io.Closer that is called after reading the Reader
	io.Closer
}

// Add implements the Source interface.
func (f File) Add(name string, w *multipart.Writer) error {
	part, err := w.CreateFormFile(name, f.Name)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f.Reader)
	if err != nil {
		return err
	}
	if f.Closer != nil {
		return f.Closer.Close()
	}
	return nil
}

// https://github.com/domodwyer/mailyak/blob/master/attachments.go#L142
func CreateQuoteTypePart(writer *multipart.Writer, contentType string) (w *quotedprintable.Writer, err error) {
	header := textproto.MIMEHeader{
		"Content-Type":              []string{contentType},
		"Content-Transfer-Encoding": []string{"quoted-printable"},
	}

	var part io.Writer
	part, err = writer.CreatePart(header)
	if err != nil {
		return
	}

	w = quotedprintable.NewWriter(part)
	return
}

type TextPart struct {
	Text        string
	ContentType string
}

// Add implements the Source interface.
func (p TextPart) Add(name string, w *multipart.Writer) error {
	quotedPart, err := CreateQuoteTypePart(w, name)
	if err != nil {
		return err
	}

	var n int
	n, err = quotedPart.Write([]byte(p.Text))
	if err != nil {
		return err
	}

	if n != len(p.Text) {
		fmt.Println("Didn't write enough!")
	}

	// Need to close the printer after writing
	// https://golang.org/pkg/mime/quotedprintable/#Writer.Close
	quotedPart.Close()

	return err
}
