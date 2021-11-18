package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func readBool(r *bufio.Reader, placeholder string) (val bool) {
	for {
		fmt.Fprintf(os.Stdout, `%s : `, placeholder)
		b, _, e := r.ReadLine()
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			continue
		}
		str := strings.ToLower(strings.TrimSpace(string(b)))
		if str == "true" || str == "t" ||
			str == "yes" || str == "y" ||
			str == "1" {
			val = true
			break
		} else if str == "false" || str == "f" ||
			str == "no" || str == "n" ||
			str == "0" {
			val = false
			break
		}
	}
	return
}
