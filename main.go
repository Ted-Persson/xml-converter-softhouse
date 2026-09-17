// people-converter läser det radbaserade personformatet och skriver
// motsvarande XML.
// go run . testdata/example.txt > testdata/expected.xml
// people-converter [indatafil] > utdata.xml
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"people-converter/internal/people"
)

// main delegerar till run och sätter exitkod vid fel.
func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "fel:", err)
		os.Exit(1)
	}
}

// run läser indata, kör parser -> XML-writer, och skriver till stdout.
func run(args []string, stdin io.Reader, stdout io.Writer) error {
	in := stdin

	switch len(args) {
	case 0:
		// läser stdin
	case 1:
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	default:
		return fmt.Errorf("användning: people-converter [indatafil]")
	}

	model, err := people.Parse(in)
	if err != nil {
		return err
	}

	out := bufio.NewWriter(stdout)
	if err := people.WriteXML(out, model); err != nil {
		return err
	}
	return out.Flush()
}