package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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
	ytdlpapi "videofetcher/internal/yt-dlp-api"

	"github.com/gin-gonic/gin"
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

// waitForFile polls until a file matching pattern appears or context is done
func waitForFile(ctx context.Context, dir, pattern string, errorCh chan<- error) *fileResult {
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
				errorCh <- err
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				errorCh <- fmt.Errorf("error opening file %s: %w", path, err)
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
func main() {
	r := gin.New()

	r.Use(gin.Logger(), errorMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/convert", func(c *gin.Context) {
		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		var args ytdlpapi.BodyArgs
		if err := c.BindJSON(&args); err != nil {
			c.Error(err)
			return
		}

		filename := randomFileName(8)

		if args.Format == "" || args.Extension == "" || args.URL == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		args.Extension = strings.TrimPrefix(args.Extension, ".")
		args.Extension = "." + args.Extension

		cmd := exec.CommandContext(ctx, YTDlPath,
			"--no-call-home",
			"--no-cache-dir",
			"--abort-on-error",
			"--proxy", args.ProxyURL,
			"--newline",
			"-N", "8",
			"--restrict-filenames",
			"-f", args.Format,
		)
		if args.BufferSize != "" {
			cmd.Args = append(cmd.Args, "--buffer-size", args.BufferSize)
		}
		if args.DownloadThumb {
			cmd.Args = append(cmd.Args, "--write-thumbnail", "--convert-thumbnails", "png")
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
		cmd.Stderr = log.Writer()
		cmd.Stdout = log.Writer()

		if err := cmd.Start(); err != nil {
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
				if res := waitForFile(ctx, tempDir, "*.info.json", errorCh); res != nil {
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
				if res := waitForFile(ctx, tempDir, "*.png", errorCh); res != nil {
					res.typeKey = "thumb"
					filesChan <- *res
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
			c.Error(err)
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

	log.Printf("Fast YTDLP API listening on port %s", port)
	srv.ListenAndServe()
}
