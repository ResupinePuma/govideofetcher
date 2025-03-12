module videofetcher

go 1.22

require (
	github.com/antchfx/htmlquery v1.3.1
	github.com/dop251/goja v0.0.0-20240220182346-e401ed450204
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/antchfx/xpath v1.3.0 // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)

//replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => ./internal/telegram-bot-api/
