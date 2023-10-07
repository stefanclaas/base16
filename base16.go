package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
)

// isInputRedirected checks if input is being redirected.
func isInputRedirected() bool {
        fileInfo, _ := os.Stdin.Stat()
        return (fileInfo.Mode() & os.ModeCharDevice) == 0
}

func encode(input io.Reader, output io.Writer, lineLength int) error {
	buf := make([]byte, 1024)
	encoder := hex.NewEncoder(output)

	totalCount := 0
	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		for i := 0; i < n; i++ {
			if totalCount > 0 && lineLength > 0 && totalCount%lineLength == 0 {
				if _, err := output.Write([]byte("\n")); err != nil {
					return err
				}
			}

			if _, err := encoder.Write(buf[i : i+1]); err != nil {
				return err
			}

			totalCount += 2 // Each byte is encoded into two characters
		}
	}

	return nil
}

func removeNewlines(r io.Reader) io.Reader {
	return &newlineRemover{r}
}

type newlineRemover struct {
	r io.Reader
}

func (n *newlineRemover) Read(p []byte) (int, error) {
	count, err := n.r.Read(p)
	j := 0
	for i := 0; i < count; i++ {
		if p[i] != '\n' {
			p[j] = p[i]
			j++
		}
	}
	return j, err
}

func decode(input io.Reader, output io.Writer) error {
	buf := make([]byte, 1024)
	decoder := hex.NewDecoder(removeNewlines(input))

	for {
		n, err := decoder.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if _, err := output.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func main() {
    // Check if there are no command-line arguments (excluding the program name).
    if len(os.Args) == 1 && !isInputRedirected() {
        fmt.Println("No arguments provided. Exiting...")
        os.Exit(1) // Exit with a non-zero status code to indicate an error.
    }

    decodeFlag := flag.Bool("d", false, "Decode input (base16 to binary)")
    lineLengthFlag := flag.Int("l", 64, "Line length for encoded output (0 for no line breaks)")
    flag.Parse()

    if *decodeFlag {
        err := decode(os.Stdin, os.Stdout)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error decoding: %v\n", err)
            os.Exit(1)
        }
    } else {
        lineLength := *lineLengthFlag
        if lineLength < 0 {
            fmt.Fprintf(os.Stderr, "Line length must be a non-negative integer\n")
            os.Exit(1)
        }
        err := encode(os.Stdin, os.Stdout, lineLength)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
            os.Exit(1)
        }
    }
    fmt.Println()
}

