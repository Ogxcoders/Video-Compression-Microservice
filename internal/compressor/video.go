package compressor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/yourusername/video-compressor/internal/models"
)

type VideoCompressor struct {
	ffmpegPath string
	tempDir    string
}

func NewVideoCompressor(ffmpegPath, tempDir string) *VideoCompressor {
	return &VideoCompressor{
		ffmpegPath: ffmpegPath,
		tempDir:    tempDir,
	}
}

func (v *VideoCompressor) Compress(inputPath string, quality models.VideoQuality) (string, error) {
	outputPath := filepath.Join(v.tempDir, fmt.Sprintf("compressed_%d.mp4", time.Now().Unix()))

	var args []string
	args = append(args, "-i", inputPath)

	switch quality {
	case models.VideoQualityLow:
		args = append(args, "-vf", "scale=854:480", "-b:v", "1000k", "-c:v", "libx264", "-preset", "fast")
	case models.VideoQualityMedium:
		args = append(args, "-vf", "scale=1280:720", "-b:v", "2500k", "-c:v", "libx264", "-preset", "medium")
	case models.VideoQualityHigh:
		args = append(args, "-vf", "scale=1920:1080", "-b:v", "5000k", "-c:v", "libx264", "-preset", "slow")
	case models.VideoQualityUltra:
		args = append(args, "-b:v", "8000k", "-c:v", "libx264", "-preset", "slow")
	default:
		return "", fmt.Errorf("unsupported quality: %s", quality)
	}

	args = append(args, "-c:a", "aac", "-b:a", "128k", "-movflags", "+faststart", "-y", outputPath)

	cmd := exec.Command(v.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return outputPath, nil
}

func (v *VideoCompressor) GenerateHLS(inputPath string, variants []string) (string, map[string]string, error) {
	hlsDir := filepath.Join(v.tempDir, fmt.Sprintf("hls_%d", time.Now().Unix()))
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create HLS directory: %w", err)
	}

	variantURLs := make(map[string]string)
	masterPlaylist := filepath.Join(hlsDir, "master.m3u8")

	masterContent := "#EXTM3U\n#EXT-X-VERSION:3\n"

	for _, variant := range variants {
		variantDir := filepath.Join(hlsDir, variant)
		if err := os.MkdirAll(variantDir, 0755); err != nil {
			return "", nil, fmt.Errorf("failed to create variant directory: %w", err)
		}

		playlistPath := filepath.Join(variantDir, "playlist.m3u8")

		var scale, bitrate, bandwidth string
		switch variant {
		case "480p":
			scale = "854:480"
			bitrate = "1000k"
			bandwidth = "1000000"
		case "720p":
			scale = "1280:720"
			bitrate = "2500k"
			bandwidth = "2500000"
		case "1080p":
			scale = "1920:1080"
			bitrate = "5000k"
			bandwidth = "5000000"
		default:
			continue
		}

		args := []string{
			"-i", inputPath,
			"-vf", fmt.Sprintf("scale=%s", scale),
			"-b:v", bitrate,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-b:a", "128k",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-hls_segment_filename", filepath.Join(variantDir, "segment-%03d.ts"),
			"-f", "hls",
			playlistPath,
		}

		cmd := exec.Command(v.ffmpegPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", nil, fmt.Errorf("ffmpeg HLS failed for %s: %w, output: %s", variant, err, string(output))
		}

		masterContent += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%s,RESOLUTION=%s\n", bandwidth, scale)
		masterContent += fmt.Sprintf("%s/playlist.m3u8\n", variant)

		variantURLs[variant] = fmt.Sprintf("%s/playlist.m3u8", variant)
	}

	if err := os.WriteFile(masterPlaylist, []byte(masterContent), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write master playlist: %w", err)
	}

	return masterPlaylist, variantURLs, nil
}

func (v *VideoCompressor) GetVideoInfo(inputPath string) (int64, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
