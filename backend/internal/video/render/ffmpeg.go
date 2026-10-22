package render

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

// FFmpegJobConfig carries everything needed to generate the ffmpeg CLI command.
type FFmpegJobConfig struct {
	Spec             VideoSpec
	ResolutionPreset string   // "720p" or "1080p"
	ImagePaths       []string // Local file paths for each clip in spec.Clips order
	AudioPath        string   // Local file path for audio, or empty
	FontDir          string   // Directory containing the TTF fonts
	OutputPath       string   // Final destination MP4 file path
}

// BuildFFmpegArgs generates the complete command line arguments for the ffmpeg executable.
// Format: ffmpeg <args...>
func BuildFFmpegArgs(cfg FFmpegJobConfig) ([]string, error) {
	spec := cfg.Spec
	if len(spec.Clips) == 0 {
		return nil, fmt.Errorf("cannot render empty clip list")
	}
	if len(cfg.ImagePaths) != len(spec.Clips) {
		return nil, fmt.Errorf("mismatch between clip count (%d) and image path count (%d)", len(spec.Clips), len(cfg.ImagePaths))
	}

	res, ok := OutputResolution(spec.Output.Aspect, cfg.ResolutionPreset)
	if !ok {
		res = Resolution{W: 1080, H: 1920}
	}
	outW, outH := res.W, res.H
	fps := spec.Output.FPS
	if fps <= 0 {
		fps = 30
	}

	var args []string
	// Overwrite without asking
	args = append(args, "-y")

	// 1. Input image files
	for i, imgPath := range cfg.ImagePaths {
		clip := spec.Clips[i]
		// Add input with loop and duration
		args = append(args,
			"-loop", "1",
			"-t", fmt.Sprintf("%.4f", clip.Duration),
			"-i", imgPath,
		)
	}

	// 2. Input audio file (if provided)
	hasAudio := cfg.AudioPath != "" && spec.Audio != nil
	audioInputIndex := len(cfg.ImagePaths)
	if hasAudio {
		if spec.Audio.TrimStart > 0 {
			args = append(args, "-ss", fmt.Sprintf("%.4f", spec.Audio.TrimStart))
		}
		args = append(args, "-i", cfg.AudioPath)
	}

	// 3. Build filter_complex graph
	var filterComplex []string

	// Per-clip visual pipelines
	for i, clip := range spec.Clips {
		inLabel := fmt.Sprintf("[%d:v]", i)
		scaledLabel := fmt.Sprintf("[s%d]", i)

		// Fit mode: cover vs contain
		if spec.Output.Fit == "contain" {
			blurRadius := int(math.Max(10, float64(outW)/25.0))
			// Contain blur background + sharp foreground
			bgFilter := fmt.Sprintf("%sscale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,boxblur=%d[bg%d]",
				inLabel, outW, outH, outW, outH, blurRadius, i)
			fgFilter := fmt.Sprintf("%sscale=%d:%d:force_original_aspect_ratio=decrease[fg%d]",
				inLabel, outW, outH, i)
			overlayFilter := fmt.Sprintf("[bg%d][fg%d]overlay=(W-w)/2:(H-h)/2%s",
				i, i, scaledLabel)
			filterComplex = append(filterComplex, bgFilter, fgFilter, overlayFilter)
		} else {
			// Cover mode (default)
			coverFilter := fmt.Sprintf("%sscale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d%s",
				inLabel, outW, outH, outW, outH, scaledLabel)
			filterComplex = append(filterComplex, coverFilter)
		}

		// Color filter & Adjustments
		adj := ResolveFilter(clip.Filter, clip.Adjust)
		eqLabel := fmt.Sprintf("[eq%d]", i)
		eqFilter := fmt.Sprintf("%seq=brightness=%.3f:contrast=%.3f:saturation=%.3f,fps=%d%s",
			scaledLabel, adj.Brightness, adj.Contrast, adj.Saturation, fps, eqLabel)
		filterComplex = append(filterComplex, eqFilter)

		currentLabel := eqLabel

		// Text Overlays
		if len(clip.Texts) > 0 {
			for tIdx, txt := range clip.Texts {
				nextTextLabel := fmt.Sprintf("[txt_%d_%d]", i, tIdx)
				fontName := FontFileName[txt.Font]
				if fontName == "" {
					fontName = "Inter-Regular.ttf"
				}
				fontFile := filepath.Join(cfg.FontDir, fontName)
				// Normalised font size relative to 1080p
				scaleFactor := float64(outW) / 1080.0
				pxSize := int(math.Round(float64(txt.Size) * scaleFactor))
				if pxSize < 14 {
					pxSize = 14
				}

				posX := int(math.Round(txt.X * float64(outW)))
				posY := int(math.Round(txt.Y * float64(outH)))

				// Escape single quotes for drawtext
				cleanText := strings.ReplaceAll(txt.Text, "'", "\\'")
				cleanText = strings.ReplaceAll(cleanText, ":", "\\:")

				endSec := txt.End
				if endSec <= 0 || endSec > clip.Duration {
					endSec = clip.Duration
				}

				enableExpr := fmt.Sprintf("between(t,%.2f,%.2f)", txt.Start, endSec)

				// Color format: 0xRRGGBB or named
				col := strings.TrimPrefix(txt.Color, "#")
				if len(col) != 6 {
					col = "ffffff"
				}

				drawTextFilter := fmt.Sprintf(
					"%sdrawtext=fontfile='%s':text='%s':fontsize=%d:fontcolor=0x%s:x=%d:y=%d:enable='%s'%s",
					currentLabel, fontFile, cleanText, pxSize, col, posX, posY, enableExpr, nextTextLabel,
				)
				filterComplex = append(filterComplex, drawTextFilter)
				currentLabel = nextTextLabel
			}
		}

		// Final ready label for clip i
		readyLabel := fmt.Sprintf("[v%d]", i)
		filterComplex = append(filterComplex, fmt.Sprintf("%snull%s", currentLabel, readyLabel))
	}

	// 4. Clip sequencing with xfade transitions
	finalVideoLabel := "[v0]"
	if len(spec.Clips) > 1 {
		lastLabel := "[v0]"
		for i := 0; i < len(spec.Clips)-1; i++ {
			c := spec.Clips[i]
			nextClipLabel := fmt.Sprintf("[v%d]", i+1)
			outXfadeLabel := fmt.Sprintf("[xf%d]", i)

			trans := c.TransitionOut
			if trans != nil && trans.Duration > 0 {
				transType := "fade"
				transDuration := trans.Duration
				switch trans.Type {
				case "slide-left":
					transType = "slideleft"
				case "slide-right":
					transType = "slideright"
				case "dissolve":
					transType = "dissolve"
				case "zoom-in":
					transType = "zoomin"
				case "zoom-out":
					transType = "fade"
				default:
					transType = "fade"
				}

				offset := XFadeOffset(spec.Clips, i)
				xfadeFilter := fmt.Sprintf(
					"%s%sxfade=transition=%s:duration=%.2f:offset=%.2f%s",
					lastLabel, nextClipLabel, transType, transDuration, offset, outXfadeLabel,
				)
				filterComplex = append(filterComplex, xfadeFilter)
				lastLabel = outXfadeLabel
			} else {
				// Direct cut: minimal 0.05s transition at the end of clip i
				offset := math.Max(0, ClipStartTime(spec.Clips, i)+c.Duration-0.05)
				xfadeFilter := fmt.Sprintf(
					"%s%sxfade=transition=fade:duration=0.05:offset=%.2f%s",
					lastLabel, nextClipLabel, offset, outXfadeLabel,
				)
				filterComplex = append(filterComplex, xfadeFilter)
				lastLabel = outXfadeLabel
			}
		}
		finalVideoLabel = lastLabel
	}

	// 5. Audio filter chain
	finalAudioLabel := ""
	totalVidDuration := TotalDuration(spec.Clips)
	if hasAudio {
		finalAudioLabel = "[aout]"
		vol := spec.Audio.Volume
		if vol <= 0 {
			vol = 1.0
		}
		fadeIn := spec.Audio.FadeIn
		fadeOut := spec.Audio.FadeOut
		fadeEndOffset := math.Max(0, totalVidDuration-fadeOut)

		audioFilters := fmt.Sprintf("[%d:a]volume=%.2f", audioInputIndex, vol)
		if fadeIn > 0 {
			audioFilters += fmt.Sprintf(",afade=t=in:ss=0:d=%.2f", fadeIn)
		}
		if fadeOut > 0 && fadeEndOffset > 0 {
			audioFilters += fmt.Sprintf(",afade=t=out:st=%.2f:d=%.2f", fadeEndOffset, fadeOut)
		}
		audioFilters += fmt.Sprintf(",atrim=0:%.2f%s", totalVidDuration, finalAudioLabel)
		filterComplex = append(filterComplex, audioFilters)
	}

	// Join filter complex
	args = append(args, "-filter_complex", strings.Join(filterComplex, ";"))

	// Map final outputs
	args = append(args, "-map", finalVideoLabel)
	if finalAudioLabel != "" {
		args = append(args, "-map", finalAudioLabel)
	}

	// Encoding parameters
	args = append(args,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "fast",
		"-threads", "0",
		"-crf", "22",
		"-t", fmt.Sprintf("%.4f", totalVidDuration),
		"-movflags", "+faststart",
	)

	if finalAudioLabel != "" {
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	}

	args = append(args, cfg.OutputPath)

	return args, nil
}
