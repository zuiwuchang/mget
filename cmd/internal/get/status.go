package get

import "strconv"

type Status int

const (
	StatusNil Status = iota
	StatusInit
	StatusDownload
	StatusError
	StatusSuccess
)

func (s Status) String() string {
	switch s {
	case StatusNil:
		return `Nil`
	case StatusInit:
		return `Init`
	case StatusDownload:
		return `Download`
	case StatusError:
		return `Error`
	case StatusSuccess:
		return `Success`
	}
	return `Unkonw<` + strconv.Itoa(int(s)) + `>`
}
