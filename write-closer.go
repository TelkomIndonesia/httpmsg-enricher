package main

import (
	"fmt"
	"io"

	"go.uber.org/multierr"
)

type nopCloser struct{ io.Writer }

func (nopCloser) Close() (err error) { return }

type nopWriteCloser struct{}

func (nopWriteCloser) Write(p []byte) (n int, err error) { return len(p), nil }

func (nopWriteCloser) Close() (err error) { return }

var nopwc io.WriteCloser = nopWriteCloser{}

func MultiCopy(r io.Reader, writers ...io.WriteCloser) (err error) {
	if len(writers) == 0 {
		return
	}

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
			if errw := w.Close(); errw != nil {
				err = multierr.Append(err, errw)
			}
		}
		if err != nil {
			err = fmt.Errorf("error closing writer(s) : %w", err)
		}
	}()

	return
}
