package lzham

import (
	"bytes"
	"hash/adler32"
)


func lzham_adler32(buf *bytes.Buffer) uint32 {
	return adler32.Checksum(buf.Bytes())
}