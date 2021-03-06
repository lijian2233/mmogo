package buffer

import (
	"errors"
	"mmogo/lib/binaryop"
	"mmogo/lib/locker"
	"unsafe"
)

// Copyright 2019 smallnest. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

var (
	ErrTooManyDataToWrite = errors.New("too many data to write")
	ErrIsFull             = errors.New("ringbuffer is full")
	ErrIsEmpty            = errors.New("ringbuffer is empty")
	ErrAccuqireLock       = errors.New("no lock to accquire")
)

// RingBuffer is a circular buffer that implement io.ReaderWriter interface.
type RingBuffer struct {
	buf    []byte
	size   int
	r      int // next position to read
	w      int // next position to write
	isFull bool
	lock   locker.EmptyLock
}

// New returns a new RingBuffer whose buffer has the given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buf:  make([]byte, size),
		size: size,
	}
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered. Even if Read returns n < len(p), it may use all of p as scratch space during the call. If some data is available but not len(p) bytes, Read conventionally returns what is available instead of waiting for more.
// When Read encounters an error or end-of-file condition after successfully reading n > 0 bytes, it returns the number of bytes read. It may return the (non-nil) error from the same call or return the error (and n == 0) from a subsequent call.
// Callers should always process the n > 0 bytes returned before considering the error err. Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
func (r *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	defer r.lock.Unlock()
	r.lock.Lock()
	n, err = r.read(p)
	return n, err
}

func (r *RingBuffer) read(p []byte) (n int, err error) {
	if r.w == r.r && !r.isFull {
		return 0, ErrIsEmpty
	}

	if r.w > r.r {
		n = r.w - r.r
		if n > len(p) {
			n = len(p)
		}
		copy(p, r.buf[r.r:r.r+n])
		r.r = (r.r + n) % r.size
		return
	}

	n = r.size - r.r + r.w
	if n > len(p) {
		n = len(p)
	}

	if r.r+n <= r.size {
		copy(p, r.buf[r.r:r.r+n])
	} else {
		c1 := r.size - r.r
		copy(p, r.buf[r.r:r.size])
		c2 := n - c1
		copy(p[c1:], r.buf[0:c2])
	}
	r.r = (r.r + n) % r.size

	r.isFull = false

	return n, err
}

// ReadByte reads and returns the next byte from the input or ErrIsEmpty.
func (r *RingBuffer) ReadByte() (b byte, err error) {
	defer r.lock.Unlock()
	r.lock.Lock()
	if r.w == r.r && !r.isFull {
		return 0, ErrIsEmpty
	}
	b = r.buf[r.r]
	r.r++
	if r.r == r.size {
		r.r = 0
	}

	r.isFull = false
	return b, err
}

// Write writes len(p) bytes from p to the underlying buf.
// It returns the number of bytes written from p (0 <= n <= len(p)) and any error encountered that caused the write to stop early.
// Write returns a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	defer r.lock.Unlock()
	r.lock.Lock()
	n, err = r.write(p)

	return n, err
}

func (r *RingBuffer) write(p []byte) (n int, err error) {
	if r.isFull {
		return 0, ErrIsFull
	}

	var avail int
	if r.w >= r.r {
		avail = r.size - r.w + r.r
	} else {
		avail = r.r - r.w
	}

	if len(p) > avail {
		err = ErrTooManyDataToWrite
		p = p[:avail]
	}
	n = len(p)

	if r.w >= r.r {
		c1 := r.size - r.w
		if c1 >= n {
			copy(r.buf[r.w:], p)
			r.w += n
		} else {
			copy(r.buf[r.w:], p[:c1])
			c2 := n - c1
			copy(r.buf[0:], p[c1:])
			r.w = c2
		}
	} else {
		copy(r.buf[r.w:], p)
		r.w += n
	}

	if r.w == r.size {
		r.w = 0
	}
	if r.w == r.r {
		r.isFull = true
	}

	return n, err
}

// WriteByte writes one byte into buffer, and returns ErrIsFull if buffer is full.
func (r *RingBuffer) WriteByte(c byte) error {
	defer r.lock.Unlock()
	r.lock.Lock()
	err := r.writeByte(c)
	return err
}

func (r *RingBuffer) IncWriteIndex(inc int) {
	if inc == 0 {
		return
	}
	r.w = (r.w + inc) % r.size
	if r.r == r.w {
		r.isFull = true
	}
}


func (r *RingBuffer) writeByte(c byte) error {
	if r.w == r.r && r.isFull {
		return ErrIsFull
	}
	r.buf[r.w] = c
	r.w++

	if r.w == r.size {
		r.w = 0
	}
	if r.w == r.r {
		r.isFull = true
	}

	return nil
}

// Length return the length of available read bytes.
func (r *RingBuffer) Length() int {
	if r.r == r.w {
		if r.isFull {
			return r.size
		}
		return 0
	}
	if r.w > r.r {
		return r.w - r.r
	}
	return r.size - r.r + r.w
}

