package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"videofetcher/internal/telemetry"
	ytdlpapi "videofetcher/internal/yt-dlp-api"

	ginzap "github.com/gin-contrib/zap"
	"github.com/samber/lo"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var YTDlPath = "yt-dlp"

// ErrNotFound is returned when no file matches the pattern
var ErrNotFound = fmt.Errorf("file not found")

type Info struct {
	Filesize       float64 `json:"filesize"`        // The number of bytes, if known in advance
	FilesizeApprox float64 `json:"filesize_approx"` // An estimate for the number of bytes
}

type fileResult struct {
	filename string
	typeKey  string
	file     io.ReadCloser
}

// findFile returns the first file matching pattern or ErrNotFound
func findFile(pattern string) (string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		return matches[0], nil
	}
	return "", ErrNotFound
}

func init() {
	otel.SetTextMapPropagator(propagation.TraceContext{}) // можно также использовать composite
}

// waitForFile polls until a file matching pattern appears or context is done
func waitForFile(ctx context.Context, dir, pattern string) *fileResult {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(time.Second)
			path, err := findFile(filepath.Join(dir, pattern))
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					continue
				}
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return nil
			}
			return &fileResult{filename: filepath.Base(path), file: f}
		}
	}
}

func parseSize(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	// Последний символ определяет суффикс (K, M, G, T и т. д.)
	unit := s[len(s)-1]
	// Числовая часть без суффикса
	numStr := s[:len(s)-1]

	// Парсим число
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number part: %w", err)
	}

	// Определяем множитель
	var multiplier int64 = 1
	switch strings.ToUpper(string(unit)) {
	case "K":
		multiplier = 1 << 10 // 1024
	case "M":
		multiplier = 1 << 20 // 1 048 576
	case "G":
		multiplier = 1 << 30 // 1 073 741 824
	case "T":
		multiplier = 1 << 40
	default:
		// Если последний символ не буква, считаем, что это просто байты
		// и переходим к парсингу всей строки как числа
		return strconv.ParseInt(s, 10, 64)
	}

	// Умножаем и возвращаем
	bytes := int64(num * float64(multiplier))
	return bytes, nil
}

func errorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		errs := c.Errors
		if len(errs) > 0 {
			err := errs[0].Err
			log.Printf("request error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
		}
	}
}
func randomFileName(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}
func ToPng(imageBytes []byte) ([]byte, error) {
	contentType := http.DetectContentType(imageBytes)

	switch contentType {
	case "image/png":
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, fmt.Errorf("unable to decode jpeg: %w", err)
		}

		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("unable to encode png: %w", err)
		}

		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unable to convert %#v to png", contentType)
}

