package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type Size int64

const (
	K = 1024
	M = K * 1024
	G = M * 1024
)

func ParseSize(str string) (result Size, e error) {
	var v, mul int64
	str = strings.ToLower(str)
	for str != `` {
		i := strings.IndexAny(str, "gmkb")
		if i < 0 {
			break
		}
		switch str[i] {
		case 'g':
			mul = G
		case 'm':
			mul = M
		case 'k':
			mul = K
		default:
			mul = 1
		}
		v, e = strconv.ParseInt(str[:i], 10, 64)
		if e != nil {
			return
		}
		result += Size(v * mul)
		str = str[i+1:]
	}
	if str != `` {
		v, e = strconv.ParseInt(str, 10, 64)
		if e != nil {
			return
		}
		result += Size(v)
	}
	return
}
func (s Size) String() string {
	var (
		v    = int64(s)
		strs []string
		tags = []struct {
			Tag  string
			Size int64
		}{
			{
				Tag:  `g`,
				Size: G,
			},
			{
				Tag:  `m`,
				Size: M,
			},
			{
				Tag:  `k`,
				Size: K,
			},
		}
	)
	for _, tag := range tags {
		if v >= tag.Size {
			strs = append(strs, fmt.Sprintf(`%v%s`, v/tag.Size, tag.Tag))
			v %= tag.Size
		}
	}
	if v > 0 {
		strs = append(strs, fmt.Sprintf(`%vb`, v))
	}
	if len(strs) == 0 {
		return `0b`
	}
	return strings.Join(strs, ``)
}
