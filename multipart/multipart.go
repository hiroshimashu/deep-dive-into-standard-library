package multipart

var emptyParams = make(map[string]string)
// TODO 
// goのmakeの仕様に目を通す
// mapもsyntaxと使い方を復習する

const peekBufferSize = 4096