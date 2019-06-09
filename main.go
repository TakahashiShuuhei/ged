package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/term/termios"
	"github.com/olekukonko/ts"
	"io"
	"os"
	"strings"
	"syscall"
)
const (
	CONTINUE = -999
	GED_VERSION = "0.0.1"
	ARROW_LEFT = 1000
	ARROW_RIGHT = 1001
	ARROW_UP = 1002
	ARROW_DOWN = 1003
	DEL_KEY = 1004
	HOME_KEY = 1005
	END_KEY = 1006
	PAGE_UP = 1007
	PAGE_DOWN = 1008
)

type EditorConfig struct {
	cx int
	cy int
	rowoff int
	coloff int
	screenRows int
	screenCols int
	row []erow
	origTermios syscall.Termios
}

type abuf struct {
	b string
}

type erow struct {
	chars string
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

func editorScroll() {
	if E.cy < E.rowoff {
		E.rowoff = E.cy
	}

	if E.cy >= E.rowoff + E.screenRows {
		E.rowoff = E.cy - E.screenRows + 1
	}

	if E.cx < E.coloff {
		E.coloff = E.cx
	}

	if E.cx >= E.coloff + E.screenCols {
		E.coloff = E.cx - E.screenCols + 1
	}
}

func editorDrawRows(ab *abuf) {
	for y := 0; y < E.screenRows; y++ {
		filerow := y + E.rowoff
		if filerow >= len(E.row) {
			if len(E.row) == 0 && y == E.screenRows / 3 {
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
		} else {
			line := ""
			if len(E.row[filerow].chars) > E.coloff {
				line = E.row[filerow].chars[E.coloff:]
			}
			if len(line) > E.screenCols {
				line = line[:E.screenCols]
			}
			abAppend(ab, line)
		}

		abAppend(ab, "\x1b[K")
		if y < E.screenRows - 1 {
			abAppend(ab, "\r\n")
		}
	}
}

func editorRefreshScreen() {
	editorScroll()

	ab := abuf{}

	abAppend(&ab, "\x1b[?25l")
	abAppend(&ab, "\x1b[H")

	editorDrawRows(&ab)

	buf := fmt.Sprintf("\x1b[%d;%dH", E.cy - E.rowoff + 1, E.cx - E.coloff + 1)
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
				if r2 >= '0' && r2 <= '9' {
					ch, err = stdin.ReadByte()
					if err != nil && err != io.EOF {
						return r
					}
					r3 := rune(ch)
					if r3 == '~' {
						switch r2 {
							case '1': return HOME_KEY
							case '3': return DEL_KEY
							case '4': return END_KEY
							case '5': return PAGE_UP
							case '6': return PAGE_DOWN
							case '7': return HOME_KEY
							case '8': return END_KEY
						}
					}
				} else {
					switch r2 {
						case 'A': return ARROW_UP
						case 'B': return ARROW_DOWN
						case 'C': return ARROW_RIGHT
						case 'D': return ARROW_LEFT
						case 'H': return HOME_KEY
						case 'F': return END_KEY
					}
				}
			} else if r1 == 'O' {
				switch r2 {
					case 'H': return HOME_KEY
					case 'F': return END_KEY
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

func editorAppendRow(s string) {
	r := erow{chars: s}
	E.row = append(E.row, r)
}

func editorOpen(filename string) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		die("Open")
	}
	reader := bufio.NewReader(f)
	for {
		bs, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			die("ReadLine")
		}
		line := fmt.Sprintf("%s", bs)
		line = strings.Replace(line, "\r", "", -1)
		line = strings.Replace(line, "\n", "", -1)

		editorAppendRow(line)
	}
}

func editorMoveCursor(key rune) {
	var row *erow = nil
	if E.cy < len(E.row) {
		row = &E.row[E.cy]
	}
	switch key {
		case ARROW_LEFT:
			if E.cx != 0 {
				E.cx--
			} else if E.cy > 0 {
				E.cy--
				E.cx = len(E.row[E.cy].chars)
			}
		case ARROW_RIGHT:
			if row != nil && E.cx < len(row.chars) {
				E.cx++
			} else if row != nil && E.cx == len(row.chars) {
				E.cy++
				E.cx = 0
			}
		case ARROW_UP:
			if E.cy != 0 {
				E.cy--
			}
		case ARROW_DOWN:
			if E.cy < len(E.row) {
				E.cy++
			}
	}

	row = nil
	if E.cy < len(E.row) {
		row = &E.row[E.cy]
	}
	rowlen := 0
	if row != nil {
		rowlen = len(row.chars)
	}
	if E.cx > rowlen {
		E.cx = rowlen
	}
}

func editorProcessKeypress(stdin *bufio.Reader) int {
	r := editorReadKey(stdin)
	switch r {
		case controlKey('q'):
		        fmt.Printf("\x1b[2J")
		        fmt.Printf("\x1b[H")
			return 0
		case HOME_KEY:
			E.cx = 0
			return CONTINUE
		case END_KEY:
			E.cx = E.screenCols - 1
			return CONTINUE
		case PAGE_UP:
			fallthrough
		case PAGE_DOWN:
			times := E.screenRows
			move := rune(ARROW_UP)
			if r == PAGE_DOWN {
				move = rune(ARROW_DOWN)
			}
			for i := 0; i < times; i++ {
				editorMoveCursor(move)
			}
			return CONTINUE
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
	E.rowoff = 0
	E.coloff = 0
	E.row = nil
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
	if len(os.Args) >= 2 {
		editorOpen(os.Args[1])
	}
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
