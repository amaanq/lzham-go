package lzham

import (
	"bytes"
	"hash/crc32"
)


func lzham_crc32(buf *bytes.Buffer) uint32 {
	return crc32.ChecksumIEEE(buf.Bytes())
}
