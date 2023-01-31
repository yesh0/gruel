package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/yesh0/gruel/internal/ir"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage:\n    %s <C_file>", os.Args[0])
	}
	cFile := os.Args[1]

	validate()
	populateC(cFile)
}

func populateC(cFile string) {
	if f, err := os.Open(cFile); err != nil {
		log.Fatalln(err)
	} else {
		s := bufio.NewScanner(f)
		output := bytes.Buffer{}
		populated := false
		for s.Scan() {
			line := string(s.Bytes())
			if strings.Contains(line, "//@start maintained by operators.go") {
				end := ""
				for s.Scan() {
					next := string(s.Bytes())
					if strings.Contains(next, "//@end maintained by operators.go") {
						end = next
						break
					}
				}
				if end == "" {
					log.Fatalln("no maintained block found")
				}
				if populated {
					log.Fatalln("multiple maintained blocks")
				}
				populated = true
				output.WriteString(line)
				output.WriteRune('\n')
				indentEnd := strings.IndexFunc(line,
					func(r rune) bool { return !unicode.IsSpace(r) })
				writeC(&output, line[0:indentEnd])
				output.WriteString(end)
				output.WriteRune('\n')
			} else {
				output.WriteString(line)
				output.WriteRune('\n')
			}
		}
		if err := f.Close(); err != nil {
			log.Fatalln(err)
		}
		os.WriteFile(cFile, output.Bytes(), 0644)
	}
}

func validate() {
	opcodes := make(map[int]string)
	for name, ops := range ir.Operators {
		for _, op := range ops {
			other, ok := opcodes[op.Opcode]
			if ok {
				log.Fatalf("conflicting opcode for %s and %s", name, other)
			}
			opcodes[op.Opcode] = name
		}
	}
}

type Pair struct {
	opcode int
	line   string
}

type Pairs []Pair

func (p Pairs) Len() int {
	return len(p)
}
func (p Pairs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p Pairs) Less(i, j int) bool {
	return p[i].opcode < p[j].opcode
}

var type_map = map[string]string{
	"i": "long",
	"s": "void_ptr",
}

func writeC(b *bytes.Buffer, indent string) {
	lines := make(Pairs, 0, 2*len(ir.Operators))
	for name, ops := range ir.Operators {
		for _, op := range ops {
			line := strings.Builder{}
			line.WriteString(fmt.Sprintf("%s// `%s`(%d)\n", indent, name, op.Argc))
			line.WriteString(indent)
			switch {
			case op.Argc == 1 && !unicode.IsPunct(rune(op.JitFunction[0])):
				line.WriteString(fmt.Sprintf("UNARY_OP (0x%02x, %s);",
					op.Opcode, op.JitFunction))
			case op.Argc == 2 && !unicode.IsPunct(rune(op.JitFunction[0])):
				line.WriteString(fmt.Sprintf("BINARY_OP(0x%02x, %s);",
					op.Opcode, op.JitFunction))
			case op.Argc == 2 && op.JitFunction[0] == '!':
				line.WriteString(fmt.Sprintf("LOGIC_OP (0x%02x, %s);",
					op.Opcode, op.JitFunction[1:]))
			case op.Argc == 1 && op.JitFunction[0] == ':':
				fields := strings.Split(op.JitFunction, ":")
				if len(fields) != 3 {
					log.Fatalf("invalid function %s\n", op.JitFunction)
				}
				typeName, ok := type_map[fields[1]]
				if !ok {
					log.Fatalf("invalid type %s\n", op.JitFunction)
				}
				line.WriteString(fmt.Sprintf("UNSTRING_OP(0x%02x, %s, %s);",
					op.Opcode, fields[2], typeName))
			case op.Argc == 2 && op.JitFunction[0] == ':':
				fields := strings.Split(op.JitFunction, ":")
				if len(fields) != 4 {
					log.Fatalf("invalid function %s\n", op.JitFunction)
				}
				for _, v := range fields[1:3] {
					_, ok := type_map[v]
					if !ok {
						log.Fatalf("invalid type %s in %s\n", v, op.JitFunction)
					}
				}
				line.WriteString(fmt.Sprintf("BISTRING_OP(0x%02x, %s, %s, %s);",
					op.Opcode, fields[3], type_map[fields[1]], type_map[fields[2]]))
			default:
				log.Fatalf("unrecognized operator %s:%d", name, op.Opcode)
			}
			line.WriteRune('\n')
			lines = append(lines, Pair{opcode: op.Opcode, line: line.String()})
		}
	}
	sort.Sort(lines)
	for _, v := range lines {
		b.WriteString(v.line)
	}
}
