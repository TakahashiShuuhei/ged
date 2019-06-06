package main

import (
	"bufio"
	"io"
	"github.com/pkg/term/termios"
	"os"
	"syscall"
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
	newTerm.Lflag &^= syscall.ECHO
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
		ch, err := stdin.ReadByte()
		if err == io.EOF {
			break
		}
		r := rune(ch)
		if r == 'q' {
			break
		}
	}
}
