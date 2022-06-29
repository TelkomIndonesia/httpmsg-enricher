package main

import "bytes"

type truncatedBuffer struct {
	buff  bytes.Buffer
	limit int64
	n     int64
}

func newTruncatedBuffer(limit int64) *truncatedBuffer {
	return &truncatedBuffer{limit: limit}
}

func (lb *truncatedBuffer) Read(p []byte) (n int, err error) {
	return lb.buff.Read(p)
}

func (lb *truncatedBuffer) Write(p []byte) (n int, err error) {
	if lb.limit <= 0 {
		lb.n++
		return len(p), nil
	}

	l := len(p)
	if lb.limit < int64(l) {
		l = int(lb.limit)
	}
	n, err = lb.buff.Write(p[:l])
	if err != nil {
		return n, err
	}
	lb.limit -= int64(n)
	lb.n += int64(len(p) - (l - n))

	return len(p) - (l - n), nil
}

func (lb *truncatedBuffer) String() string {
	return lb.buff.String()
}

func (lb *truncatedBuffer) Close() error {
	return nil
}

func (lb *truncatedBuffer) Len() int64 {
	return lb.n
}
