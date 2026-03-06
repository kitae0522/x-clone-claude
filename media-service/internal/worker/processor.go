package worker

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/model"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/storage"

	_ "golang.org/x/image/webp"
)

// Job represents a media processing job.
type Job struct {
	ID          string
	File        io.ReadSeeker
	ContentType string
	MediaType   model.MediaType
	Size        int64
}

// Processor handles async media processing with a bounded worker pool.
type Processor struct {
	store    storage.ObjectStorage
	registry *Registry
	queue    chan Job
	tempDir  string
}

func NewProcessor(store storage.ObjectStorage, reg *Registry, maxWorkers int, tempDir string) *Processor {
	p := &Processor{
		store:    store,
		registry: reg,
		queue:    make(chan Job, 256),
		tempDir:  tempDir,
	}

	if err := os.MkdirAll(tempDir, 0700); err != nil {
		slog.Error("failed to create temp dir", "error", err)
	}

	for i := range maxWorkers {
		go p.worker(i)
	}
	slog.Info("media processor started", "workers", maxWorkers)

	return p
}

func (p *Processor) Enqueue(job Job) {
	p.queue <- job
}

func (p *Processor) worker(id int) {
	for job := range p.queue {
		slog.Info("processing job", "worker", id, "job", job.ID, "type", job.MediaType)
		p.registry.UpdateStatus(job.ID, model.StatusProcessing, "")

		if err := p.process(job); err != nil {
			slog.Error("job failed", "worker", id, "job", job.ID, "error", err)
			p.registry.UpdateStatus(job.ID, model.StatusFailed, err.Error())
		} else {
			slog.Info("job completed", "worker", id, "job", job.ID)
			p.registry.UpdateStatus(job.ID, model.StatusReady, "")
		}
	}
}

func (p *Processor) process(job Job) error {
	// Save uploaded file to temp disk
	tmpPath := filepath.Join(p.tempDir, job.ID+".tmp")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, job.File); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	switch job.MediaType {
	case model.MediaTypeImage, model.MediaTypeGIF:
		return p.processImage(job.ID, tmpPath)
	case model.MediaTypeVideo:
		return p.processVideo(job.ID, tmpPath)
	default:
		return fmt.Errorf("unknown media type: %s", job.MediaType)
	}
}

func (p *Processor) processImage(id string, srcPath string) error {
	src, err := imaging.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}

	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	p.registry.UpdateDimensions(id, origW, origH)

	ctx := context.Background()

	for _, variant := range model.AllVariants {
		maxW := model.VariantMaxWidth[variant]
		resized := src

		if origW > maxW {
			resized = imaging.Resize(src, maxW, 0, imaging.Lanczos)
		}

		var buf bytes.Buffer
		if err := imaging.Encode(&buf, resized, imaging.PNG); err != nil {
			return fmt.Errorf("encode %s: %w", variant, err)
		}

		// Convert PNG to WebP via cwebp
		webpBuf, err := pngToWebP(buf.Bytes())
		if err != nil {
			return fmt.Errorf("webp convert %s: %w", variant, err)
		}

		key := string(variant) + "/" + id + ".webp"
		if err := p.store.Put(ctx, key, bytes.NewReader(webpBuf), "image/webp"); err != nil {
			return fmt.Errorf("upload %s: %w", variant, err)
		}
	}

	return nil
}

func (p *Processor) processVideo(id string, srcPath string) error {
	ctx := context.Background()

	// Probe original dimensions
	w, h, err := probeVideoDimensions(srcPath)
	if err == nil {
		p.registry.UpdateDimensions(id, w, h)
	}

	for _, variant := range model.AllVariants {
		maxW := model.VariantMaxWidth[variant]
		outPath := filepath.Join(p.tempDir, fmt.Sprintf("%s_%s.webm", id, variant))

		scaleFilter := fmt.Sprintf("scale='min(%d,iw)':-2", maxW)

		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-i", srcPath,
			"-c:v", "libvpx",
			"-b:v", "1M",
			"-c:a", "libvorbis",
			"-vf", scaleFilter,
			"-deadline", "realtime",
			"-cpu-used", "4",
			"-y",
			outPath,
		)
		cmd.Stderr = io.Discard

		if err := cmd.Run(); err != nil {
			os.Remove(outPath)
			return fmt.Errorf("ffmpeg %s: %w", variant, err)
		}

		f, err := os.Open(outPath)
		if err != nil {
			os.Remove(outPath)
			return fmt.Errorf("open output %s: %w", variant, err)
		}

		key := string(variant) + "/" + id + ".webm"
		if err := p.store.Put(ctx, key, f, "video/webm"); err != nil {
			f.Close()
			os.Remove(outPath)
			return fmt.Errorf("upload %s: %w", variant, err)
		}

		f.Close()
		os.Remove(outPath)
	}

	return nil
}

func pngToWebP(pngData []byte) ([]byte, error) {
	cmd := exec.Command("cwebp", "-quiet", "-o", "-", "--", "-")
	cmd.Stdin = bytes.NewReader(pngData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cwebp: %w", err)
	}
	return out.Bytes(), nil
}

func probeVideoDimensions(path string) (int, int, error) {
	wCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		path,
	)
	out, err := wCmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var w, h int
	_, err = fmt.Sscanf(string(out), "%dx%d", &w, &h)
	if err != nil {
		// Try alternative parsing
		parts := bytes.Split(bytes.TrimSpace(out), []byte("x"))
		if len(parts) == 2 {
			w, _ = strconv.Atoi(string(parts[0]))
			h, _ = strconv.Atoi(string(parts[1]))
		}
	}
	return w, h, nil
}

// Decode image dimensions without loading entire image into memory.
func init() {
	// Register decoders via blank imports above.
	_ = image.Decode
}
