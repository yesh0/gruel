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

func writeC(b *bytes.Buffer, indent string) {
	lines := make(Pairs, 0, 2*len(ir.Operators))
	for name, ops := range ir.Operators {
		for _, op := range ops {
			line := strings.Builder{}
			line.WriteString(fmt.Sprintf("%s// `%s`(%d)\n", indent, name, op.Argc))
			line.WriteString(indent)
			switch {
			case op.Argc == 1 && op.JitFunction[0] != ':':
				line.WriteString(fmt.Sprintf("UNARY_OP (0x%02x, %s);",
					op.Opcode, op.JitFunction))
			case op.Argc == 2 && op.JitFunction[0] != ':':
				line.WriteString(fmt.Sprintf("BINARY_OP(0x%02x, %s);",
					op.Opcode, op.JitFunction))
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
