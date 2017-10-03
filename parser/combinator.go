package parser

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Input is the input to a parser.
type Input string

// Cons returns the first rune and the remaining input.
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

func (i Input) cut(n int) Input {
	j := strings.IndexAny(string(i), " \n\t\r")
	if j < 0 {
		return Input([]rune(string(i))[:n])
	}
	return Input([]rune(string(i))[:j])
}

// Sample returns a string consisting of the next `n` characters or as many as
// are left in the input. If `n` is smaller than `len(i)`, the result string
// will be ellipsized. This is mostly just useful for error messaging.
func (i Input) Sample(n int) string {
	if n >= len(i) {
		return string(i)
	}
	return string([]rune(string(i))[:n]) + "..."
}

// MapFunc maps a parse result value from one value/type to another.
type MapFunc func(interface{}) interface{}

// MapSliceFunc maps a parse result slice value from one type to another
type MapSliceFunc func([]interface{}) interface{}

// Result represents the result of a parse.
type Result struct {
	// ParserName is the name of the parser that produced the result.
	ParserName string

	// Value is an optional value that the parser can return. Typically this
	// will be an AST node. This should be ignored if Result.Err != nil.
	Value interface{}

	// Rest is the input from the parser which the parser has not consumed.
	// If Err != nil, Rest should be set to the input of the parser which
	// failed, providing context about the failure; however, it's each parser's
	// responsibility to implement this convention.
	Rest Input

	// Err is set if the parser encountered a syntax error. Value should be
	// ignored if this is non-nil.
	Err error
}

// Map applies `f` to `r.Value` if `r` is valid; otherwise it short circuits
// and returns `r` directly
func (r Result) Map(f func(interface{}) interface{}) Result {
	if r.Err != nil {
		return r
	}
	return Result{
		ParserName: r.ParserName,
		Value:      f(r.Value),
		Rest:       r.Rest,
		Err:        r.Err,
	}
}

// Then applies `p` to `r.Rest` if `r` is valid; otherwise it short circuits
// and returns `r` directly.
func (r Result) Then(p Parser) Result {
	if r.Err != nil {
		return r
	}
	return p(r.Rest)
}

// OK returns a valid result from the provided value and input. The result's
// parser name will be that of OK's caller.
func OK(value interface{}, rest Input) Result {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return Result{"???", value, rest, nil}
	}
	return Result{runtime.FuncForPC(pc).Name(), value, rest, nil}
}

// ERR returns an error result from the provided error and input. The result's
// parser name will be that of ERR's caller. The "Rest" field will contain the
// value passed as input, which should be the input provided to the parser when
// it failed.
func ERR(err error, input Input) Result {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return Result{"???", nil, "", err}
	}
	return Result{runtime.FuncForPC(pc).Name(), nil, input, err}
}

// Error implements the error interface for Result. This should only be used to
// nest a Result as the error of another result, and so build a sort of parser
// stack trace. Indeed, Error returns a string containing the result's parser
// information and the error it encountered, which is always a Result until the
// very deepest error.
func (r Result) Error() string {
	return fmt.Sprintf(
		"%s(%#v):\n%s",
		r.ParserName,
		r.Rest.Sample(15),
		r.Err,
	)
}

// Rename takes a name and returns a copy of the original result except with
// the ParserName field set to `name`.
func (r Result) Rename(name string) Result {
	return Result{ParserName: name, Value: r.Value, Rest: r.Rest, Err: r.Err}
}

// Recover takes an input returns the original Result if it was successful,
// otherwise it returns a new successful Result with the same ParserName but
// Value is set to nil and Rest is set to `input`.
func (r Result) Recover(input Input) Result {
	if r.Err != nil {
		return Result{r.ParserName, nil, input, nil}
	}
	return r
}

// Parser is a function that takes input and returs a parse Result.
type Parser func(input Input) Result

// Then takes a parser and returns the result of the original parser followed
// by the input parser. If the original parser fails, its result is returned
// immediately, short-circuiting the subsequent parser.
func (p Parser) Then(next Parser) Parser {
	return func(input Input) Result { return p(input).Then(next) }
}

// Rename takes a name and returns a parser whose results' ParserName fields
// are set to `name`.
func (p Parser) Rename(name string) Parser {
	return func(input Input) Result { return p(input).Rename(name) }
}

// Map returns a parser that calls the original parser and then maps the result
// value using `f`.
func (p Parser) Map(f MapFunc) Parser {
	return func(input Input) Result { return p(input).Map(f) }
}

