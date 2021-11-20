package cmd

import (
	"bufio"
	"context"
	"log"
	"os"
	"runtime"

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
	)
	cmd := &cobra.Command{
		Use:   `get`,
		Short: `http get download file`,
		Example: `mget get -u http://127.0.0.1/tools/source.exe
mget get -u http://127.0.0.1/tools/source.exe -o a.exe`,
		Run: func(cmd *cobra.Command, args []string) {
			conf, e := metadata.NewConfigure(url, output, proxy,
				agent, head, headers, cookies, insecure,
				worker, blockStr,
			)
			if e != nil {
				log.Fatalln(e)
			}
			conf.Println()
			if !yes {
				val := readBool(bufio.NewReader(os.Stdin), `Are you sure you want to start downloading <y/n>`)
				if !val {
					return
				}
			}
			e = get.NewManager(context.Background(), conf).Serve()
			if e != nil {
				log.Fatalln(e)
			}
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
	rootCmd.AddCommand(cmd)
}
