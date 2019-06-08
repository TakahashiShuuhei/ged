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
	GED_VERSION = "0.0.1"
	ARROW_LEFT = 1000
	ARROW_RIGHT = 1001
	ARROW_UP = 1002
	ARROW_DOWN = 1003
)

type EditorConfig struct {
	cx int
	cy int
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
		if y == E.screenRows / 3 {
			welcome := fmt.Sprintf("GED editor -- version %s", GED_VERSION)
			if len(welcome) > E.screenCols {
				welcome = welcome[:E.screenCols]
			}
			padding := (E.screenCols - len(welcome)) / 2
			if padding > 0 {
				abAppend(ab, "~")
				padding--
			}
			for i := 0; i < padding; i++ {
				abAppend(ab, " ")
			}
			abAppend(ab, welcome)
		} else {
			abAppend(ab, "~")
		}

		abAppend(ab, "\x1b[K")
		if y < E.screenRows - 1 {
			abAppend(ab, "\r\n")
		}
	}
}

func editorRefreshScreen() {
	ab := abuf{}

	abAppend(&ab, "\x1b[?25l")
	abAppend(&ab, "\x1b[H")

	editorDrawRows(&ab)

	buf := fmt.Sprintf("\x1b[%d;%dH", E.cy + 1, E.cx + 1)
	abAppend(&ab, buf)

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

		r := rune(ch)

		if r == '\x1b' {
			ch, err = stdin.ReadByte()
			if err != nil && err != io.EOF {
				return r
			}
			r1 := rune(ch)
			ch, err = stdin.ReadByte()
			if err != nil && err != io.EOF {
				return r
			}
			r2 := rune(ch)

			if r1 == '[' {
				switch r2 {
					case 'A': return ARROW_UP
					case 'B': return ARROW_DOWN
					case 'C': return ARROW_RIGHT
					case 'D': return ARROW_LEFT
				}
			}
			return r
		}
		return r
	}
}

func getWindowSize() (rows int, cols int, err error) {
	s, err := ts.GetSize()
	if err != nil {
		return -1, -1, err
	}
	return int(s.Row()), int(s.Col()), nil
}

func editorMoveCursor(key rune) {
	switch key {
		case ARROW_LEFT:
			E.cx--
		case ARROW_RIGHT:
			E.cx++
		case ARROW_UP:
			E.cy--
		case ARROW_DOWN:
			E.cy++
	}
}

func editorProcessKeypress(stdin *bufio.Reader) int {
	r := editorReadKey(stdin)
	switch r {
		case controlKey('q'):
		        fmt.Printf("\x1b[2J")
		        fmt.Printf("\x1b[H")
			return 0
		case ARROW_UP:
			fallthrough
		case ARROW_DOWN:
			fallthrough
		case ARROW_LEFT:
			fallthrough
		case ARROW_RIGHT:
			editorMoveCursor(r)
			return CONTINUE
		default:
			return CONTINUE
	}
}

func initEditor() {
	E.cx = 0
	E.cy = 0
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