// Wrap returns a parser whose resultant caller information are those of Wrap's
// own caller. So if Foo() calls Wrap(), the returned parser will produce
// Results with a ParserName of "Foo".
func (p Parser) Wrap() Parser {
	name := "???"
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		name = runtime.FuncForPC(pc).Name()
	}
	return p.Rename(name)
}

// MapSlice is a convenience wrapper around Map which first asserts that the
// result value is a []interface{} before passing its arguments into the
// provided MapSliceFunc. If the result value is not a []interface{}, MapSlice
// panics.
func (p Parser) MapSlice(f MapSliceFunc) Parser {
	return p.Map(func(v interface{}) interface{} {
		return f(v.([]interface{}))
	})
}

// Get takes an integer index `i` and returns a parser which assumes the input
// parser returns a result value with type []interface{} and returns a result
// whose value is at position `i` in the slice. If the original parser returned
// an error result, the result is returned immediately. If the result was
// successful, but the value is not a slice, Get panics. If the result is
// successful and the value is a slice whose length is less than or equal to
// `i`, Get panics.
func (p Parser) Get(i int) Parser {
	return p.MapSlice(func(vs []interface{}) interface{} { return vs[i] })
}

func lit(r rune, input Input) Result {
	head, tail := input.Cons()
	if head == r {
		return OK(head, tail)
	}
	return ERR(
		fmt.Errorf("Wanted %#v, got %#v", string(r), string(head)),
		input,
	)
}

func strlit(s string, input Input) Result {
	if !strings.HasPrefix(string(input), s) {
		return ERR(
			fmt.Errorf("Wanted '%s'; got '%s'", s, input.cut(len(s))),
			input,
		)
	}
	return OK(s, input[len(s):])
}

// Lit returns a parser that expects the first rune to match `r`. On success,
// `r` is returned as the result value.
func Lit(r rune) Parser {
	return Parser(func(input Input) Result { return lit(r, input) }).Wrap()
}

// NotLit takes a rune `r` and returns a Parser that expects the first rune of
// its input to be anything except `r`. If successful, it returns the non-`r`
// rune as the result value.
func NotLit(r rune) Parser {
	return Parser(func(input Input) Result {
		head, tail := input.Cons()
		if head == r {
			return ERR(
				fmt.Errorf(
					"Wanted anything but '%s', but got it anyway",
					string(r),
				),
				input,
			)
		}
		return OK(head, tail)
	}).Wrap()
}

// StrLit returns a parser that expects the input to have the prefix `s`. On
// success, the result value is `s`.
func StrLit(s string) Parser {
	return Parser(func(input Input) Result { return strlit(s, input) }).Wrap()
}

// Seq returns a parser that expects the input to contain all of its input
// parsers in order. If successful, the result value will be a slice of values,
// one per parser, in parser order.
func Seq(parsers ...Parser) Parser {
	return func(input Input) Result {
		r := Result{Rest: input}
		values := make([]interface{}, len(parsers))
		for i, parser := range parsers {
			if r = parser(r.Rest); r.Err != nil {
				return ERR(r, input)
			}
			values[i] = r.Value
		}
		return OK(values, r.Rest)
	}
}

// UnicodeClass represents a class of unicode characters.
type UnicodeClass struct {
	// Name is the name of the unicode class
	Name string

	// RangeTable is the unicode range table
	RangeTable *unicode.RangeTable
}

var (
	// UnicodeClassDigit is the "Digit" unicode class
	UnicodeClassDigit = UnicodeClass{"Digit", unicode.Digit}

	// UnicodeClassLetter is the "Letter" unicode class
	UnicodeClassLetter = UnicodeClass{"Letter", unicode.Letter}

	// UnicodeClassSpace is the "Space" unicode class
	UnicodeClassSpace = UnicodeClass{"Space", unicode.Space}

	// UnicodeClassWhiteSpace is the "WhiteSpace" unicode class
	UnicodeClassWhiteSpace = UnicodeClass{
		"WhiteSpace",
		unicode.Pattern_White_Space,
	}
)

// IsClass returns a parser that expects its first rune to be of unicode class
// `rt`.
func IsClass(class UnicodeClass) Parser {
	return Parser(func(input Input) Result {
		head, rest := input.Cons()
		if unicode.Is(class.RangeTable, head) {
			return OK(head, rest)
		}
		return ERR(
			fmt.Errorf("Wanted <%s>; got %#v", class.Name, string(head)),
			input,
		)
	}).Wrap()
}

