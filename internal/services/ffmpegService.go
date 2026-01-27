package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// FFmpegService define la interfaz para operaciones de ffmpeg/ffprobe
type FFmpegService interface {
	ConvertToHLS(ctx context.Context, inputPath, outputDir string) (string, error)
	ExtractDuration(ctx context.Context, videoPath string) (string, error)
	GenerateThumbnail(ctx context.Context, videoPath, outputDir string) (string, error)
}

type ffmpegServiceImp struct {
	hlsTimeout       time.Duration
	thumbnailTimeout time.Duration
	probeTimeout     time.Duration
}

// NewFFmpegService crea una nueva instancia del servicio FFmpeg
func NewFFmpegService() FFmpegService {
	return &ffmpegServiceImp{
		hlsTimeout:       10 * time.Minute,
		thumbnailTimeout: 30 * time.Second,
		probeTimeout:     15 * time.Second,
	}
}

// ConvertToHLS convierte un video a formato HLS usando ffmpeg
// Retorna la ruta de la carpeta con los archivos generados
func (f *ffmpegServiceImp) ConvertToHLS(ctx context.Context, inputPath, outputDir string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, f.hlsTimeout)
	defer cancel()

	outputPath := filepath.Join(outputDir, "output.m3u8")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,
		"-c", "copy",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("ffmpeg HLS timeout después de %v", f.hlsTimeout)
	}
	if err != nil {
		return "", fmt.Errorf("ffmpeg HLS error: %w, output: %s", err, string(output))
	}

	return outputDir, nil
}

// ffprobeOutput estructura para parsear la salida JSON de ffprobe
type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// ExtractDuration obtiene la duración de un video usando ffprobe
func (f *ffmpegServiceImp) ExtractDuration(ctx context.Context, videoPath string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, f.probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		videoPath,
	)

	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("ffprobe timeout después de %v", f.probeTimeout)
	}
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %w", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return "", fmt.Errorf("error parseando ffprobe output: %w", err)
	}

	seconds, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil {
		return "", fmt.Errorf("error convirtiendo duración: %w", err)
	}

	return formatDuration(seconds), nil
}

// GenerateThumbnail genera una miniatura WebP del video
// Retorna la ruta del archivo thumbnail generado
func (f *ffmpegServiceImp) GenerateThumbnail(ctx context.Context, videoPath, outputDir string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, f.thumbnailTimeout)
	defer cancel()

	thumbnailPath := filepath.Join(outputDir, "thumbnail.webp")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", "00:00:08",
		"-i", videoPath,
		"-frames:v", "1",
		"-vf", "scale=480:-1",
		"-y",
		thumbnailPath,
	)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("ffmpeg thumbnail timeout después de %v", f.thumbnailTimeout)
	}
	if err != nil {
		return "", fmt.Errorf("ffmpeg thumbnail error: %w, output: %s", err, string(output))
	}

	return thumbnailPath, nil
}

// formatDuration formatea segundos a formato legible (ej: "1:30" o "45s")
func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}

	minutes := math.Floor(seconds / 60)
	remainingSeconds := math.Round(seconds - (minutes * 60))

	return fmt.Sprintf("%.0f:%.0f", minutes, remainingSeconds)
}
