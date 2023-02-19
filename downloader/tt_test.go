package downloader

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestTikTok_Download(t *testing.T) {
	type fields struct {
		SizeLimit int
		Timeout   int
		reader    *os.File
		log       AbstractLogger
		ntf       AbstractNotifier
	}
	type args struct {
		ctx context.Context
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
				log:       &logg{},
				ntf:       &notitier{},
				SizeLimit: 50 * 1024 * 1024,
			},
			args: args{
				url: "https://vt.tiktok.com/ZS8axpGa3/",
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &TikTok{
				SizeLimit: tt.fields.SizeLimit,
				Timeout:   tt.fields.Timeout,
				log:       tt.fields.log,
				ntf:       tt.fields.ntf,
			}
			_, vid, err := tk.Download(tt.args.ctx, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("TikTok.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if vid == nil {
				t.Error("TikTok.Download() error = empty file")
			}

			v, err := os.Create("tmp.mp4")
			io.Copy(v, vid)

		})
	}
}
