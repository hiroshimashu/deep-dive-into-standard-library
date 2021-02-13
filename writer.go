package multipart

import (
	"errors"
	"fmt"
	"io"
)

type Writer struct {
	w io.Writer
	boundary string 
	lastpart *part 
}

func NewWriter(w io.Writer) *Writer {
	return &Writer {
		w: w,
		boundary: randomBoundary(),		
	}
}

func (w *Writer) Boundary() string {
	return w.boundary
}

func (w *Writer) SetBoundary(boundary string) error {
	if w.lastpart != nil {
		return errors.New("mime: SetBoundary caleed after write")
	}

	if len(boundary) < 1 || len(boundary) > 70 {
		return errors.New("mime: invalid boundary length")
	}

	end := len(boundary) - 1

	for i, b := range boundary {
		if 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || '0' <= b && b <= '9' {
			continue 
		}
		switch b {
		case '\'', '(', ')', '+', '-', ',', '.', '/', ':', '=', '?':
			continue
		case ' ':
			if i != end {
				continue
			}
		}
		return errors.New("mime: invalid boundary character")
	}
	w.boundary = boundary
	return nil
}

func (w *Writer) Close() error {
	if w.lastpart != nil {
		if err := w.lastpart.close(); err != nil {
			return err
		}
		w.lastpart = nil
	}
	_, err := fmt.Fprintf(w.w, "\r\n--%s--\r\n", w.boundary)
	return err
}

