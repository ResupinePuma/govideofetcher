package tiktok

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
	"videofetcher/internal/proxiedHTTP"
)

type notitier struct{}

var client http.Client

func (l *notitier) UpdTextNotify(text string) (err error)       { log.Println(text); return }
func (l *notitier) StartTicker(ctx context.Context) (err error) { return }

func TestTikTok_Download(t *testing.T) {
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
			name: "tt",
			fields: fields{
				SizeLimit: 50 * 1024 * 1024,
			},
			args: args{
				url: "https://www.tiktok.com/@za_vali/video/7053747039313743105?is_from_webapp=1&sender_device=pc&web_id=7478371002204931639",
				ctx: dcontext.NewDownloaderContext(context.Background(), &notitier{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &TikTok{
				SizeLimit: 10000000,
				Client:    proxiedHTTP.NewProxiedHTTPClient(os.Getenv("HTTPS_PROXY")),
			}

			u, _ := url.Parse(tt.args.url)
			tt.args.ctx.SetUrl(u)

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				vid := <-tt.args.ctx.Results()

				if vid == nil {
					t.Error("YTdl.Download() error = empty file")
				}
				res, _ := os.Create("res.mp4")
				for _, v := range vid {
					_, r, _ := v.UploadData()
					io.Copy(res, r)
				}

			}()

			err := tk.Download(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("TikTok.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wg.Wait()

		})
	}
}
