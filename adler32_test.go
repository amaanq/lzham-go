package lzham

import (
	"bytes"
	"testing"
)

func Test_lzham_adler32(t *testing.T) {
	type args struct {
		buf *bytes.Buffer
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{name: "Hello, world!", args: args{buf: bytes.NewBufferString("Hello, world!")}, want: 0x205e048a},
		{name: "Goodbye", args: args{buf: bytes.NewBufferString("Goodbye")}, want: 0xaca02ca},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lzham_adler32(tt.args.buf); got != tt.want {
				t.Errorf("lzham_adler32() = %v, want %v", got, tt.want)
			}
		})
	}
}
