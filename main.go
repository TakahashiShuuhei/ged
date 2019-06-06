package main

import (
	"bufio"
	"io"
	"os"
	"fmt"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)
	for {
		ch, err := stdin.ReadByte()
		if err == io.EOF {
			break
		}
		r := rune(ch)
		fmt.Printf("AAA %c\n", r)
	}
}
