package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/philippfranke/multipart-related/related"
)

func main() {
	fileContents1 := []byte(`Life? Don't talk to me about life!`)
	fileContents2 := []byte(`Marvin`)

	var b bytes.Buffer
	w := related.NewWriter(&b)
	w.SetType("multipart/mixed")
	{
		part, err := w.CreateRoot("", "text/plain", nil)
		if err != nil {
			log.Fatalf("CreateRoot: %v", err)
		}
		part.Write(fileContents1)

		nextPart, err := w.CreatePart("", nil)
		if err != nil {
			log.Fatalf("CreatePart 2: %v", err)
		}

		nextPart.Write(fileContents2)

		if err := w.Close(); err != nil {
			log.Fatalf("Close: %v", err)
		}

		s := b.String()
		if len(s) == 0 {
			log.Fatal("String: unexpected empty result")
		}

		fmt.Println("Boundary", w.Boundary())

		fmt.Printf("The compound Object Content-Type:\n %s \n", w.FormDataContentType())
		fmt.Println(s)
	}

}
