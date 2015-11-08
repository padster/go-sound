package fakeflac

import (
	"errors"
)

type Decoder struct {
	Rate     int
}

type Encoder struct {
	Depth    int
	Rate     int
}

type Frame struct {
	Channels int
	Depth    int
	Rate     int
	Buffer   []int32
}

func NewDecoder(name string) (d *Decoder, err error) {
	return nil, errors.New("Can't use flac decoder with fake flac")
}

func (d *Decoder) ReadFrame() (f *Frame, err error) {
	return nil, errors.New("Can't use flac decoder with fake flac")
}

func (d *Decoder) Close() {
}

func NewEncoder(name string, channels int, depth int, rate int) (e *Encoder, err error) {
	return nil, errors.New("Can't use flac encoder with fake flac")
}

func (e *Encoder) WriteFrame(f Frame) (err error) {
	return errors.New("Can't use flac encoder with fake flac")
}

func (e *Encoder) Close() {
}
