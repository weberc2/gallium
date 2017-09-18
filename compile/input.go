package main

import "encoding/json"

type Input interface {
	Next() Input
	Value() rune
	Take(n int) string
	More() bool
}

type StringInput struct {
	Runes  []rune
	Cursor int
}

func NewStringInput(s string) StringInput {
	return StringInput{Runes: []rune(s)}
}

func (si StringInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Runes  string
		Cursor int
	}{string(si.Runes), si.Cursor})
}

func (si StringInput) Next() Input {
	return StringInput{Runes: si.Runes, Cursor: si.Cursor + 1}
}

func (si StringInput) Take(n int) string {
	end := si.Cursor + n
	if end >= len(si.Runes) {
		end = len(si.Runes)
	}
	println("N:", n)
	println("CURSOR:", si.Cursor)
	println("END:", end)
	println("LEN:", len(si.Runes))
	return string(si.Runes[si.Cursor:end])
}

func (si StringInput) More() bool {
	return si.Cursor < len(si.Runes)
}

func (si StringInput) Value() rune {
	if si.Cursor >= len(si.Runes) {
		return rune(0)
	}
	return si.Runes[si.Cursor]
}
