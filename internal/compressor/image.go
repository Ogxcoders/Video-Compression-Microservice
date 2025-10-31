package compressor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/yourusername/video-compressor/internal/models"
)

type ImageCompressor struct {
	imageMagickPath string
	tempDir         string
}

func NewImageCompressor(imageMagickPath, tempDir string) *ImageCompressor {
	return &ImageCompressor{
		imageMagickPath: imageMagickPath,
		tempDir:         tempDir,
	}
}

func (i *ImageCompressor) CompressWithVariants(inputPath string, quality models.ImageQuality, variants []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, variant := range variants {
		outputPath, err := i.generateVariant(inputPath, variant, quality)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s variant: %w", variant, err)
		}
		results[variant] = outputPath
	}

	return results, nil
}

func (i *ImageCompressor) generateVariant(inputPath, variant string, quality models.ImageQuality) (string, error) {
	ext := filepath.Ext(inputPath)
	outputPath := filepath.Join(i.tempDir, fmt.Sprintf("%s_%d%s", variant, time.Now().Unix(), ext))

	var args []string
	args = append(args, inputPath)

	qualityValue := i.getQualityValue(quality, variant)

	switch variant {
	case "thumbnail":
		args = append(args, "-resize", "150x150^", "-gravity", "center", "-extent", "150x150")
	case "medium":
		args = append(args, "-resize", "400x300")
	case "large":
		args = append(args, "-resize", "800x600")
	case "original":
	default:
		return "", fmt.Errorf("unsupported variant: %s", variant)
	}

	args = append(args, "-quality", fmt.Sprintf("%d", qualityValue), outputPath)

	cmd := exec.Command(i.imageMagickPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("imagemagick failed: %w, output: %s", err, string(output))
	}

	return outputPath, nil
}

func (i *ImageCompressor) getQualityValue(quality models.ImageQuality, variant string) int {
	baseQuality := map[models.ImageQuality]int{
		models.ImageQualityLow:    60,
		models.ImageQualityMedium: 75,
		models.ImageQualityHigh:   85,
		models.ImageQualityUltra:  95,
	}

	q := baseQuality[quality]

	if variant == "thumbnail" && q > 75 {
		return 75
	}
	if variant == "original" && q < 95 {
		return 95
	}

	return q
}

func (i *ImageCompressor) GetImageInfo(imagePath string) (int64, string, error) {
	info, err := os.Stat(imagePath)
	if err != nil {
		return 0, "", err
	}

	cmd := exec.Command("identify", "-format", "%wx%h", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return info.Size(), "", nil
	}

	return info.Size(), string(output), nil
}
