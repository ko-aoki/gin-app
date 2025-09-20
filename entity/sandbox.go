package entity

import "cloud.google.com/go/spanner"

type Sandbox struct {
	keyCl       int64
	IntCl       int64
	StrCl       string
	BoolCl      bool
	ByteCl      []byte
	DateCl      spanner.NullDate
	TimeStampCl spanner.NullTime
	JsonCl      spanner.NullJSON
}
