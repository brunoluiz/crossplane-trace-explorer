package xplane

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
)

// CLITraceQuerier defines a trace querier using the crossplane CLI
type CLITraceQuerier struct {
	app  string
	args []string
}

func NewCLITraceQuerier(cmd string, namespace string, name string) *CLITraceQuerier {
	s := strings.Split(cmd, " ")
	app := s[0]
	args := s[1:]
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	args = append(args, name)

	return &CLITraceQuerier{
		app:  app,
		args: args,
	}
}

func (q *CLITraceQuerier) GetTrace() (*Resource, error) {
	//nolint // trust the user input
	stdout, err := exec.Command(q.app, q.args...).Output()
	if err != nil {
		return nil, err
	}

	return Parse(bytes.NewReader(stdout))
}

func (q *CLITraceQuerier) MustGetTrace() *Resource {
	o, err := q.GetTrace()
	if err != nil {
		panic(err)
	}
	return o
}

// ReaderTraceQuerier defines a trace querier using piped files through stdin
type ReaderTraceQuerier struct {
	r io.Reader
}

func NewReaderTraceQuerier(r io.Reader) *ReaderTraceQuerier {
	return &ReaderTraceQuerier{r: r}
}

func (q *ReaderTraceQuerier) GetTrace() (*Resource, error) {
	return Parse(q.r)
}

func (q *ReaderTraceQuerier) MustGetTrace() *Resource {
	return MustParse(q.r)
}
