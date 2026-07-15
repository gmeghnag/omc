package stern

import (
	"bytes"
	"fmt"
	"io"
)

// prefixWriter prepends a fixed prefix to every line written to it.
// Writes are not guaranteed to be line-aligned by callers (e.g.
// logs.LogReader.Read may io.Copy raw file contents), so partial lines are
// buffered until a newline is seen. Call Flush after the last Write to emit
// any trailing partial line.
type prefixWriter struct {
	w      io.Writer
	prefix string
	buf    []byte
}

func newPrefixWriter(w io.Writer, podName, containerName string) *prefixWriter {
	return &prefixWriter{
		w:      w,
		prefix: fmt.Sprintf("%s %s", podName, containerName),
	}
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	for {
		idx := bytes.IndexByte(p.buf, '\n')
		if idx < 0 {
			break
		}
		line := p.buf[:idx]
		if _, err := fmt.Fprintf(p.w, "%s %s\n", p.prefix, line); err != nil {
			return len(b), err
		}
		p.buf = p.buf[idx+1:]
	}
	return len(b), nil
}

// Flush writes out any buffered partial line that was never terminated by a
// newline.
func (p *prefixWriter) Flush() error {
	if len(p.buf) == 0 {
		return nil
	}
	_, err := fmt.Fprintf(p.w, "%s %s\n", p.prefix, p.buf)
	p.buf = nil
	return err
}
