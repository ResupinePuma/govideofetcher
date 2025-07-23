package notice

type nerror struct {
	notice
}

func NewError(text string) *nerror {
	return &nerror{notice: notice{Text: text}}
}
func (e *nerror) As(a any) bool {
	if n, ok := a.(*nerror); ok {
		*n = *e
		return true
	}
	return false
}

func (e nerror) Error() string {
	return e.Text
}

var (
	ErrSizeLimitReached   = NewError("size limit reached")
	ErrTimeout            = NewError("timeout")
	ErrInvalidResponse    = NewError("external service has invalid response")
	ErrNotFound           = NewError("not found")
	ErrUnsupportedService = NewError("unsuppoted service")
	ErrInvalidURL         = NewError("invalid url")
	ErrUnexpectedError    = NewError("unexpected error")
)
