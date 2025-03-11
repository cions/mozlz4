// Copyright (c) 2024-2025 cions
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
                        a terminal, and writing compressed data to a terminal
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

func (cmd *Command) Kind(name string) options.Kind {
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

func (cmd *Command) Option(name string, value string, hasValue bool) error {
	switch name {
	case "-z", "--compress":
		cmd.ForceCompress = true
		cmd.ForceDecompress = false
	case "-d", "--decompress":
		cmd.ForceCompress = false
		cmd.ForceDecompress = true
	case "-c", "--stdout":
		cmd.Destination = "-"
	case "-o", "--output":
		cmd.Destination = value
	case "-S", "--suffix":
		cmd.Suffix = value
	case "-k", "--keep":
		cmd.Delete = false
	case "--rm":
		cmd.Delete = true
	case "-f", "--force":
		cmd.Force = true
	case "-h", "--help":
		return options.ErrHelp
	case "--version":
		return options.ErrVersion
	default:
		return options.ErrUnknown
	}
	return nil
}

func (cmd *Command) readFile(name string) ([]byte, error) {
	if name == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(name)
}

func (cmd *Command) writeFile(name string, data []byte) error {
	if name == "-" {
		if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
		return nil
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !cmd.Force {
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

func (cmd *Command) getDestination(name string, compress bool) (string, error) {
	if cmd.Destination != "" {
		return cmd.Destination, nil
	}
	if name == "-" {
		return "-", nil
	}
	if compress {
		return name + cmd.Suffix, nil
	}
	ext := filepath.Ext(name)
	if base := filepath.Base(name); ext == base {
		ext = ""
	}
	if ext == "" {
		return "", fmt.Errorf("%v has no extension. use -o/--output option.", name)
	}
	return strings.TrimSuffix(name, ext), nil
}

func (cmd *Command) processFile(file string) error {
	if !cmd.Force && file == "-" && term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("the standard input is a terminal")
	}

	input, err := cmd.readFile(file)
	if err != nil {
		return err
	}

	compress := cmd.ForceCompress || (!cmd.ForceDecompress && !bytes.HasPrefix(input, mozlz4.HEADER))

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

	dest, err := cmd.getDestination(file, compress)
	if err != nil {
		return err
	}
	if !cmd.Force && compress && dest == "-" && term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("the standard output is a terminal")
	}

	if err := cmd.writeFile(dest, output); err != nil {
		return err
	}

	if cmd.Delete && file != dest && file != "-" && dest != "-" {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	return nil
}

func run(args []string) error {
	cmd := &Command{
		Suffix: ".mozlz4",
	}

	files, err := options.Parse(cmd, args)
	switch {
	case errors.Is(err, options.ErrHelp):
		usage := strings.ReplaceAll(USAGE, "$NAME", NAME)
		fmt.Print(usage)
		return nil
	case errors.Is(err, options.ErrVersion):
		version := VERSION
		if bi, ok := debug.ReadBuildInfo(); ok {
			version = bi.Main.Version
		}
		fmt.Printf("%v %v\n", NAME, version)
		return nil
	case err != nil:
		return err
	case len(files) > 1 && cmd.Destination != "" && cmd.Destination != "-":
		return options.Errorf("-o/--output cannot be used if multiple input files is given")
	}

	if len(files) == 0 {
		if err := cmd.processFile("-"); err != nil {
			return err
		}
	} else {
		for _, file := range files {
			if err := cmd.processFile(file); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v: error: %v\n", NAME, err)
		if errors.Is(err, options.ErrCmdline) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
