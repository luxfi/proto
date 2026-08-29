// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package schema reads a .zap schema file.
//
// The generator (zap-proto/go cmd/zapgen) has a parser, but it lives in a
// main package and cannot be imported, so a schema shipped next to a codec
// is prose that nothing checks. This reader closes that: a test can load the
// .zap and hold the shipped codec to it.
//
// Grammar, as zapgen defines it:
//
//	file   := 'package' ident (alias | struct | interface)*
//	alias  := 'type' ident '=' type
//	struct := 'struct' ident '{' field* '}'
//	field  := ident type '@' int
//	type   := 'list' '<' type '>' | 'bytes_fixed' '[' int ']'
//	        | bool | u8 | u16 | u32 | u64 | i8 | i16 | i32 | i64
//	        | f32 | f64 | bytes | text | ident
//
// '#' runs to end of line. Interfaces are skipped: they declare calls, and
// this reader is about layout.
package schema

import (
	"fmt"
	"strconv"
	"strings"
)

// Kind is a field's declared type, spelled as the schema spells it.
type Kind string

// Width is the field's byte width in the fixed section, matching zapgen:
// scalars inline, bytes/text/list an {offset,length} pair, bytes_fixed[N]
// N bytes inline, a nested struct a single offset.
func (k Kind) Width(fixed int) int {
	switch k {
	case "bool", "u8", "i8":
		return 1
	case "u16", "i16":
		return 2
	case "u32", "i32", "f32":
		return 4
	case "u64", "i64", "f64":
		return 8
	case "bytes", "text", "list":
		return 8
	case "bytes_fixed":
		return fixed
	}
	return 4 // nested struct
}

// Field is one declared field.
type Field struct {
	Name  string
	Kind  Kind
	Fixed int  // bytes_fixed[N]
	Elem  Kind // list<T>
	Off   int
}

// Struct is one declared struct. Size is max(off+width), which is what
// zapgen emits as the object's fixed-section size.
type Struct struct {
	Name   string
	Fields []Field
}

func (s *Struct) Size() int {
	n := 0
	for _, f := range s.Fields {
		if end := f.Off + f.Kind.Width(f.Fixed); end > n {
			n = end
		}
	}
	return n
}

// Field returns the named field, or false.
func (s *Struct) Field(name string) (Field, bool) {
	for _, f := range s.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// File is a parsed schema.
type File struct {
	Package string
	Structs []*Struct
}

// Struct returns the named struct, or false.
func (f *File) Struct(name string) (*Struct, bool) {
	for _, s := range f.Structs {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

// Parse reads a .zap source file.
func Parse(src []byte) (*File, error) {
	var (
		file    = &File{}
		aliases = map[string]Field{}
		cur     *Struct
	)
	for n, raw := range strings.Split(string(src), "\n") {
		line := raw
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		at := func(err error) error { return fmt.Errorf("line %d: %w", n+1, err) }

		switch {
		case cur != nil && words[0] == "}":
			file.Structs, cur = append(file.Structs, cur), nil

		case cur != nil:
			f, err := parseField(words, aliases)
			if err != nil {
				return nil, at(err)
			}
			cur.Fields = append(cur.Fields, f)

		case words[0] == "package":
			if len(words) != 2 {
				return nil, at(fmt.Errorf("package takes one name"))
			}
			file.Package = words[1]

		case words[0] == "type":
			// type name = T
			if len(words) != 4 || words[2] != "=" {
				return nil, at(fmt.Errorf("alias is `type NAME = TYPE`"))
			}
			f, err := parseType(words[3], aliases)
			if err != nil {
				return nil, at(err)
			}
			aliases[words[1]] = f

		case words[0] == "struct":
			if len(words) < 2 {
				return nil, at(fmt.Errorf("struct needs a name"))
			}
			cur = &Struct{Name: words[1]}

		case words[0] == "interface":
			return nil, at(fmt.Errorf("interface declarations are not read here"))

		default:
			return nil, at(fmt.Errorf("unexpected %q", words[0]))
		}
	}
	if cur != nil {
		return nil, fmt.Errorf("unterminated struct %q", cur.Name)
	}
	if file.Package == "" {
		return nil, fmt.Errorf("no package declaration")
	}
	return file, nil
}

// parseField reads `Name Type @ N`, with or without space around the @.
func parseField(words []string, aliases map[string]Field) (Field, error) {
	joined := strings.Join(words[1:], "")
	i := strings.IndexByte(joined, '@')
	if i < 0 {
		return Field{}, fmt.Errorf("field %q has no @offset", words[0])
	}
	f, err := parseType(joined[:i], aliases)
	if err != nil {
		return Field{}, err
	}
	off, err := strconv.Atoi(joined[i+1:])
	if err != nil {
		return Field{}, fmt.Errorf("field %q: bad offset: %w", words[0], err)
	}
	f.Name, f.Off = words[0], off
	return f, nil
}

func parseType(s string, aliases map[string]Field) (Field, error) {
	if a, ok := aliases[s]; ok {
		return a, nil
	}
	switch {
	case strings.HasPrefix(s, "list<"):
		if !strings.HasSuffix(s, ">") {
			return Field{}, fmt.Errorf("unterminated list<")
		}
		elem, err := parseType(s[len("list<"):len(s)-1], aliases)
		if err != nil {
			return Field{}, err
		}
		return Field{Kind: "list", Elem: elem.Kind, Fixed: elem.Fixed}, nil

	case strings.HasPrefix(s, "bytes_fixed["):
		if !strings.HasSuffix(s, "]") {
			return Field{}, fmt.Errorf("unterminated bytes_fixed[")
		}
		n, err := strconv.Atoi(s[len("bytes_fixed[") : len(s)-1])
		if err != nil || n <= 0 {
			return Field{}, fmt.Errorf("bytes_fixed[N] needs N > 0")
		}
		return Field{Kind: "bytes_fixed", Fixed: n}, nil
	}
	return Field{Kind: Kind(s)}, nil
}
