package ytdl

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
	"videofetcher/internal/downloader/dcontext"
)

type logg struct{}

func (l *logg) Debugf(ctx context.Context, format string, v ...interface{}) { log.Printf(format, v...) }
func (l *logg) Errorf(ctx context.Context, format string, v ...interface{}) { log.Printf(format, v...) }
func (l *logg) Warnf(ctx context.Context, format string, v ...interface{})  { log.Printf(format, v...) }
func (l *logg) Infof(ctx context.Context, format string, v ...interface{})  { log.Printf(format, v...) }
func (l *logg) Print(v ...interface{})                                      { log.Print(v...) }

type notitier struct{}

func (l *notitier) UpdTextNotify(text string) (err error)       { log.Println(text); return }
func (l *notitier) StartTicker(ctx context.Context) (err error) { return }

func TestYTdl_Download(t *testing.T) {
	Logging = &logg{}

	type fields struct {
		SizeLimit float64
		Timeout   int
	}
	type args struct {
		ctx context.Context
		u   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nvgyp",
			fields: fields{
				SizeLimit: 50 * 1024 * 1024,
				Timeout:   30,
			},
			args: args{
				u:   "https://youtube.com/shorts/69Hyj2tWVbc",
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yt := &YtDl{
				SizeLimit: 50 * 1024 * 1024,
				Timeout:   time.Duration(tt.fields.Timeout),
				Format:    "18/17/bestvideo+worstaudio/(mp4)[ext=mp4][vcodec^=h26]/worst[width>=480][ext=mp4]/worst[ext=mp4]",
				mode:      modeVideo,
				ProxyURL:  "http://172.30.1.2:3128",
			}
			var err error
			//var vid io.Reader
			u, _ := url.Parse(tt.args.u)
			ccc := dcontext.NewDownloaderContext(context.Background(), &notitier{})
			ccc.SetUrl(u)

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				vid := <-ccc.Results()

				if vid == nil {
					t.Error("YTdl.Download() error = empty file")
				}
				res, _ := os.Create("res.mp4")
				for _, v := range vid {
					_, r, _ := v.UploadData()
					io.Copy(res, r)
				}

			}()

			err = yt.Download(ccc)
			if (err != nil) != tt.wantErr {
				t.Errorf("YTdl.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wg.Wait()

		})
	}
}
