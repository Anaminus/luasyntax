package token

import (
	"strconv"
	"sync"
)

// A Position describes a position with a file, including the name of the file,
// the line, and the column.
type Position struct {
	Filename string // The name of the file, if specified.
	Offset   int    // The offset within the file, starting at 0.
	Line     int    // The line number, starting at 1.
	Column   int    // The column number, starting at 1.
}

// IsValid returns whether the position is valid.
func (pos *Position) IsValid() bool {
	return pos.Line > 0
}

// String returns a formatted representation of the position, which may be one
// of several forms:
//
//     file:line:column    Filename with valid position.
//     line:column         No filename with valid position.
//     file                Filename with invalid position.
//     -                   No filename with invalid position.
func (pos Position) String() string {
	s := pos.Filename
	if pos.IsValid() {
		s += ":" + strconv.Itoa(pos.Line) + ":" + strconv.Itoa(pos.Column)
	}
	if s == "" {
		s = "-"
	}
	return s
}

// A File represents a Lua source file. Methods are safe to use concurrently.
type File struct {
	name  string
	mutex sync.Mutex
	lines []int
}

// NewFile creates a new file with the given name.
func NewFile(filename string) *File {
	return &File{
		name:  filename,
		lines: []int{0},
	}
}

// Name returns the name of the file.
func (f *File) Name() string {
	return f.name
}

// Returns the number of lines in the file.
func (f *File) LineCount() (c int) {
	f.mutex.Lock()
	c = len(f.lines)
	f.mutex.Unlock()
	return c
}

// AddLine specifies the location of the start of a line in the file.
func (f *File) AddLine(offset int) {
	f.mutex.Lock()
	if i := len(f.lines); i == 0 || f.lines[i-1] < offset {
		f.lines = append(f.lines, offset)
	}
	f.mutex.Unlock()
}

// SetLinesForContent sets the line offset for a given file content.
func (f *File) SetLinesForContent(content []byte) {
	lines := []int{}
	line := 0
	for offset, b := range content {
		if line >= 0 {
			lines = append(lines, line)
		}
		line = -1
		if b == '\n' {
			line = offset + 1
		}
	}

	f.mutex.Lock()
	f.lines = lines
	f.mutex.Unlock()
}

// ClearLines resets the line offsets for a file.
func (f *File) ClearLines() {
	f.mutex.Lock()
	f.lines = f.lines[:0]
	f.mutex.Unlock()
}

// searchInts performs a binary search for the nearest position of an int within
// a slice of ints.
func searchInts(a []int, x int) int {
	// return sort.Search(len(a), func(i int) bool { return a[i] > x }) - 1
	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)/2
		if a[h] <= x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i - 1
}

// Position returns the Position value for a given offset within the file.
func (f *File) Position(offset int) (pos Position) {
	pos.Offset = offset
	pos.Filename = f.name
	f.mutex.Lock()
	if i := searchInts(f.lines, offset); i >= 0 {
		pos.Line, pos.Column = i+1, offset-f.lines[i]+1
	}
	f.mutex.Unlock()
	return pos
}
