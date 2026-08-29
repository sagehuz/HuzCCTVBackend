package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func runLogs(args []string) int {
	opts, err := parseLogOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 2
	}
	path := logFilePath()
	if _, err := os.Stat(path); err != nil {
		fmt.Fprintf(os.Stderr, "No log file yet (%s). Run 'huzbackend start' first.\n", path)
		return 1
	}
	if !opts.follow {
		lines, err := tailLines(path, opts.tail)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read the log: %v\n", err)
			return 1
		}
		fmt.Println(strings.Join(lines, "\n"))
		return 0
	}
	fmt.Printf("Following %s (Ctrl+C to stop)...\n", path)
	if err := followLog(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error while following the log: %v\n", err)
		return 1
	}
	return 0
}

type logOptions struct {
	tail   int
	follow bool
}

func parseLogOptions(args []string) (logOptions, error) {
	opts := logOptions{tail: 50}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-f" || a == "--follow":
			opts.follow = true
		case a == "-n" || a == "--tail":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing value after %s", a)
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n <= 0 {
				return opts, fmt.Errorf("invalid value %q for %s", args[i+1], a)
			}
			opts.tail = n
			i++
		case strings.HasPrefix(a, "--tail="):
			n, err := strconv.Atoi(strings.TrimPrefix(a, "--tail="))
			if err != nil || n <= 0 {
				return opts, fmt.Errorf("invalid value %q for --tail", a)
			}
			opts.tail = n
		default:
			return opts, fmt.Errorf("unknown option: %s", a)
		}
	}
	return opts, nil
}

// tailLines reads the last n lines of a log file.
func tailLines(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 {
		return nil, nil
	}

	const maxChunk = 256 * 1024
	pos := fi.Size()
	readLen := int64(maxChunk)
	if pos < readLen {
		readLen = pos
	}
	start := pos - readLen
	buf := make([]byte, readLen)
	if _, err := f.ReadAt(buf, start); err != nil {
		return nil, err
	}

	content := bytes.TrimRight(buf, "\n")
	parts := bytes.Split(content, []byte("\n"))
	if len(parts) > n {
		parts = parts[len(parts)-n:]
	}
	lines := make([]string, len(parts))
	for i, p := range parts {
		lines[i] = string(p)
	}
	return lines, nil
}

// followLog continuously prints new lines appended to the log file (tail -f style).
func followLog(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	offset, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	for {
		buf := make([]byte, 16*1024)
		n, err := f.ReadAt(buf, offset)
		if n > 0 {
			_, _ = os.Stdout.Write(buf[:n])
			offset += int64(n)
		}
		if err != nil && err != io.EOF {
			return err
		}
		time.Sleep(400 * time.Millisecond)
	}
}
