package multipart

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"sort"
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

func (w *Writer) CreateFormField(fieldname string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", 
		fmt.Sprintf(`form-data; name=%s`, escapeQuotes(fieldname)))
	return w.CreatePart(h)
}

func (w *Writer) CreatePart(header textproto.MIMEHeader) (io.Writer, error) {
	if w.lastpart != nil {
		if err := w.lastpart.close(); err != nil {
			return nil, err
		}
	}
	var b bytes.Buffer 
	if w.lastpart != nil {
		fmt.Fprintf(&b, "\r\n--%s\r\n", w.boundary)
	} else {
		fmt.Fprintf(&b, "--%s\r\n", w.boundary)
	}
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			fmt.Fprintf(&b, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprintf(&b, "\r\n")
	_, err := io.Copy(w.w, &b)
	if err != nil {
		return nil, err
	}
	p := &part {
		mw: w,
	}
	w.lastpart = p
	return p, nil 
}

