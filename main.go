package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/term/termios"
	"io"
	"os"
	"syscall"
	"unicode"
)

func die(message string) {
	fmt.Fprintf(os.Stderr, message)
	os.Exit(1)
}

func disableRawMode(term *syscall.Termios) {
	err := termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, term)
	if err != nil {
		die("tcsetattr")
	}
}

func enableRawMode(term *syscall.Termios) {
	newTerm := *term
	err := termios.Tcgetattr(uintptr(syscall.Stdin), &newTerm)
	if err != nil {
		die("tcgetattr")
	}
	newTerm.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP |  syscall.IXON
	newTerm.Oflag &^= syscall.OPOST
	newTerm.Cflag |= syscall.CS8
	newTerm.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	newTerm.Cc[syscall.VMIN] = 0
	newTerm.Cc[syscall.VTIME] = 1
	if err = termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, &newTerm); err != nil {
		die("tcsetattr")
	}
}

func getTerm() *syscall.Termios {
	var term syscall.Termios
	err := termios.Tcgetattr(uintptr(syscall.Stdin), &term)
	if err != nil {
		die("tcgetattr")
	}
	return &term
}

func main() {
	var term = getTerm()
	enableRawMode(term)
	defer disableRawMode(term)
	stdin := bufio.NewReader(os.Stdin)
	for {
		ch, err := stdin.ReadByte()
		if err != nil && err != io.EOF {
			die("read")
		}
		r := rune(ch)
		if r == 'q' {
			break
		}
		if unicode.IsControl(r) {
			fmt.Printf("%d\r\n", r)
		} else {
			fmt.Printf("%d ('%c')\r\n", r, r)
		}
	}
}
