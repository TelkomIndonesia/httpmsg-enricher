package main

import (
	"fmt"
	"io"
)

type closableWriter interface {
	io.Writer
	io.Closer
}

type wcloserNoop struct {
	io.Writer
}

func (wcloserNoop) Close() (err error) { return }

type wNoop struct{}

func (wNoop) Write(p []byte) (n int, err error) { return len(p), nil }

func (wNoop) Close() (err error) { return }

var wnoop closableWriter = wNoop{}

func MultiCopy(r io.Reader, writers ...closableWriter) (err error) {
	iow := []io.Writer{}
	for _, w := range writers {
		iow = append(iow, w)
	}

	mw := io.MultiWriter(iow...)
	if _, err = io.Copy(mw, r); err != nil {
		return fmt.Errorf("error copying: %w", err)
	}

	defer func() {
		for _, w := range writers {
			errw := w.Close()
			if err == nil {
				err = errw
			}
		}
	}()

	return
}
