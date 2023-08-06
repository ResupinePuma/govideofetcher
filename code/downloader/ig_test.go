package downloader

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestIG_Download(t *testing.T) {
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
			name: "ig",
			fields: fields{
				log:       &logg{},
				ntf:       &notitier{},
				SizeLimit: 50 * 1024 * 1024,
			},
			args: args{
				url: "https://instagram.com/stories/4chanvideo.x/3163406295553564208?igshid=YTUzYTFiZDMwYg==",
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &IG{
				SizeLimit: tt.fields.SizeLimit,
				Timeout:   tt.fields.Timeout,
				log:       tt.fields.log,
				ntf:       tt.fields.ntf,
				StoryDUrl: "https://igram.world/api/ig/story?url=%s",
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
