package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/yesh0/gruel/pkg/grueljit"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("Usage: %s <expr> [var1=value1] [var2=value2] [...]\n", os.Args[0])
	}
	expr := os.Args[1]
	types := make(map[string]byte, len(os.Args)-2)
	values := make(map[string]any, len(os.Args)-2)

	log.Println("Environment:")
	for _, arg := range os.Args[2:] {
		split := strings.Split(arg, "=")
		if len(split) != 2 {
			log.Fatal("Malformed pairs", arg)
		}
		k, v := split[0], split[1]
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		log.Println("    ", k, "=", v)
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatal(err)
		}
		types[k] = grueljit.TypeFloat
		values[k] = n
	}
	log.Println("Evaluating:\n    ", expr)

	f, err := grueljit.Compile(expr, types)
	if err != nil {
		log.Fatal(err)
	}
	out, err := f.Call(values)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Result:\n    ", out)
}
