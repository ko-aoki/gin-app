package param

import (
	"time"

	"cloud.google.com/go/civil"
)

type ReqSandbox struct {
	KeyCl int64 `json:"key" binding:"required"`
}

type ResSandbox struct {
	KeyCl       int64      `json:"key"`
	IntCl       int64      `json:"int"`
	StrCl       string     `json:"str"`
	BoolCl      bool       `json:"bool"`
	ByteCl      []byte     `json:"byteArray"`
	DateCl      civil.Date `json:"date"`
	TimeStampCl time.Time  `json:"time "`
	JsonCl      string     `json:"json"`
}
