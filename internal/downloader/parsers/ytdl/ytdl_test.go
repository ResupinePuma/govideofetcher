package ytdl

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"
	"testing"
	"videofetcher/internal/downloader/dcontext"
	"videofetcher/internal/downloader/dresult"
)

type logg struct{}

func (l *logg) Debugf(format string, v ...interface{}) { log.Printf(format, v...) }
func (l *logg) Errorf(format string, v ...interface{}) { log.Printf(format, v...) }
func (l *logg) Warnf(format string, v ...interface{})  { log.Printf(format, v...) }
func (l *logg) Infof(format string, v ...interface{})  { log.Printf(format, v...) }
func (l *logg) Print(v ...interface{})                 { log.Print(v...) }

type notitier struct{}

func (l *notitier) UpdTextNotify(text string) (err error)       { log.Println(text); return }
func (l *notitier) StartTicker(ctx context.Context) (err error) { return }

func TestYTdl_Download(t *testing.T) {
	Logger = &logg{}

	type fields struct {
		downloadResult *dresult.DownloadResult
		SizeLimit      float64
		Timeout        int
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
				u:   "https://youtube.com/shorts/mU8sm-1u504",
				ctx: context.Background(),
			},
		},
		// {
		// 	name: "9gag",
		// 	fields: fields{
		// 		SizeLimit: 50 * 1024 * 1024,
		// 		Timeout:   30,
		// 	},
		// 	args: args{
		// 		url: "https://youtube.com/shorts/QbGyVwblgQQ?feature=share",
		// 		ctx: context.Background(),
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yt := &YtDl{
				//downloadResult: tt.fields.downloadResult,
				//SizeLimit:      int(tt.fields.SizeLimit),
				SizeLimit: 50 * 1024 * 1024,
				Timeout:   tt.fields.Timeout,
				Format:    "18/17/bestvideo+worstaudio/(mp4)[ext=mp4][vcodec^=h26]/worst[width>=480][ext=mp4]/worst[ext=mp4]",
			}
			var err error
			//var vid io.Reader
			u, _ := url.Parse(tt.args.u)
			ccc := dcontext.NewDownloaderContext(context.Background(), &notitier{})
			vid, err := yt.Download(ccc, u)
			if (err != nil) != tt.wantErr {
				t.Errorf("YTdl.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if vid == nil {
				t.Error("YTdl.Download() error = empty file")
			}
			res, _ := os.Create("res.mp4")
			for _, v := range vid {
				io.Copy(res, v.Reader)
			}

		})
	}
}
