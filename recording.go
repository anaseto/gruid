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
	vd := &FrameDecoder{}
	var err error
	vd.gzr, err = gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("video decoding: gzip: %v", err)
	}
	vd.gbd = gob.NewDecoder(vd.gzr)
	return vd, nil
}

// Decode retrieves the next frame from the input stream. If the input is at
// EOF, it returns the error io.EOF.
func (vd *FrameDecoder) Decode() (Frame, error) {
	var frame Frame
	err := vd.gbd.Decode(&frame)
	return frame, err
}

type frameEncoder struct {
	gzw *gzip.Writer
	gbe *gob.Encoder
}

func newFrameEncoder(w io.Writer) *frameEncoder {
	ve := &frameEncoder{}
	ve.gzw = gzip.NewWriter(w)
	ve.gbe = gob.NewEncoder(ve.gzw)
	return ve
}

func (ve *frameEncoder) encode(fr Frame) error {
	err := ve.gbe.Encode(fr)
	if err != nil {
		return err
	}
	return nil
}
