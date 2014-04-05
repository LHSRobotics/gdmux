package main

import (
	"bufio"
	"io"
	"log"
	"unicode"
)

type Word struct {
	Type    rune
	Content string
}

type Line struct {
	Words   []Word
	Comment string
}

type Parser struct {
	scan *bufio.Scanner
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		scan: bufio.NewScanner(r),
	}
}

func (p *Parser) Next() (*Line, error) {
	if !p.scan.Scan() {
		if p.scan.Err() != nil {
			return nil, p.scan.Err()
		}
		return nil, io.EOF
	}

	return line(p.scan.Text()), nil
}

func line(t string) *Line {
	l := Line{}
	pos := 0

	for pos < len(t) {
		switch b := t[pos]; {
		case unicode.IsSpace(rune(b)):
			pos++
		case b == ';': // ;-style comment
			return &l
		case b == '(': // ()-style comment
			end := pos + 1
			for end < len(t) {
				if t[end] == ')' {
					l.Comment = t[pos:end]
					break
				}
				end++
			}
			pos = end
		case b == 'n' || b == 'N': // Line number, probably not worth parsing here...
			end := pos + 1
			for end < len(t) {
				if unicode.IsSpace(rune(t[end])) || end == len(t)-1 {
					// parse number
					break
				}
				end++
			}
			pos = end
		case b >= 'A' && b <= 'z': // Regular code
			end := pos + 1
			for end < len(t) {
				if unicode.IsSpace(rune(t[end])) {
					break
				}
				end++
			}
			w := Word{
				Type: unicode.ToUpper(rune(b)),
				Content: t[pos:end],
			}
			l.Words = append(l.Words, w)
			pos = end
		default:
			log.Printf("couldn't parse line: %v %v", b, t)
			return nil
		}
	}
	return &l
}
