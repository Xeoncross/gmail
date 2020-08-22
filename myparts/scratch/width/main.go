package main

import (
	"fmt"

	"golang.org/x/text/width"
)

func main() {
	s := "１２３ ∑ - ði ıntə"
	n := width.Narrow.String(s)
	fmt.Printf("%U: %s\n", []rune(s), s)
	fmt.Printf("%U: %s\n", []rune(n), n)
}
