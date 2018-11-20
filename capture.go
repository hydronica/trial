package trial

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type capture struct {
	stdout   *os.File
	stderr   *os.File
	reader   *os.File
	writer   *os.File
	done     chan struct{}
	lines    []string
	logFlags int
}

// CaptureLog overrides the log output for reading
func CaptureLog() *capture {
	reader, writer, _ := os.Pipe()
	c := &capture{
		logFlags: log.Flags(),
		reader:   reader,
		writer:   writer,
		done:     make(chan struct{}),
	}
	log.SetOutput(writer)
	log.SetFlags(0)
	go c.read()
	return c
}

// CaptureStdErr redirects stderr for reading
// note this does not redirect log output
func CaptureStdErr() *capture {
	reader, writer, _ := os.Pipe()

	c := &capture{
		stderr: os.Stderr,
		reader: reader,
		writer: writer,
		done:   make(chan struct{}),
	}
	os.Stderr = writer
	go c.read()
	return c
}

// CaptureStdOut redirects stdout for reading
func CaptureStdOut() *capture {
	reader, writer, _ := os.Pipe()

	c := &capture{
		stdout: os.Stdout,
		reader: reader,
		writer: writer,
		done:   make(chan struct{}),
	}
	os.Stdout = writer
	go c.read()
	return c
}

func (c *capture) read() {
	scanner := bufio.NewScanner(bufio.NewReader(c.reader))
	for scanner.Scan() {
		c.lines = append(c.lines, scanner.Text())
	}
	close(c.done)
}

// ReadAll stops the capturing of data and returns the collected data
func (c *capture) ReadAll() string {
	c.reset()
	<-c.done
	return strings.Join(c.lines, "\n")
}

// ReadLines stops the capturing of data and returns the collected lines
func (c *capture) ReadLines() []string {
	c.reset()
	<-c.done
	return c.lines
}

func (c *capture) reset() {
	if c.writer != nil {
		c.writer.Close()
		c.writer = nil
	}
	if c.stderr != nil {
		os.Stderr = c.stderr
		c.stderr = nil
	}
	if c.logFlags != 0 {
		log.SetFlags(c.logFlags)
		log.SetOutput(os.Stderr)
	}
	if c.stdout != nil {
		os.Stdout = c.stdout
		c.stdout = nil
	}
}
