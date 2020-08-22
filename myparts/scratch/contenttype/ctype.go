package main

import (
	"fmt"
	"mime"
)

func main() {
	param := map[string]string{
		"charset": "utf-8",
		"name":    "foobar.jpg",
	}
	mt := mime.FormatMediaType("attachment", param)

	fmt.Println(mt)
}
