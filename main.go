package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/term/termios"
	"os"
	"syscall"
	"unicode"
)

func disableRawMode(term *syscall.Termios) error {
	err := termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, term)
	if err != nil {
		return err
	}
	return nil
}

func enableRawMode(term *syscall.Termios) error {
	newTerm := *term
	err := termios.Tcgetattr(uintptr(syscall.Stdin), &newTerm)
	if err != nil {
		return err
	}
	newTerm.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP |  syscall.IXON
	newTerm.Oflag &^= syscall.OPOST
	newTerm.Cflag |= syscall.CS8
	newTerm.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	newTerm.Cc[syscall.VMIN] = 0
	newTerm.Cc[syscall.VTIME] = 1
	if err = termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, &newTerm); err != nil {
		return err
	}
	return nil
}

func getTerm() (*syscall.Termios, error) {
	var term syscall.Termios
	err := termios.Tcgetattr(uintptr(syscall.Stdin), &term)
	if err != nil {
		return nil, err
	}
	return &term, nil
}

func main() {
	var term, err = getTerm()
	if err != nil {
		return
	}
	err = enableRawMode(term)
	defer disableRawMode(term)
	if err != nil {
		return
	}
	stdin := bufio.NewReader(os.Stdin)
	for {
		ch, _ := stdin.ReadByte()
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
