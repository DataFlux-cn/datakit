// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package multiline wrap regexp/match functions
package multiline

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Multiline struct {
	patternRegexp *regexp.Regexp
	buff          bytes.Buffer
	lines         int
	maxLines      int

	// prefixSpace 用以标记 pattern 为空的情况，属于默认行为，
	// 如果一行数据，它的首字符是 WhiteSpace，那它就是多行
	// WhiteSpace 定义为 '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0
	prefixSpace bool
}

func New(pattern string, maxLines int) (*Multiline, error) {
	var r *regexp.Regexp
	var err error

	if pattern == "" {
		return &Multiline{maxLines: maxLines, prefixSpace: true}, nil
	}

	if r, err = regexp.Compile(pattern); err != nil {
		return nil, err
	}
	return &Multiline{
		patternRegexp: r,
		maxLines:      maxLines,
	}, nil
}

var newline = []byte{'\n'}

func (m *Multiline) ProcessLine(text []byte) []byte {
	if !m.match(text) {
		if m.lines != 0 {
			m.buff.Write(newline)
		}
		m.buff.Write(text)
		m.lines++

		if m.lines >= m.maxLines {
			return m.Flush()
		}
		return nil
	}

	previousText := m.Flush()
	m.buff.Write(text)
	m.lines = 1 // always 1th line

	return previousText
}

func (m *Multiline) ProcessLineString(text string) string {
	return m.ProcessLineStringWithFlag(text, false)
}

// ProcessLineStringWithFlag process line string with multiFlag
// when the multiFlag is true, it indicates the text is part of multiline and ignore the match pattern,
// and will not insert '\n' before the text.
func (m *Multiline) ProcessLineStringWithFlag(text string, multiFlag bool) string {
	if multiFlag || !m.MatchString(text) {
		if m.lines != 0 && !multiFlag {
			m.buff.WriteString("\n")
		}
		m.buff.WriteString(text)
		m.lines++

		if m.lines >= m.maxLines {
			return m.FlushString()
		}
		return ""
	}

	previousText := m.FlushString()
	m.buff.WriteString(text)
	m.lines = 1 // always 1th line

	return previousText
}

func (m *Multiline) CacheLines() int {
	return m.lines
}

func (m *Multiline) Flush() []byte {
	if m.buff.Len() == 0 {
		return nil
	}

	text := make([]byte, m.buff.Len())
	copy(text, m.buff.Bytes())

	m.buff.Reset()
	m.lines = 0

	return text
}

func (m *Multiline) match(text []byte) bool {
	if m.prefixSpace {
		return m.matchOfPrefixSpace(text)
	}
	return m.patternRegexp.Match(text)
}

func (m *Multiline) matchOfPrefixSpace(text []byte) bool {
	if len(text) == 0 {
		return true
	}
	return !unicode.IsSpace(rune(text[0]))
}

func (m *Multiline) FlushString() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	m.lines = 0
	return text
}

func (m *Multiline) MatchString(text string) bool {
	if m.prefixSpace {
		return m.MatchStringOfPrefixSpace(text)
	}
	return m.patternRegexp.MatchString(text)
}

func (m *Multiline) MatchStringOfPrefixSpace(text string) bool {
	if len(text) == 0 {
		return true
	}
	return !unicode.IsSpace(rune(text[0]))
}

func (m *Multiline) BuffString() string {
	return m.buff.String()
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func TrimRightSpace(s string) string {
	end := len(s)
	for ; end > 0; end-- {
		c := s[end-1]
		if c >= utf8.RuneSelf {
			return strings.TrimFunc(s[:end], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}
	return s[:end]
}