// Capacity returns the size of the underlying buffer.
func (r *RingBuffer) Capacity() int {
	return r.size
}

// Free returns the length of available bytes to write.
func (r *RingBuffer) Free() int {
	defer r.lock.Unlock()
	r.lock.Lock()

	if r.r == r.w {
		if r.isFull {
			return 0
		}
		return r.size
	}
	if r.w < r.r {
		return r.r - r.w
	}

	return r.size - r.w + r.r
}

// WriteString writes the contents of the string s to buffer, which accepts a slice of bytes.
func (r *RingBuffer) WriteString(s string) (n int, err error) {
	defer r.lock.Unlock()
	r.lock.Lock()
	buf := binaryop.String2Byte(s)
	return r.Write(buf)
}

func (r *RingBuffer) Erase(n int) {
	r.lock.Lock()
	if r.Length() <= n {
		r.r = r.w
		r.isFull = false
	}else{
		r.r = (r.r + n) % r.size
		r.isFull = false
	}
	r.lock.Unlock()

	if r.Length() > r.size/3 {
		r.Adjust()
	}
}

func (r *RingBuffer) UnsafeReadBytes() []byte {
	defer r.lock.Unlock()
	r.lock.Lock()

	if r.IsEmpty() {
		return nil
	}

	ptr := uintptr(unsafe.Pointer(&r.buf[0]))
	ptr += uintptr(r.r)

	if r.w > r.r {
		h := [3]uintptr{ptr, uintptr(r.w - r.r), uintptr(r.w - r.r)}
		buf := *(*[]byte)(unsafe.Pointer(&h))
		return buf
	} else {
		h := [3]uintptr{ptr, uintptr(r.size - r.r), uintptr(r.size - r.r)}
		buf := *(*[]byte)(unsafe.Pointer(&h))
		return buf
	}
}

func (r *RingBuffer) UnsafeWriteSpace() []byte {
	r.lock.Lock()
	if r.w == r.r  {
		if r.isFull {
			r.lock.Unlock()
			return nil
		}
		r.r,r.w = 0, 0
		r.isFull = false
	}
	defer r.lock.Unlock()


	ptr := uintptr(unsafe.Pointer(&r.buf[0]))
	ptr += uintptr(r.w)
	if (r.w >= r.r) {
		h := [3]uintptr{ptr, uintptr(r.size - r.w), uintptr(r.size - r.w)}
		buf := *(*[]byte)(unsafe.Pointer(&h))
		return buf
	} else {
		h := [3]uintptr{ptr, uintptr(r.r - r.w), uintptr(r.r - r.w)}
		buf := *(*[]byte)(unsafe.Pointer(&h))
		return buf
	}
}

/*
*??????????????????, r?????????????????????, ????????????????????????
 */
func (r *RingBuffer) Adjust() {
	defer r.lock.Unlock()
	r.lock.Lock()

	if r.isFull {
		return
	}

	if r.r == r.w {
		r.r, r.w = 0, 0
	} else if r.r < r.w {
		copy(r.buf[0:r.w-r.r], r.buf[r.r:r.w])
		len := r.w - r.r
		r.r = 0
		r.w = len
	}else{
		if r.r - r.w >= r.size - r.r {
			//??????copy,??????????????????
			l := r.size - r.r
			for i:=0; i< r.w; i++{
				r.buf[r.w + l - 1 -i] = r.buf[r.w - 1 - i]
			}
			copy(r.buf[0:], r.buf[r.r:r.size])
			r.w = r.w + l
			r.r = 0
		}
	}
}

// Bytes returns all available read bytes. It does not move the read pointer and only copy the available data.
func (r *RingBuffer) Bytes() []byte {
	defer r.lock.Unlock()
	r.lock.Lock()

	if r.w == r.r {
		if r.isFull {
			buf := make([]byte, r.size)
			copy(buf, r.buf[r.r:])
			copy(buf[r.size-r.r:], r.buf[:r.w])
			return buf
		}
		return nil
	}

	if r.w > r.r {
		buf := make([]byte, r.w-r.r)
		copy(buf, r.buf[r.r:r.w])
		return buf
	}

	n := r.size - r.r + r.w
	buf := make([]byte, n)

	if r.r+n < r.size {
		copy(buf, r.buf[r.r:r.r+n])
	} else {
		c1 := r.size - r.r
		copy(buf, r.buf[r.r:r.size])
		c2 := n - c1
		copy(buf[c1:], r.buf[0:c2])
	}
	return buf
}

// IsFull returns this ringbuffer is full.
func (r *RingBuffer) IsFull() bool {
	return r.isFull
}

// IsEmpty returns this ringbuffer is empty.
func (r *RingBuffer) IsEmpty() bool {
	return !r.isFull && r.w == r.r
}

// Reset the read pointer and writer pointer to zero.
func (r *RingBuffer) Reset() {
	r.lock.Lock()
	r.r = 0
	r.w = 0
	r.isFull = false
	r.lock.Unlock()
}
