package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/term/termios"
	"io"
	"os"
	"syscall"
)
const (
	CONTINUE = -999
)

func controlKey(r rune) rune {
	return r & 0x1f
}

func die(message string) {
        fmt.Printf("\x1b[2J")
        fmt.Printf("\x1b[H")

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

func editorDrawRows() {
	for y := 0; y < 24; y++ {
		fmt.Printf("~\r\n")
	}
}

func editorRefreshScreen() {
	fmt.Printf("\x1b[2J")
	fmt.Printf("\x1b[H")

	editorDrawRows()

	fmt.Printf("\x1b[H")
}

func editorReadKey(stdin *bufio.Reader) rune {
	for {
		ch, err := stdin.ReadByte()
		if err != nil && err != io.EOF {
			die("read")
		}
		return rune(ch)
	}
}

func editorProcessKeypress(stdin *bufio.Reader) int {
	r := editorReadKey(stdin)

	switch r {
		case controlKey('q'):
		        fmt.Printf("\x1b[2J")
		        fmt.Printf("\x1b[H")
			return 0
		default:
			return CONTINUE
	}
}

func main() {
	os.Exit(_main())
}

func _main() int {
        var term = getTerm()
        enableRawMode(term)
        defer disableRawMode(term)
        stdin := bufio.NewReader(os.Stdin)
        for {
		editorRefreshScreen()
		ret := editorProcessKeypress(stdin)
		if ret != CONTINUE {
			return ret
		}
        }
	return 0

}