// Repeat takes a parser and continues to invoke it until the input fails to
// match. It always succeeds, and it returns a result value that is a slice of
// values, one per successful invocation of the input parser. If the input
// parser never succeeds, the result value will be an empty slice and the
// result's Err field will be nil. It is similar to OneOrMore except that it
// always succeeds.
func Repeat(p Parser) Parser {
	return Parser(func(input Input) Result {
		var values []interface{}
		for {
			r := p(input)
			if r.Err != nil {
				return OK(values, input)
			}
			values = append(values, r.Value)
			input = r.Rest
		}
	}).Wrap()
}

// OneOrMore takes a parser and expects at least one consecutive match.
// It is similar to Repeat with the exception that OneOrMore will fail if the
// first attempt fails. If successful, it will return a slice of values, one
// for each consecutive parser match.
func OneOrMore(p Parser) Parser {
	return Seq(p, Repeat(p)).Map(func(v interface{}) interface{} {
		values := v.([]interface{})
		head, tail := values[0], values[1]
		return append([]interface{}{head}, tail.([]interface{})...)
	}).Wrap()
}

// Opt takes an input parser and returns a parser that attempts to invoke the
// input parser, returning its result on success or recovering on failure and
// returning a successful result (r.Err == nil), no value, and r.Rest is set to
// the original input (the parser doesn't advance).
func Opt(p Parser) Parser {
	return Parser(func(input Input) Result {
		return p(input).Recover(input)
	}).Wrap()
}

// Any takes a list of input parsers and returns a parser which tries each
// input parser until it finds a match. If it finds a match, it returns the
// result, otherwise it returns an error result.
func Any(parsers ...Parser) Parser {
	return Parser(func(input Input) Result {
		parserNames := make([]string, len(parsers))
		for i, p := range parsers {
			r := p(input)
			if r.Err != nil {
				parserNames[i] = r.ParserName
				continue
			}
			return r
		}

		return ERR(
			fmt.Errorf(
				"Failed to match parsers: [%s]",
				strings.Join(parserNames, ", "),
			),
			input,
		)
	}).Wrap()
}

func collectRunes(vs []interface{}) interface{} {
	runes := make([]rune, len(vs))
	for i, v := range vs {
		runes[i] = v.(rune)
	}
	return string(runes)
}

var (
	// WS is a parser in the form OneOrMore(IsClass(UnicodeClassWhiteSpace)).
	WS = OneOrMore(IsClass(UnicodeClassWhiteSpace)).
		MapSlice(collectRunes).
		Rename("WS")

	// CanWS is a parser in the form Repeat(IsClass(UnicodeClassWhiteSpace))
	CanWS = Repeat(IsClass(UnicodeClassWhiteSpace)).
		MapSlice(collectRunes).
		Rename("CanWS")

	// Digits is a parser in the form OneOrMore(IsClass(UnicodeClassDigit))
	Digits = OneOrMore(IsClass(UnicodeClassDigit)).
		MapSlice(collectRunes).
		Rename("Digits")

	// Letters is a parser in the form OneOrMore(IsClass(UnicodeClassLetter))
	Letters = OneOrMore(IsClass(UnicodeClassLetter)).
		MapSlice(collectRunes).
		Rename("Letters")

	// String is a parser that matches single-line string literals in source
	// code
	String = Seq(Lit('"'), Repeat(NotLit('"')), Lit('"')).
		Get(1).
		MapSlice(collectRunes).
		Rename("String")

	// Int is a parser that matches decimal integers in source code
	Int = OneOrMore(IsClass(UnicodeClassDigit)).
		MapSlice(collectRunes).
		Map(func(v interface{}) interface{} {
			i, err := strconv.Atoi(v.(string))
			if err != nil {
				panic("Invalid integer: " + v.(string))
			}
			return i
		}).
		Rename("Int")

	// Ident is a parser that matches identifiers in source code
	Ident = Seq(
		Any(
			Lit('_'),
			IsClass(UnicodeClassLetter),
		),
		Repeat(Any(
			Lit('_'),
			IsClass(UnicodeClassLetter),
			IsClass(UnicodeClassDigit),
		)),
	).MapSlice(func(vs []interface{}) interface{} {
		runes := []rune{vs[0].(rune)}
		for _, v := range vs[1].([]interface{}) {
			runes = append(runes, v.(rune))
		}
		return string(runes)
	}).Rename("Ident")

	// EOF is a parser that matches the zero-value rune, which is simply this
	// library's convention for signaling end-of-file.
	EOF = Lit(0).Rename("EOF")

	// EOS is a parser that matches the end-of-statement semi-colon, as well as
	// leading and trailing whitespace.
	EOS = Seq(CanWS, Lit(';'), CanWS).Rename("EOS")
)
