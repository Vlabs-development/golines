package golines

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func Format(config ShortenerConfig, paths []string) error {
	shortener := NewShortener(config)

	if len(paths) == 0 {
		// Read input from stdin
		contents, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		result, err := shortener.Shorten(contents)
		if err != nil {
			return err
		}
		err = handleOutput("", contents, result)
		if err != nil {
			return err
		}
	} else {
		// Read inputs from paths provided in arguments
		for _, path := range paths {
			switch info, err := os.Stat(path); {
			case err != nil:
				return err
			case info.IsDir():
				// Path is a directory- walk it
				err = filepath.Walk(
					path,
					func(subPath string, subInfo os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						if !subInfo.IsDir() && strings.HasSuffix(subPath, ".go") {
							// Shorten file and generate output
							contents, result, err := processFile(shortener, subPath)
							if err != nil {
								return err
							}
							err = handleOutput(subPath, contents, result)
							if err != nil {
								return err
							}
						}

						return nil
					},
				)
				if err != nil {
					return err
				}
			default:
				// Path is a file
				contents, result, err := processFile(shortener, path)
				if err != nil {
					return err
				}
				err = handleOutput(path, contents, result)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// processFile uses the provided Shortener instance to shorten the lines
// in a file. It returns the original contents (useful for debugging), the
// shortened version, and an error.
func processFile(shortener *Shortener, path string) ([]byte, []byte, error) {
	slog.Debug("Processing file", "path", path)

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	result, err := shortener.Shorten(contents)
	return contents, result, err
}

// handleOutput generates output according to the value of the tool's
// flags; depending on the latter, the output might be written over
// the source file, printed to stdout, etc.
func handleOutput(path string, contents []byte, result []byte) error {
	if contents == nil {
		return nil
	} else {
		if path == "" {
			return errors.New("No path to write out to")
		}

		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		if bytes.Equal(contents, result) {
			slog.Debug("Contents unchanged, skipping write")
			return nil
		}

		slog.Debug("Contents changed, writing output to path", "path", path)
		return os.WriteFile(path, result, info.Mode())
	}
}
