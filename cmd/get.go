package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zuiwuchang/mget/cmd/internal/get"
	"github.com/zuiwuchang/mget/cmd/internal/metadata"
)

func init() {
	var (
		url           string
		output        string
		proxy         string
		agent         string
		head          bool
		headers       []string
		cookies       []string
		blockStr      string
		worker        int
		yes, insecure bool
		ascii         bool
	)
	cmd := &cobra.Command{
		Use:   `get`,
		Short: `http get download file`,
		Example: `mget get -u http://127.0.0.1/tools/source.exe
mget get -u http://127.0.0.1/tools/source.exe -o a.exe`,
		Run: func(cmd *cobra.Command, args []string) {
			if proxy == `` {
				env := []string{
					`socket_proxy`,
					`SOCKET_proxy`,
					`http_proxy`,
					`HTTP_PROXY`,
				}
				for _, k := range env {
					v := os.Getenv(k)
					if v == `` {
						continue
					}
					k = strings.ToLower(k)
					if strings.HasPrefix(k, `http`) {
						if !strings.HasPrefix(v, `http://`) && !strings.HasPrefix(v, `https://`) {
							continue
						}
					} else {
						if !strings.HasPrefix(v, `socks5://`) {
							continue
						}
					}
					proxy = v
					break
				}
			}
			conf, e := metadata.NewConfigure(url, output, proxy,
				agent, head, headers, cookies, insecure,
				worker, blockStr,
			)
			if e != nil {
				log.Fatalln(e)
			}
			conf.ASCII = ascii
			conf.Println()
			if !yes {
				val := readBool(bufio.NewReader(os.Stdin), `Are you sure you want to start downloading <y/n>`)
				if !val {
					return
				}
			}
			exists, e := conf.Exists()
			if e != nil {
				log.Fatalln(e)
			}
			if exists && !yes {
				fmt.Println(`File already exists:`, conf.Output)
				val := readBool(bufio.NewReader(os.Stdin), `Are you sure you want to overwrite the existing file <y/n>`)
				if !val {
					return
				}
			}

			last := time.Now()
			e = get.NewManager(context.Background(), conf).Serve()
			if e != nil {
				log.Fatalln(e)
			}
			fmt.Println(`success:`, conf.Output, time.Since(last))
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&url,
		`url`, `u`,
		``,
		`http download address`,
	)
	flags.StringVarP(&output,
		`output`, `o`,
		``,
		`download target output file path`,
	)
	flags.StringVarP(&proxy,
		`proxy`, `p`,
		``,
		`socks5://xxx http://xxx`,
	)
	flags.StringSliceVarP(&headers,
		`Header`, `H`,
		[]string{},
		`http request header key: value`,
	)
	flags.StringVarP(&agent,
		`agent`, `a`,
		``,
		`http header User-Agent (default `+metadata.UserAgent+`)`,
	)
	flags.StringSliceVarP(&cookies,
		`cookie`, `c`,
		[]string{},
		`http request cookie`,
	)
	flags.IntVarP(&worker,
		`worker`, `w`,
		runtime.NumCPU(),
		`number of workers performing downloads`,
	)
	flags.StringVarP(&blockStr,
		`block size`, `b`,
		`5m`,
		`download block size for each worker [g m k b]`,
	)
	flags.BoolVarP(&yes,
		`yes`, `y`,
		false,
		`answer yes to all questions`,
	)
	flags.BoolVar(&head,
		`head`,
		false,
		`send HEAD request file meta information before download`,
	)
	flags.BoolVarP(&insecure,
		`insecure`, `k`,
		false,
		`allow insecure server connections when using SSL`,
	)
	if runtime.GOOS == `windows` {
		ascii = true
	}
	flags.BoolVarP(&ascii,
		`ASCII`, `A`,
		ascii,
		`if ASCII is true then use ASCII instead of unicode to draw the`,
	)

	rootCmd.AddCommand(cmd)
}
