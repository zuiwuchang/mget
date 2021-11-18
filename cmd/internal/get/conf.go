package get

import (
	"fmt"
	"net/http"
	"net/textproto"
	net_url "net/url"
	"strings"

	"github.com/zuiwuchang/mget/utils"
	"github.com/zuiwuchang/mget/version"
)

var UserAgent = `mget/` + version.Version + `; ` + version.Platform

type Configure struct {
	URL       string
	Output    string
	Proxy     string
	UserAgent string
	Head      bool
	Header    http.Header
	Cookie    []string
	Insecure  bool
	Worker    int
	Block     utils.Size
}

func NewConfigure(url, output, proxy string,
	agent string, head bool, headers, cookies []string, insecure bool,
	worker int, blockStr string,
) (conf *Configure, e error) {
	_, e = net_url.ParseRequestURI(url)
	if e != nil {
		return
	}
	if proxy != `` {
		var p *net_url.URL
		p, e = net_url.ParseRequestURI(proxy)
		if e != nil {
			return
		} else if p.Scheme != `socks5` && p.Scheme != `http` {
			e = fmt.Errorf(`proxy only supported socks5 or http, not supported %s`, p.Scheme)
			return
		}
	}

	h := make(http.Header)
	c := make([]string, 0, len(cookies))
	for _, v := range cookies {
		if v != `` {
			c = append(c, v)
		}
	}
	for _, str := range headers {
		strs := strings.SplitN(str, `: `, 2)
		key := textproto.CanonicalMIMEHeaderKey(strs[0])
		var val string
		if len(strs) == 2 {
			val = strs[1]
		}
		if key == `User-Agent` {
			if val != `` {
				agent = val
			}
		} else if key == `Cookie` {
			if val != `` {
				c = append(c, val)
			}
		} else {
			h.Add(key, val)
		}
	}

	if agent == `` {
		agent = UserAgent
	}
	if worker < 1 {
		e = fmt.Errorf(`worker must be greater than 0, not supported %d`, worker)
		return
	}
	block, e := utils.ParseSize(blockStr)
	if e != nil {
		return
	} else if block < 0 {
		e = fmt.Errorf(`block size must be greater than 0, not supported %s`, blockStr)
		return
	}

	conf = &Configure{
		URL:       url,
		Output:    output,
		Proxy:     proxy,
		UserAgent: agent,
		Head:      head,
		Header:    h,
		Cookie:    c,
		Insecure:  insecure,
		Worker:    worker,
		Block:     block,
	}
	return
}

func (c *Configure) Println() {
	fmt.Println(`Configure {`)
	fmt.Println(`	URL:`, c.URL)
	fmt.Println(`	Output:`, c.Output)
	fmt.Println(`	Proxy:`, c.Proxy)
	fmt.Println(`	UserAgent:`, c.UserAgent)
	fmt.Println(`	Head:`, c.Head)
	if len(c.Header) != 0 {
		fmt.Println(`	Header: [`)
		for k, v := range c.Header {
			fmt.Printf(`		%s = [`, k)
			for i, s := range v {
				if i == 0 {
					fmt.Printf("%q", s)
				} else {
					fmt.Printf(",%q", s)
				}
			}
			fmt.Println(`]`)
		}
		fmt.Println(`	]`)
	}
	if len(c.Cookie) != 0 {
		fmt.Println(`	Cookie: [`)
		for _, v := range c.Cookie {
			fmt.Println(`		`, v)
		}
		fmt.Println(`	]`)
	}
	fmt.Println(`	Insecure:`, c.Insecure)
	fmt.Println(`	Worker:`, c.Worker)
	fmt.Println(`	Block:`, c.Block)
	fmt.Println(`}`)
}
