package gruid

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
)

// FrameDecoder manages the decoding of the frame recording stream produced by
// the running of an application, in case a FrameWriter was provided. It can be
// used to replay an application session.
type FrameDecoder struct {
	gzr *gzip.Reader
	gbd *gob.Decoder
}

// NewFrameDecoder returns a FrameDecoder using a given reader as source for
// frames.
//
// It is your responsibility to call Close on the reader when done.
func NewFrameDecoder(r io.Reader) (*FrameDecoder, error) {
	fd := &FrameDecoder{}
	var err error
	fd.gzr, err = gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("video decoding: gzip: %v", err)
	}
	fd.gbd = gob.NewDecoder(fd.gzr)
	return fd, nil
}

// Decode retrieves the next frame from the input stream. If the input is at
// EOF, it returns the error io.EOF.
func (fd *FrameDecoder) Decode() (Frame, error) {
	var frame Frame
	err := fd.gbd.Decode(&frame)
	return frame, err
}

type frameEncoder struct {
	gzw *gzip.Writer
	gbe *gob.Encoder
}

func newFrameEncoder(w io.Writer) *frameEncoder {
	fe := &frameEncoder{}
	fe.gzw = gzip.NewWriter(w)
	fe.gbe = gob.NewEncoder(fe.gzw)
	return fe
}

func (fe *frameEncoder) encode(fr Frame) error {
	err := fe.gbe.Encode(fr)
	if err != nil {
		return err
	}
	return nil
}
