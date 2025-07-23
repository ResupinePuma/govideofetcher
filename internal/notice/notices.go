package notice

type notice struct {
	Text string
}

func New(text string) notice {
	return notice{Text: text}
}

func (n notice) String() string {
	return n.Text
}

var (
	NoticeMediaFound = New("📣 I found %v media. Start sending")
	NoticeGotLink    = New("Got link. 👀 at 📼")
	NoticeDone       = New("Done! ✅")

	NoticeUsageGet         = New("<URL> <caption (optional)> - Downloаd video from URL")
	NoticeUsageNotFound    = New("command not found")
	NoticeDownloadingMedia = New("⏬ downloading media")
)
