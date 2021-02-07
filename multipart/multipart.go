package multipart

import (
	"io"
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