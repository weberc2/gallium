package main

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

const indent = "    "

type Node interface {
	Equal(Node) bool
	Pretty(base string) string
}

type Nodes []Node

func (nodes Nodes) Equal(other Node) bool {
	if otherNodes, ok := other.(Nodes); ok && len(nodes) == len(otherNodes) {
		for i, n := range otherNodes {
			if nodes[i] != n {
				return false
			}
		}
		return true
	}
	return false
}

func (nodes Nodes) Pretty(base string) string {
	nodeStrings := make([]string, len(nodes))
	for i, node := range nodes {
		nodeStrings[i] = "\n" + base + indent + node.Pretty(
			base+indent+indent,
		)
	}
	return "[" + strings.Join(nodeStrings, "") + "\n" + base + "]"
}

type Input string

func (i Input) cut(n int) Input {
	j := strings.IndexAny(string(i), " \n\t\r")
	if j < 0 {
		return Input([]rune(string(i))[:n])
	}
	return Input([]rune(string(i))[:j])
}

func (i Input) Cons() (rune, Input) {
	r, sz := utf8.DecodeRuneInString(string(i))
	if r == utf8.RuneError {
		if sz == 0 { // According to utf8 docs, (RuneError, 0) means eof
			return rune(0), ""
		}
		// According to utf8 docs, (RuneError, 1) means invalid utf-8
		panic("Invalid utf-8")
	}
	return r, i[sz:]
}

func (i Input) Sample(n int) string {
	if n >= len(i) {
		return string(i)
	}
	return string([]rune(string(i))[:n]) + "..."
}

func (input Input) SkipWS() Input {
	for {
		head, tail := input.Cons()
		if !unicode.IsSpace(head) {
			return input
		}
		input = tail
	}
}

type Result struct {
	Parser string
	Node   Node
	Rest   Input
	Err    error
}

func (r *Result) IsErr() bool {
	return r.Err != nil
}

func (r *Result) Equal(other Result) bool {
	return r.Parser == other.Parser &&
		r.Rest == other.Rest &&
		r.Err == other.Err
}

func (r Result) Error() string {
	return fmt.Sprintf("%s(\"%s\"):\n%s", r.Parser, r.Rest.Sample(15), r.Err)
}

type Parser func(Input) Result

func Err(input Input, err error) Result {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return Result{"???", nil, input, err}
	}

	return Result{runtime.FuncForPC(pc).Name(), nil, input, err}
}

func Ok(n Node, input Input) Result {
	r := Err(input, nil)
	r.Node = n
	return r
}

func lit(r rune, input Input) Result {
	head, tail := input.Cons()
	if head == r {
		return Ok(nil, tail)
	}
	return Err(
		input,
		fmt.Errorf("Wanted '%s', got '%s'", string(r), string(head)),
	)
}

func strlit(s string, input Input) Result {
	if !strings.HasPrefix(string(input), s) {
		return Err(
			input,
			fmt.Errorf("Wanted '%s'; got '%s'", s, input.cut(len(s))),
		)
	}
	return Ok(String(s), input[len(s):])
}

func Rename(name string, p Parser) Parser {
	return func(input Input) Result {
		r := p(input)
		r.Parser = name
		return r
	}
}

func Map(p Parser, f func(Node) Node) Parser {
	return func(input Input) Result {
		r := p(input)
		if !r.IsErr() {
			r.Node = f(r.Node)
		}
		return r
	}
}

func Seq(parsers ...Parser) Parser {
	return func(input Input) Result {
		rslt := Result{Rest: input}
		nodes := make(Nodes, len(parsers))
		for i, parser := range parsers {
			if rslt = parser(rslt.Rest); rslt.IsErr() {
				return Err(input, rslt)
			}
			nodes[i] = rslt.Node
		}
		return Ok(nodes, rslt.Rest)
	}
}

func Lit(r rune) Parser {
	return func(input Input) Result {
		return lit(r, input)
	}
}

func StrLit(s string) Parser {
	return func(input Input) Result {
		return strlit(s, input)
	}
}

type WS string

func (ws WS) Equal(other Node) bool {
	if otherWS, ok := other.(WS); ok {
		return ws == otherWS
	}
	return false
}

func (ws WS) Pretty(base string) string {
	x := string(ws)
	for s, char := range map[string]string{
		" ":  "·",
		"\n": "\\n",
		"\r": "\\r",
		"\t": "\\t",
	} {
		x = strings.Replace(x, s, char, -1)
	}
	return "WS(\"" + x + ")"
}

func (ws WS) String() string {
	return ws.Pretty("")
}

func CanWS(input Input) Result {
	last := input
	runes := []rune{}
	for {
		head, tail := last.Cons()
		if !unicode.IsSpace(head) {
			return Ok(WS(string(runes)), last)
		}
		runes = append(runes, head)
		last = tail
	}
}

func MustWS(input Input) Result {
	head, _ := input.Cons()
	if !unicode.IsSpace(head) {
		return Err(input, fmt.Errorf("Wanted <space>, got '%s'", string(head)))
	}
	return CanWS(input)
}

func Opt(p Parser) Parser {
	return func(input Input) Result {
		rslt := p(input)
		if rslt.IsErr() {
			return Ok(nil, input)
		}
		return rslt
	}
}

func Wrap(p Parser) Parser {
	return func(input Input) Result {
		r := p(input)
		if r.IsErr() {
			return Err(input, r)
		}
		return Ok(r.Node, r.Rest)
	}
}

func Any(parsers ...Parser) Parser {
	return Rename("Any", func(input Input) Result {
		for _, p := range parsers {
			r := p(input)
			if r.IsErr() {
				continue
			}
			return r
		}
		return Err(input, fmt.Errorf("Failed to match parsers"))
	})
}

func Repeat(p Parser) Parser {
	return func(input Input) Result {
		last := input
		nodes := Nodes{}
		for {
			r := p(last)
			if r.IsErr() {
				return Ok(nodes, last)
			}
			nodes = append(nodes, r.Node)
			last = r.Rest
		}
	}
}
