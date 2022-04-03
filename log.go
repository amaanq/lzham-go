package lzham

import (
	"os"

	"github.com/withmandala/go-log"
)

var (
	logger = log.New(os.Stdout).WithColor().WithDebug().WithTimestamp()
)
