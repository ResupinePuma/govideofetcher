package ytdl

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/ResupinePuma/goutubedl"
)

type logg struct{}

func (l *logg) Debug(format string, v ...interface{})   { log.Printf(format, v...) }
func (l *logg) Error(format string, v ...interface{})   { log.Printf(format, v...) }
func (l *logg) Warning(format string, v ...interface{}) { log.Printf(format, v...) }
func (l *logg) Info(format string, v ...interface{})    { log.Printf(format, v...) }

type notitier struct{}

func (l *notitier) Message(text string) (err error)   { log.Println(text); return }
func (l *notitier) Count(percent float64) (err error) { log.Println(percent); return }

func TestYTdl_Download(t *testing.T) {
	type fields struct {
		downloadResult *goutubedl.DownloadResult
		log            YTdlLog
		ntf            AbstractNotifier
		SizeLimit      float64
		Timeout        int
	}
	type args struct {
		ctx context.Context
		url string
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
				log:       YTdlLog{AbstractLogger: &logg{}},
				ntf:       &notitier{},
				SizeLimit: 50 * 1024 * 1024,
				Timeout:   30,
			},
			args: args{
				url: "https://www.youtube.com/watch?v=zGDzdps75ns",
				ctx: context.Background(),
			},
		},
		{
			name: "9gag",
			fields: fields{
				log:       YTdlLog{AbstractLogger: &logg{}},
				ntf:       &notitier{},
				SizeLimit: 50 * 1024 * 1024,
				Timeout:   30,
			},
			args: args{
				url: "https://youtube.com/shorts/QbGyVwblgQQ?feature=share",
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yt := &YTdl{
				downloadResult: tt.fields.downloadResult,
				log:            tt.fields.log,
				ntf:            tt.fields.ntf,
				SizeLimit:      int(tt.fields.SizeLimit),
				Timeout:        tt.fields.Timeout,
			}
			var err error
			var vid io.Reader
			_, vid, err = yt.Download(tt.args.ctx, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("YTdl.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if vid == nil {
				t.Error("YTdl.Download() error = empty file")
			}

		})
	}
}
