package metadata

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	net_url "net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/zuiwuchang/mget/cmd/internal/log"
	"github.com/zuiwuchang/mget/utils"
	"github.com/zuiwuchang/mget/version"
	"golang.org/x/net/proxy"
)

var (
	UserAgent  = `mget/` + version.Version + `; ` + version.Platform
	MaxWorkers = runtime.NumCPU() * 10
)

type Configure struct {
	URL          string
	Output       string
	Proxy        string
	UserAgent    string
	Head         bool
	Header       http.Header
	Cookie       []string
	offsetCookie int
	Insecure     bool
	Worker       int
	Block        utils.Size
	m            sync.Mutex
	client       *http.Client
	ASCII        bool
}

func NewConfigure(url, output, proxy string,
	agent string, head bool, headers, cookies []string, insecure bool,
	worker int, blockStr string,
) (conf *Configure, e error) {
	u, e := net_url.ParseRequestURI(url)
	if e != nil {
		return
	}
	if output == `` {
		output = path.Base(path.Clean(u.Path))
	}
	if output == `` || output == `/` || output == `.` || output == `..` {
		e = errors.New(`not supported output`)
		return
	}
	if filepath.IsAbs(output) {
		output = filepath.Clean(output)
	} else {
		output, e = filepath.Abs(output)
		if e != nil {
			return
		}
	}
	if proxy != `` {
		var p *net_url.URL
		p, e = net_url.ParseRequestURI(proxy)
		if e != nil {
			return
		} else if p.Scheme != `socks5` && p.Scheme != `http` && p.Scheme != `https` {
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
	} else if worker > MaxWorkers {
		worker = MaxWorkers
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
func (c *Configure) String() string {
	var w bytes.Buffer
	c.WriteFormat(&w, ``)
	return w.String()
}
func (c *Configure) WriteFormat(w io.Writer, prefix string) (n int, e error) {
	num, e := fmt.Fprintln(w, prefix+`      URL:`, c.URL)
	if e != nil {
		return
	}
	n += num
	num, e = fmt.Fprintln(w, prefix+`   Output:`, c.Output)
	if e != nil {
		return
	}
	n += num

	num, e = fmt.Fprint(w, prefix+`     Head: `, c.Head)
	if e != nil {
		return
	}
	n += num
	num, e = fmt.Fprint(w, `  Insecure: `, c.Insecure)
	if e != nil {
		return
	}
	n += num
	num, e = fmt.Fprint(w, `  Worker: `, c.Worker)
	if e != nil {
		return
	}
	n += num
	num, e = fmt.Fprintln(w, ` Block:`, c.Block)
	if e != nil {
		return
	}
	n += num

	num, e = fmt.Fprintln(w, prefix+`    Proxy:`, c.Proxy)
	if e != nil {
		return
	}
	n += num
	num, e = fmt.Fprintln(w, prefix+`UserAgent:`, c.UserAgent)
	if e != nil {
		return
	}
	n += num

	if len(c.Header) != 0 {
		num, e = fmt.Fprintln(w, prefix+`   Header: [`)
		if e != nil {
			return
		}
		n += num
		for k, v := range c.Header {
			num, e = fmt.Fprintf(w, "   "+prefix+prefix+`%s = [`, k)
			if e != nil {
				return
			}
			n += num
			for i, s := range v {
				if i == 0 {
					num, e = fmt.Fprintf(w, "%q", s)
					if e != nil {
						return
					}
					n += num
				} else {
					num, e = fmt.Fprintf(w, ",%q", s)
					if e != nil {
						return
					}
					n += num
				}
			}
			num, e = fmt.Fprintln(w, `]`)
			if e != nil {
				return
			}
			n += num
		}
		num, e = fmt.Fprintln(w, prefix+`]`)
		if e != nil {
			return
		}
		n += num
	}
	if len(c.Cookie) != 0 {
		num, e = fmt.Fprintln(w, prefix+`   Cookie: [`)
		if e != nil {
			return
		}
		n += num
		for _, v := range c.Cookie {
			num, e = fmt.Fprintln(w, `   `+prefix+prefix, v)
			if e != nil {
				return
			}
			n += num
		}
		num, e = fmt.Fprintln(w, prefix+`]`)
		if e != nil {
			return
		}
		n += num
	}
	return
}
func (c *Configure) Println() {
	fmt.Println(`Configure {`)
	c.WriteFormat(os.Stdout, `   `)
	fmt.Println(`}`)
}
func (c *Configure) Do(req *http.Request) (*http.Response, error) {
	client, e := c.Client()
	if e != nil {
		return nil, e
	}
	return client.Do(req)
}
func (c *Configure) GetMetadata(ctx context.Context) (modified string, size int64, e error) {
	var req *http.Request
	if c.Head {
		req, e = c.NewRequestWithContext(ctx, http.MethodHead, c.URL, nil)
	} else {
		req, e = c.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	}
	if e != nil {
		return
	}
	log.Info(req.Method, ` `, req.URL)
	resp, e := c.Do(req)
	if e != nil {
		return
	}
	defer resp.Body.Close()
	if resp.Header.Get(`Accept-Ranges`) != `bytes` {
		e = errors.New(`server not supported: Accept-Ranges`)
		return
	}
	modified = resp.Header.Get(`Last-Modified`)
	size, e = strconv.ParseInt(resp.Header.Get(`Content-Length`), 10, 64)
	if e != nil {
		return
	}
	return
}
func (c *Configure) NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (req *http.Request, e error) {
	req, e = http.NewRequestWithContext(ctx, method, url, body)
	if e != nil {
		return
	}
	header := req.Header
	for m, k := range c.Header {
		header[m] = k
	}
	header.Set(`User-Agent`, c.UserAgent)
	cookie := c.cookie()
	if cookie != `` {
		header.Set(`Cookie`, cookie)
	}
	return
}
func (c *Configure) cookie() string {
	if len(c.Cookie) == 0 {
		return ``
	}
	c.m.Lock()
	v := c.Cookie[c.offsetCookie]
	c.offsetCookie++
	if c.offsetCookie >= len(c.Cookie) {
		c.offsetCookie = 0
	}
	c.m.Unlock()
	return v
}
func (c *Configure) Exists() (exists bool, e error) {
	info, e := os.Stat(c.Output)
	if e != nil {
		if os.IsNotExist(e) {
			e = nil
		}
		return
	}
	exists = true
	if info.IsDir() {
		e = fmt.Errorf(`dir already exists: %s`, c.Output)
		return
	}
	return
}
func (c *Configure) Client() (client *http.Client, e error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.client != nil {
		client = c.client
		return
	}
	if c.Insecure {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	} else if c.Proxy != `` {
		var transport http.Transport
		var p *net_url.URL
		p, e = net_url.ParseRequestURI(c.Proxy)
		if e != nil {
			return
		} else if p.Scheme == `http` || p.Scheme == `https` {
			transport.Proxy = http.ProxyURL(p)
		} else if p.Scheme == `socks5` {
			var dialer proxy.Dialer
			dialer, e = proxy.SOCKS5(`tcp`, p.Host, nil, proxy.Direct)
			if e != nil {
				return
			}
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		} else {
			e = fmt.Errorf(`not supported proxy scheme: %v`, c.URL)
			return
		}
		if c.Insecure {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		client = &http.Client{
			Transport: &transport,
		}
	} else {
		client = http.DefaultClient
	}
	c.client = client
	return
}
