package main

import (
	"bytes"
	"fmt"
	"strings"
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
