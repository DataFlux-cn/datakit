package dialtesting

import (
	"time"
)

const (
	StatusStop = "stop"

	ClassHTTP  = "HTTP"
	ClassTCP   = "TCP"
	ClassDNS   = "DNS"
	ClassOther = "OTHER"
)

type Task interface {
	ID() string
	Status() string
	Run() error
	Init() error
	CheckResult() []string
	Class() string
	Stop() error
	RegionName() string
	AccessKey() string
	Check() error
	UpdateTimeUs() int64

	Ticker() *time.Ticker
}
