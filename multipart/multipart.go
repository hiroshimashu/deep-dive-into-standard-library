package multipart

import (
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