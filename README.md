# mozlz4

[![GitHub Releases](https://img.shields.io/github/v/release/cions/mozlz4?sort=semver)](https://github.com/cions/mozlz4/releases)
[![LICENSE](https://img.shields.io/github/license/cions/mozlz4)](https://github.com/cions/mozlz4/blob/master/LICENSE)
[![CI](https://github.com/cions/mozlz4/actions/workflows/ci.yml/badge.svg)](https://github.com/cions/mozlz4/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cions/mozlz4.svg)](https://pkg.go.dev/github.com/cions/mozlz4)
[![Go Report Card](https://goreportcard.com/badge/github.com/cions/mozlz4)](https://goreportcard.com/report/github.com/cions/mozlz4)

Compress or decompress mozlz4 files.

## Usage

```sh
$ mozlz4 --help
Usage: mozlz4 [OPTIONS] [FILE...]

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
```

## Installation

[Download from GitHub Releases](https://github.com/cions/mozlz4/releases)

### Build from source

```sh
$ go install github.com/cions/mozlz4/cmd/mozlz4@latest
```

## License

MIT
