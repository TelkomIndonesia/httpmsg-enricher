package main

import "bytes"

type truncatedBuffer struct {
	buff   bytes.Buffer
	limit  int
	oriLen int
}

func newTruncatedBuffer(limit int) *truncatedBuffer {
	return &truncatedBuffer{limit: limit}
}

func (lb *truncatedBuffer) Read(p []byte) (n int, err error) {
	return lb.buff.Read(p)
}

func (lb *truncatedBuffer) Write(p []byte) (n int, err error) {
	curlen := lb.buff.Len()
	if curlen >= lb.limit {
		lb.oriLen += len(p)
		return len(p), nil
	}

	l := len(p)
	if left := lb.limit - curlen; left < l {
		l = left
	}
	n, err = lb.buff.Write(p[:l])
	if err != nil {
		return n, err
	}

	n = len(p) - (l - n)
	lb.oriLen += n
	return n, nil
}

func (lb *truncatedBuffer) String() string {
	return lb.buff.String()
}

func (lb *truncatedBuffer) Close() error {
	return nil
}

func (lb *truncatedBuffer) Len() int {
	return lb.oriLen
}
