package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/term/termios"
	"github.com/olekukonko/ts"
	"io"
	"os"
	"syscall"
)
const (
	CONTINUE = -999
)

type EditorConfig struct {
	screenRows int
	screenCols int
	origTermios syscall.Termios
}

type abuf struct {
	b string
}

func abAppend(ab *abuf, s string) {
	ab.b += s
}

func abFree(ab *abuf) {
	ab.b = ""
}

var E EditorConfig = EditorConfig{
	origTermios: syscall.Termios{},
}

func controlKey(r rune) rune {
	return r & 0x1f
}

func die(message string) {
        fmt.Printf("\x1b[2J")
        fmt.Printf("\x1b[H")

	fmt.Fprintf(os.Stderr, message)
	os.Exit(1)
}

func disableRawMode() {
	err := termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, &E.origTermios)
	if err != nil {
		die("tcsetattr")
	}
}

func enableRawMode() {
	newTerm := E.origTermios
	newTerm.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP |  syscall.IXON
	newTerm.Oflag &^= syscall.OPOST
	newTerm.Cflag |= syscall.CS8
	newTerm.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	newTerm.Cc[syscall.VMIN] = 0
	newTerm.Cc[syscall.VTIME] = 1
	if err := termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, &newTerm); err != nil {
		die("tcsetattr")
	}
}

func getTerm() {
	err := termios.Tcgetattr(uintptr(syscall.Stdin), &E.origTermios)
	if err != nil {
		die("tcgetattr")
	}
}

func editorDrawRows(ab *abuf) {
	for y := 0; y < E.screenRows; y++ {
		abAppend(ab, "~")
		if y < E.screenRows - 1 {
			abAppend(ab, "\r\n")
		}
	}
}

func editorRefreshScreen() {
	ab := abuf{}

	abAppend(&ab, "\x1b[?25l")
	abAppend(&ab, "\x1b[2J")
	abAppend(&ab, "\x1b[H")

	editorDrawRows(&ab)

	abAppend(&ab, "\x1b[H")
	abAppend(&ab, "\x1b[?25h")

	fmt.Printf(ab.b)
	abFree(&ab)
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

func getWindowSize() (rows int, cols int, err error) {
	s, err := ts.GetSize()
	if err != nil {
		return -1, -1, err
	}
	return int(s.Row()), int(s.Col()), nil
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

func initEditor() {
	rows, cols, err := getWindowSize()
	if err != nil {
		die("getWindowSize")
	}
	E.screenRows = rows
	E.screenCols = cols
}

func main() {
	os.Exit(_main())
}

func _main() int {
	getTerm()
	enableRawMode()
	defer disableRawMode()
	initEditor()
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
