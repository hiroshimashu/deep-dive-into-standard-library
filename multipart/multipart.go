package multipart

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/textproto"
)

var emptyParams = make(map[string]string)
// TODO 
// goのmakeの仕様に目を通す
// mapもsyntaxと使い方を復習する

const peekBufferSize = 4096

type Part struct {
	Header textproto.MIMEHeader

	mr *Reader

	disposition string 
	dispositionParams map[string]string

	r io.Reader

	n int 
	total int64
	err error
	readErr error
}

func (p *Part) Close() error {
	io.Copy(io.Discard, p)
}

func (p *Part) FileName() string {
	if p.dispositionParams == nil {
		p.parseContentDisposition()
	}
	return p.dispositionParams["filename"]
}

func (p *Part) parseContentDisposition() {
	v := p.Header.Get("Content-Disposition")
	var err error 
	p.disposition, p.dispositionParams, err = mime.ParseMediaType(v)
	if err != nil {
		p.dispositionParams = emptyParams
	}
}

func (p *Part) FormName() string {
	if p.dispositionParams == nil {
		p.parseContentDisposition()
	}

	if p.disposition != "form-data" {
		return ""
	}

	return p.dispositionParams["name"]
}

func (p *Part) Read(d []byte) (n int, err error) {
	return p.r.Read(d)
}

type Reader struct {
	bufReader *bufio.Reader

	currentPart *Part 
	partsRead int 

	nl []byte 
	nlDashBoundary []byte
	dashBoundaryDash []byte
	dashBoundary []byte
}

func (r *Reader) NextPart() (*Part, error) {
	return r.nextPart(false)
}

func (r *Reader) nextPart(rawPart bool) (*Part, error) {
	r.currentPart != nil {
		r.currentPart.Close()
	}
	if string(r.dashBoundary) == "--" {
		return nil, fmt.Errorf("multipart: boundary is empty")
	}
	expectNewPart := false

	for {
		line, err := r.bufReader.ReadSlice('\n')

		if err == io.EOF && r.isFinaleBoundary(line) {
			return nil, io.EOF
		}
		if err != nil {
			return nil, fmt.Errorf("mutipart: NextPart: %v", err)
		}

		if r.isBoundaryDelimiterLine(line) {
			r.partsRead++
			bp, err := newPart(r, rawPart)
			if err != nil {
				return nil, err
			}
			r.currentPart = bp
			return bp, nil
		}

		if r.partsRead == 0 {
			continue
		}

		return nil, fmt.Errorf("multipart: unexpected line in Next(): %q", line)
	}
}

func (mr *Reader) isBoundaryDelimiterLine(line []byte) (ret bool) {
	if !bytes.HasPrefix(line, mr.dashBoundary) {
		return false
	}
	rest := line[len(mr.dashBoundary):]
	rest = skipLWSPChar(rest)

	if mr.partsRead == 0 && len(rest) == 1 && rest[0] == '\n' {
		mr.nl = mr.nl[1:]
		mr.nlDashBoundary = mr.nlDashBoundary[1:]
	}
	return bytes.Equal(rest, mr.nl)
}

func skipLWSPChar(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

func NewReader(r io.Reader, boundary string) *Reader {
	b := []byte("\r\n--" + boundary + "--")
	return &Reader{
		bufReader: bufio.NewReaderSize(&stickyErrorReader{r: r}, peekBufferSize),
		nl: b[:2],
		nlDashBoundary: b[:len(b)-2],
		dashBoundaryDash: b[2:],
		dashBoundary: b[2: len(b)-2 ],
	}
}

type stickyErrorReader struct {
	r io.Reader
	err error 
}

func (r *stickyErrorReader) Read(p []byte) (n int, _ error) {
	if r.err != nil {
		return 0, r.err
	}
	n, r.err = r.r.Read(p)
	return n, r.err
}