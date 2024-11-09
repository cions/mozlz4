// Copyright (c) 2024 cions
// Licensed under the MIT License. See LICENSE for details.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/cions/go-options"
	"github.com/cions/mozlz4"
	"golang.org/x/term"
)

var NAME = "mozlz4"
var VERSION = "(devel)"
var USAGE = `Usage: $NAME [OPTIONS] [FILE...]

Compress or decompress mozlz4 files.

Options:
  -z, --compress        Force compression
  -d, --decompress      Force decompression
  -c, --stdout          Write to the standard output and keep the input files
  -o, --output=FILE     Write to the FILE
  -S, --suffix=SUFFIX   Add SUFFIX on compressed file names
  -k, --keep            Don't delete the input files (default)
      --rm              Delete the input files after successful (de)compression
  -f, --force           Allow overwriting existing files, reading input from
                        a terminal, writing compressed data to a terminal
  -h, --help            Show this help message and exit
      --version         Show version information and exit
`

type Command struct {
	ForceCompress   bool
	ForceDecompress bool
	Destination     string
	Suffix          string
	Delete          bool
	Force           bool
}

func (c *Command) Kind(name string) options.Kind {
	switch name {
	case "-z", "--compress":
		return options.Boolean
	case "-d", "--decompress":
		return options.Boolean
	case "-c", "--stdout":
		return options.Boolean
	case "-o", "--output":
		return options.Required
	case "-S", "--suffix":
		return options.Required
	case "-k", "--keep":
		return options.Boolean
	case "--rm":
		return options.Boolean
	case "-f", "--force":
		return options.Boolean
	case "-h", "--help":
		return options.Boolean
	case "--version":
		return options.Boolean
	default:
		return options.Unknown
	}
}

func (c *Command) Option(name string, value string, hasValue bool) error {
	switch name {
	case "-z", "--compress":
		c.ForceCompress = true
		c.ForceDecompress = false
	case "-d", "--decompress":
		c.ForceCompress = false
		c.ForceDecompress = true
	case "-c", "--stdout":
		c.Destination = "-"
	case "-o", "--output":
		c.Destination = value
	case "-S", "--suffix":
		c.Suffix = value
	case "-k", "--keep":
		c.Delete = false
	case "--rm":
		c.Delete = true
	case "-f", "--force":
		c.Force = true
	case "-h", "--help":
		return options.ErrHelp
	case "--version":
		return options.ErrVersion
	default:
		return options.ErrUnknown
	}
	return nil
}

func (c *Command) readFile(name string) ([]byte, error) {
	if name == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(name)
}

func (c *Command) writeFile(name string, data []byte) error {
	if name == "-" {
		if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
		return nil
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !c.Force {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(name, flags, 0o666)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		err2 := f.Close()
		return errors.Join(err, err2)
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Command) getDestination(name string, compress bool) (string, error) {
	if c.Destination != "" {
		return c.Destination, nil
	}
	if name == "-" {
		return "-", nil
	}
	if compress {
		return name + c.Suffix, nil
	}

	ext := filepath.Ext(name)
	if base := filepath.Base(name); ext == base {
		ext = ""
	}
	if ext == "" {
		return "", fmt.Errorf("%v has no extension. use -o/--output option.", name)
	}
	if dest, found := strings.CutSuffix(name, ext); !found {
		return "", fmt.Errorf("%v has no extension. use -o/--output option.", name)
	} else {
		return dest, nil
	}
}

func (c *Command) processFile(file string) error {
	if !c.Force && file == "-" && term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("stdin is a terminal")
	}

	input, err := c.readFile(file)
	if err != nil {
		return err
	}

	compress := c.ForceCompress || (!c.ForceDecompress && !bytes.HasPrefix(input, mozlz4.HEADER))
	var output []byte
	if compress {
		output, err = mozlz4.Compress(input)
	} else {
		output, err = mozlz4.Decompress(input)
	}
	if err != nil {
		if file == "-" {
			return fmt.Errorf("<stdin>: %w", err)
		} else {
			return fmt.Errorf("%v: %w", file, err)
		}
	}

	dest, err := c.getDestination(file, compress)
	if err != nil {
		return err
	}
	if !c.Force && compress && dest == "-" && term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("stdout is a terminal")
	}

	if err := c.writeFile(dest, output); err != nil {
		return err
	}

	if c.Delete && file != dest && file != "-" && dest != "-" {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	return nil
}

func run(args []string) error {
	c := &Command{
		Suffix: ".mozlz4",
	}

	files, err := options.Parse(c, args)
	if errors.Is(err, options.ErrHelp) {
		usage := strings.ReplaceAll(USAGE, "$NAME", NAME)
		fmt.Print(usage)
		return nil
	} else if errors.Is(err, options.ErrVersion) {
		version := VERSION
		if bi, ok := debug.ReadBuildInfo(); ok {
			version = bi.Main.Version
		}
		fmt.Printf("%v %v\n", NAME, version)
		return nil
	} else if err != nil {
		return err
	} else if len(files) > 1 && c.Destination != "" && c.Destination != "-" {
		return errors.New("-o/--output cannot be used if multiple input files is given")
	}

	if len(files) == 0 {
		return c.processFile("-")
	}

	for _, file := range files {
		if err := c.processFile(file); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v: error: %v\n", NAME, err)
		os.Exit(1)
	}
}