func main() {
	r := gin.New()

	serviceName := flag.String("serviceName", "", "name of service for telemetry")
	address := flag.String("otelAddress", "", "address of telemetry server")
	flag.Parse()

	telemetry.ServiceName = *serviceName
	telemetry.TracerEndpoint = *address

	ctx := context.Background()
	telemetry.InitTracer(ctx)

	meter := telemetry.InitMeter()

	requestDuration, err := meter.Float64Histogram(
		"http.server.duration",
		metric.WithDescription("duration of HTTP requests in seconds"),
	)
	if err != nil {
		log.Fatalf("failed to create histogram: %v", err)
	}

	zapCFG := zap.NewProductionConfig()
	zapCFG.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapCFG.DisableCaller = true
	zapCFG.Encoding = "json"
	logger, _ := zapCFG.Build()

	sugar := logger.Sugar()
	defer logger.Sync() // Flushes buffer, if any

	r.Use(ginzap.Ginzap(logger, time.RFC3339, true), errorMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/convert", otelgin.Middleware(telemetry.ServiceName,
		otelgin.WithGinMetricAttributeFn(func(c *gin.Context) []attribute.KeyValue {
			start := time.Now()
			c.Next()
			duration := time.Since(start).Seconds()

			attrs := []attribute.KeyValue{
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPTargetKey.String(c.Request.URL.Path),
				semconv.HTTPStatusCodeKey.Int(c.Writer.Status()),
			}
			// Сохраняем кастомную метрику
			requestDuration.Record(c.Request.Context(), duration)

			return attrs
		}),
	), func(c *gin.Context) {
		// Извлекаем trace ID и кладём в gin.Context
		span := trace.SpanFromContext(c.Request.Context())
		traceID := span.SpanContext().TraceID().String()
		c.Set("trace_id", traceID)

		// Продолжаем цепочку
		c.Next()
	},
		func(c *gin.Context) {
			ctx, cancel := context.WithCancel(c.Request.Context())
			defer cancel()

			var args ytdlpapi.BodyArgs
			if err := c.BindJSON(&args); err != nil {
				c.Error(err)
				return
			}

			sugar.Infow("got message", "params", args, "trace_id", c.GetString("trace_id"))

			filename := randomFileName(8)

			if args.Format == "" || args.Extension == "" || args.URL == "" {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			args.Extension = strings.TrimPrefix(args.Extension, ".")
			mformats := []string{"avi", "flv", "mkv", "mov", "mp4", "webm"}
			cmd := exec.CommandContext(ctx, YTDlPath,
				"--no-call-home",
				"--no-cache-dir",
				"--ignore-errors",
				"--no-abort-on-error",
				"--proxy", args.ProxyURL,
				"--newline",
				"-N", "8",
				"--restrict-filenames",
				"-f", args.Format,
			)
			
			if lo.Contains(mformats, args.Extension) {
				cmd.Args = append(cmd.Args, "--merge-output-format", args.Extension)
			}
			args.Extension = "." + args.Extension

			if args.BufferSize != "" {
				cmd.Args = append(cmd.Args, "--buffer-size", args.BufferSize)
			}
			if args.DownloadThumb {
				cmd.Args = append(cmd.Args, "--write-thumbnail", "--convert-thumbnails", "jpg")
			}
			if args.DownloadInfo {
				cmd.Args = append(cmd.Args, "--write-info-json")
			}
			if args.MaxFilesize != "" {
				s, err := parseSize(args.MaxFilesize)
				if err != nil {
					c.Error(err)
					return
				}
				args.MaxSize = s
				cmd.Args = append(cmd.Args, "--max-filesize", args.MaxFilesize)
			}
			if args.Headers != nil {
				for k, v := range args.Headers {
					line := fmt.Sprintf("%s: %s", k, strings.Join(v, "; "))
					cmd.Args = append(cmd.Args, "--add-header", line)
				}
			}

			if args.FFMpeg != "" {
				cmd.Args = append(cmd.Args, "-o", filename+args.Extension)
				cmd.Args = append(cmd.Args, "--exec", fmt.Sprintf("%s -i {} %s %s", "ffmpeg", args.FFMpeg, "conv_"+filename+args.Extension))
				cmd.Args = append(cmd.Args, args.URL)
			} else {
				cmd.Args = append(cmd.Args, args.URL)
				cmd.Args = append(cmd.Args, "-o", filename+args.Extension)
			}

			tempDir, err := os.MkdirTemp("", "ydls")
			if err != nil {
				c.Error(err)
				return
			}
			defer os.RemoveAll(tempDir)

			cmd.Dir = tempDir
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				c.Error(err)
				return
			}

			stderr, err := cmd.StderrPipe()
			if err != nil {
				c.Error(err)
				return
			}

			go streamLogs(stdout, c.GetString("trace_id"), logger)
			go streamLogs(stderr, c.GetString("trace_id"), logger)

			if err := cmd.Start(); err != nil {
				sugar.Errorw("got error", "error", err, "trace_id", c.GetString("trace_id"))
				c.Error(err)
				return
			}

			var multipartWriter *multipart.Writer
			var setupMultipart = func() {
				multipartWriter = multipart.NewWriter(c.Writer)
				c.Header("Content-Type", multipartWriter.FormDataContentType())
			}

			errorCh := make(chan error, 1)
			filesChan := make(chan fileResult)
			var wg sync.WaitGroup

			if args.DownloadInfo {
				wg.Add(1)
				go func(msize int64) {
					var info Info
					defer wg.Done()
					if res := waitForFile(ctx, tempDir, "*.info.json"); res != nil {
						res.typeKey = "info"
						b, err := io.ReadAll(res.file)
						if err != nil {
							errorCh <- fmt.Errorf("err read info: %v", err)
							return
						}

						err = json.Unmarshal(b, &info)
						if err != nil {
							errorCh <- fmt.Errorf("err decode info: %v", err)
							return
						}
						res.file.Close()

						size := max(info.Filesize, info.FilesizeApprox)
						if msize > 0 && size > float64(msize) {
							errorCh <- fmt.Errorf("size limit reached")
							return
						}

						res.file = io.NopCloser(bytes.NewReader(b))

						filesChan <- *res
					}
				}(args.MaxSize)
			}

			if args.DownloadThumb {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if res := waitForFile(ctx, tempDir, "*.jpg"); res != nil {
						res.typeKey = "thumb"

						b, err := io.ReadAll(io.LimitReader(res.file, 10*1024*1024*1024))
						if err != nil {
							errorCh <- err
							return
						}
						res.file.Close()

						pngb, err := ToPng(b)
						if err != nil {
							errorCh <- err
							return
						}

						r := io.NopCloser(bytes.NewReader(pngb))
						fname := strings.TrimSuffix(res.filename, ".jpg")
						filesChan <- fileResult{filename: fname + ".png", file: r}
					}
				}()
			}

			go func() {
				wg.Wait()
				close(filesChan)
				if multipartWriter != nil {
					multipartWriter.Close()
				}
			}()

			// stream each found file
			go func() {
				for fil := range filesChan {
					if multipartWriter == nil {
						setupMultipart()
					}

					if fil.file == nil {
						continue
					}

					writer, err := multipartWriter.CreateFormFile(fil.typeKey, fil.filename)
					if err != nil {
						errorCh <- err
						fil.file.Close()
						return
					}
					_, err = io.Copy(writer, fil.file)
					if err != nil {
						errorCh <- err
						fil.file.Close()
						return
					}
				}
			}()

			wg.Add(1)
			defer wg.Done()

			if err := cmd.Wait(); err != nil {
				sugar.Errorw("got error", "error", err, "trace_id", c.GetString("trace_id"))
				return
			}

			if !cmd.ProcessState.Success() {
				return
			}

			var vf io.ReadCloser
			fname := ""
			if args.FFMpeg != "" {
				f, err := os.Open(filepath.Join(tempDir, "conv_"+filename+args.Extension))
				if err != nil && errors.Is(err, fs.ErrNotExist) {
					return
				}
				fname = "conv_" + filename + args.Extension
				vf = f
			} else {
				f, err := os.Open(filepath.Join(tempDir, filename+args.Extension))
				if err != nil && errors.Is(err, fs.ErrNotExist) {
					return
				}
				fname = filename + args.Extension
				vf = f
			}

			if multipartWriter == nil {
				setupMultipart()
			}

			writer, err := multipartWriter.CreateFormFile("media", fname)
			if err != nil {
				c.Error(err)
				vf.Close()
				return
			}
			io.Copy(writer, vf)
			vf.Close()

		})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}

	sugar.Infof("fast YTDLP API listening on port %s", port)
	srv.ListenAndServe()
}
