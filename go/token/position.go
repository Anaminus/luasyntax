package token

import (
	"strconv"
	"sync"
)

type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

func (pos *Position) IsValid() bool {
	return pos.Line > 0
}

func (pos *Position) String() string {
	s := pos.Filename
	if pos.IsValid() {
		s += ":" + strconv.Itoa(pos.Line) + ":" + strconv.Itoa(pos.Column)
	}
	if s == "" {
		s = "-"
	}
	return s
}

type File struct {
	name  string
	mutex sync.Mutex
	lines []int
}

func NewFile(filename string) *File {
	return &File{
		name:  filename,
		lines: []int{0},
	}
}

func (f *File) Name() string {
	return f.name
}

func (f *File) LineCount() (c int) {
	f.mutex.Lock()
	c = len(f.lines)
	f.mutex.Unlock()
	return c
}

func (f *File) AddLine(offset int) {
	f.mutex.Lock()
	if i := len(f.lines); i == 0 || f.lines[i-1] < offset {
		f.lines = append(f.lines, offset)
	}
	f.mutex.Unlock()
}

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
