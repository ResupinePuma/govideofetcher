package instagram

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/logging"
	"videofetcher/internal/proxiedHTTP"
)

type notitier struct{}

var client http.Client

func (l *notitier) UpdTextNotify(text string) (err error)       { log.Println(text); return }
func (l *notitier) StartTicker(ctx context.Context) (err error) { return }

func TestMain(m *testing.M) {

	proxyURL, err := url.Parse(os.Getenv("HTTPS_PROXY"))
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client = http.Client{
		Transport: transport,
	}
	Logging = &logging.Logger{}

	os.Exit(m.Run())
}

func TestIG_Download(t *testing.T) {
	type fields struct {
		SizeLimit int
		Timeout   int
		reader    *os.File
	}
	type args struct {
		ctx *dcontext.Context
		url string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantTitle string
		wantRdr   io.ReadCloser
		wantErr   bool
	}{
		{
			name: "ig",
			fields: fields{
				SizeLimit: 50 * 1024 * 1024,
			},
			args: args{
				url: "https://www.instagram.com/reel/DGV6_87i286/?igsh=MXc5dXhlaDBiYXhwZw==",
				ctx: dcontext.NewDownloaderContext(context.Background(), &notitier{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &IG{
				Timeout:   tt.fields.Timeout,
				SizeLimit: int64(tt.fields.SizeLimit),

				Client: proxiedHTTP.NewProxiedHTTPClient(os.Getenv("HTTPS_PROXY")),
			}
			u, _ := url.Parse(tt.args.url)
			tt.args.ctx.SetUrl(u)

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				vid := <-tt.args.ctx.Results()

				if vid == nil {
					t.Error("IG.Download() error = empty file")
				}
				res, _ := os.Create("res.mp4")
				for _, v := range vid {
					_, r, _ := v.UploadData()
					io.Copy(res, r)
				}

			}()

			err := tk.Download(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("IG.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wg.Wait()

		})
	}
}
