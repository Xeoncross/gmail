package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// func (wc *WriteCounter) Write(p []byte) (int, error) {

type Wrapper struct {
	w io.Writer
}

// func Write(p []byte) (int, error) {
//
// 	pipeReader, pipeWriter := io.Pipe()
//
// 	buf := make([]byte, 76)
// 	for {
// 		n, err := in.Read(buf)
//
// 		if err != nil {
// 			if err != io.EOF {
// 				log.Fatal(err)
// 				break
// 			}
// 		}
//
// 		// Process
// 		fmt.Println(n, buf[:n])
//
// 		// handle any remainding bytes before exit
// 		if err == io.EOF {
// 			break
// 		}
// 	}
// }

func main() {
	// 1: Source stream
	in := strings.NewReader(strings.Repeat("\x00Hello World! 012345.", 30))

	// 5: Destination
	// out := os.Stdout

	buf := make([]byte, 76)
	for {
		n, err := in.Read(buf)

		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
				break
			}
		}

		// Process
		fmt.Fprintf(os.Stdout, "%s\n", buf[:n])

		// handle any remainding bytes before exit
		if err == io.EOF {
			break
		}
	}

	// 3: transform
	// chunked := httputil.NewChunkedWriter(out)
	//
	// // 2: Encode
	// encoder := base64.NewEncoder(base64.StdEncoding, chunked)
	//
	// // 4: Run
	// io.Copy(encoder, in)

	// Must close the encoder when finished to flush any partial blocks.
	// If you comment out the following line, the last partial block "r"
	// won't be encoded.
	// encoder.Close()
	// chunked.Close()
}

// func main() {
// 	// 1: Source stream
// 	in := strings.NewReader(strings.Repeat("\x00Hello World! 012345.", 30))
//
// 	// 5: Destination
// 	out := os.Stdout
//
// 	// 3: transform
// 	chunked := httputil.NewChunkedWriter(out)
//
// 	// 2: Encode
// 	encoder := base64.NewEncoder(base64.StdEncoding, chunked)
//
// 	// 4: Run
// 	io.Copy(encoder, in)
//
// 	// Must close the encoder when finished to flush any partial blocks.
// 	// If you comment out the following line, the last partial block "r"
// 	// won't be encoded.
// 	encoder.Close()
// 	chunked.Close()
// }

// func Wrap(w io.Writer) (out io.Writer) {
// 	pipeReader, pipeWriter := io.Pipe()
//
// 	for {
//
// 	}
//
// 	scanner := bufio.NewScanner(pipeReader)
// 	scanner.SplitFunc(bufio.ScanRunes(data, atEOF))
//
// }
