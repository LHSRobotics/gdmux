package gcode

import (
	"bufio"
	"io"
	"fmt"
	"unicode"
)

type Code string

type Line struct {
	Codes   []Code
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

	return line(p.scan.Text())
}

func line(t string) (*Line, error) {
	l := Line{}
	pos := 0

	for pos < len(t) {
		switch b := t[pos]; {
		case unicode.IsSpace(rune(b)):
			pos++
		case b == ';': // ;-style comment
			return &l, nil
		case b == '(': // ()-style comment
			end := pos + 1
			for end < len(t) {
				if t[end] == ')' {
					l.Comment = t[pos:end]
					end++
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
			l.Codes = append(l.Codes, Code(t[pos:end]))
			pos = end
		default:
			return nil, fmt.Errorf("couldn't parse line: %c %v", b, t)
		}
	}
	return &l, nil
}
