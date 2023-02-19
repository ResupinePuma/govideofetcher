package main

import (
	"testing"
)

func Test_extractUrlAndText(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name      string
		args      args
		wantPurl  string
		wantLabel string
		wantErr   bool
	}{
		{
			name: "link",
			args: args{
				str: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			},
			wantPurl: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantLabel: "",
		},
		{
			name: "label_after_link",
			args: args{
				str: "https://www.youtube.com/watch?v=dQw4w9WgXcQ Text here",
			},
			wantPurl:   "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantLabel: "Text here",
		},
		{
			name: "label_before_link",
			args: args{
				str: "Text here https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			},
			wantPurl:   "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantLabel: "Text here",
		},
		{
			name: "some_video_w_label",
			args: args{
				str: "http://10.20.30.40/video.mp4?read Click here",
			},
			wantPurl:   "http://10.20.30.40/video.mp4?read",
			wantLabel: "Click here",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPurl, gotLabel, err := extractUrlAndText(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractUrlAndText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPurl != tt.wantPurl {
				t.Errorf("extractUrlAndText() gotPurl = %v, want %v", gotPurl, tt.wantPurl)
			}
			if gotLabel != tt.wantLabel {
				t.Errorf("extractUrlAndText() gotLabel = %v, want %v", gotLabel, tt.wantLabel)
			}
		})
	}
}
