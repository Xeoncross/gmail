package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {

	parts := Parts{
		Part{
			Name: TextPlain,
			Source: TextPart{
				Text: "This is the text that goes in the plain part. It will need to be wrapped to 76 characters and quoted.",
			},
		},
		Part{
			Name: "filepart1",
			Source: File{
				Name:   "filename.jpg",
				Reader: strings.NewReader("Filename text content"),
			},
		},
		Part{
			Name: "filepart1",
			Source: File{
				Name:   "filename-2 שלום.txt",
				Inline: true,
				Reader: strings.NewReader("Filename text content"),
			},
		},
		Part{
			Name: "jsonpart1",
			Source: JSON{
				Value: map[string]int{"one": 1, "two": 2},
			},
		},
	}

	buf := &bytes.Buffer{}

	header, _ := parts.Into(buf)

	fmt.Println(header)
	fmt.Println(buf)
}
