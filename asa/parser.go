package asa

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
)

type TokenType int32

const (
	TokenAtom = TokenType(iota)
	TokenLineStart
	TokenError
)

func (t TokenType) String() string {
	switch t {
	case TokenAtom:
		return "TOKEN-ATOM"
	case TokenError:
		return "TOKEN-ERROR"
	default:
		return "*UNRECOGNIZED-TOKEN*"
	}
}

type Lexer struct {
	r *bufio.Reader
}

type Token struct {
	Value string
	Type  TokenType
}

func eatWhiteSpace(r *bufio.Reader) {
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			return
		}
		if unicode.IsSpace(c) == false || c == '\n' {
			r.UnreadRune()
			return
		}
	}
}

func (l Lexer) EatComment() {
	for {
		if c, _, err := l.r.ReadRune(); err == nil || c == '\n' {
			return
		}
	}
}

func NewErrorf(s string, args ...interface{}) *Token {
	return &Token{
		Type:  TokenError,
		Value: fmt.Sprintf(s, args...),
	}
}

func (l Lexer) ReadAtom() *Token {
	b := strings.Builder{}

	for {
		c, _, err := l.r.ReadRune()

		if err != nil {
			return NewErrorf("error while reading: %v", err)
		}

		if (unicode.IsLetter(c) || unicode.IsDigit(c) || unicode.IsPunct(c)) == false {
			l.r.UnreadRune()
			break
		}

		b.WriteRune(c)
	}

	return &Token{
		Value: b.String(),
		Type:  TokenAtom,
	}
}

func (l Lexer) LineStart() *Token {
	c, _, err := l.r.ReadRune()

	if err != nil {
		return NewErrorf("error while scanning for newline: %v", err)
	}

	if c != '\n' {
		return NewErrorf("unexpected rune while scanning for newline: %c", c)
	}

	// we have a new line.

	// peek ahead -- if space, return nil or coninuation token.
	if unicode.IsSpace(l.peek()) {
		return nil
	}

	return &Token{
		Type:  TokenLineStart,
		Value: "NEWLINE TOKEN",
	}
}

func (l Lexer) peek() rune {
	c, _, _ := l.r.ReadRune()
	l.r.UnreadRune()
	return c
}

func (l Lexer) lex(tokChan chan<- *Token) {
	defer close(tokChan)

	tokChan <- &Token{Type: TokenLineStart, Value: "VERY-BEGINNING-OF-FILE"}

	for {
		eatWhiteSpace(l.r)
		c := l.peek()

		switch {
		case c == '!':
			// eat everything until newline.
			l.EatComment()

		case c == '\n':
			tokChan <- l.LineStart()

		case unicode.IsLetter(c) || unicode.IsDigit(c) || unicode.IsPunct(c):
			tokChan <- l.ReadAtom()
		default:
			return
		}
	}
}

type Config struct {
	Interfaces map[string]string
}

func Parse(r io.Reader) *Config {

	rc := Config{
		Interfaces: make(map[string]string),
	}

	l := NewLexer(r)

	tokChan := l.Lex()

	tokBuf := []*Token{}
	cmdBuf := []*Token{}

	handler := map[string]func([]*Token){
		"interface": func(tok []*Token) {
			log.Printf("HANDING INTERFACE: %#v", tok)
			for _, t := range tok {
				log.Printf("%s", t.Value)
			}
		},
	}

	for tok := range tokChan {
		if tok == nil {
			continue
		}
		tokBuf = append(tokBuf, tok)

		if len(tokBuf) < 2 {
			continue
		}

		cmdBuf = append(cmdBuf, tok)

		switch {
		case tok.Type == TokenLineStart:
			// we should have a complete command at this point
			cmd := cmdBuf
			cmdBuf = cmdBuf[0:0]

			h, found := handler[cmd[0].Value]
			if found == false {
				// no handler found
				continue
			}
			h(cmd)
		}
	}

	return &rc
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r: bufio.NewReader(r),
	}
}

func (l Lexer) Lex() <-chan *Token {
	rc := make(chan *Token)
	go l.lex(rc)
	return rc
}
