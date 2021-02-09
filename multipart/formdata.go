package multipart

import (
	"bytes"
	"errors"
	"io"
	"net/textproto"
	"os"
)

// Q. FileHeader内のtmpFileの扱い

var ErrMessageTooLarge = errors.New("multipart: message too large")


type FileHeader struct {
	Filename string 
	Header textproto.MIMEHeader
	Size int64 

	content []byte 
	tmpfile string 
}

// Open :Read binary data and return File
func (fh *FileHeader) Open() (File, error) {
	if b := fh.content; b != nil {
		r := io.NewSectionReader(bytes.NewReader(b), 0, int64(len(b)))
		return sectionReadCloser{ r }, nil
	}
}


type File interface {
	io.Reader 
	io.ReaderAt
	io.Seeker 
	io.Closer 
}


// Turn binary data to File
type sectionReadCloser struct {
	*io.SectionReader 
}

func  (rc sectionReadCloser) Close() error {
	return nil
}
// TODO 
// io packageもついでに覗いてみる

type Form struct {
	Value map[string][]string
	File map[string][]*FileHeader
}

func (f *Form) RemoveAll() error {
	var err error
	for _, fhs := range f.File {
		for _, fh := range fhs {
			if fh.tmpfile != "" {
				e := os.Remove(fh.tmpfile)
				if e != nil && err == nil {
					err = e 
				}
			}
		}
	}
	return err 
}