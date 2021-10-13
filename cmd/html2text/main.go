package main

import (
	"bufio"
	"fmt"
	"os"

	"jaytaylor.com/html2text"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	opts := html2text.Options{}
	out, err := html2text.FromReader(reader, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(out)
}
