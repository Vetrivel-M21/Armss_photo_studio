package render

import (
	"strings"
	"testing"
)

func TestBuildFFmpegArgs_Basic(t *testing.T) {
	spec := DefaultSpec()
	spec.Clips = []ClipSpec{
		{
			ID:       "c1",
			Duration: 3.0,
			Filter:   "none",
			Crop:     CropSpec{W: 1, H: 1},
			TransitionOut: &TransitionSpec{
				Type:     "fade",
				Duration: 0.8,
			},
		},
		{
			ID:       "c2",
			Duration: 4.0,
			Filter:   "vivid",
			Crop:     CropSpec{W: 1, H: 1},
		},
	}

	cfg := FFmpegJobConfig{
		Spec:             spec,
		ResolutionPreset: "1080p",
		ImagePaths:       []string{"/tmp/img1.jpg", "/tmp/img2.jpg"},
		FontDir:          "/fonts",
		OutputPath:       "/tmp/out.mp4",
	}

	args, err := BuildFFmpegArgs(cfg)
	if err != nil {
		t.Fatalf("BuildFFmpegArgs failed: %v", err)
	}

	cmdStr := strings.Join(args, " ")
	if !strings.Contains(cmdStr, "-i /tmp/img1.jpg") {
		t.Errorf("Expected image 1 in command, got: %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "-i /tmp/img2.jpg") {
		t.Errorf("Expected image 2 in command, got: %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "xfade=transition=fade") {
		t.Errorf("Expected xfade in command, got: %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "/tmp/out.mp4") {
		t.Errorf("Expected output path in command, got: %s", cmdStr)
	}
}

func TestBuildFFmpegArgs_WithAudioAndText(t *testing.T) {
	spec := DefaultSpec()
	spec.Clips = []ClipSpec{
		{
			ID:       "c1",
			Duration: 5.0,
			Filter:   "warm",
			Texts: []TextSpec{
				{
					Text:  "Hello World",
					Font:  "Inter",
					Size:  48,
					Color: "#FF0000",
					X:     0.1,
					Y:     0.8,
					Start: 1.0,
					End:   4.0,
				},
			},
		},
	}
	spec.Audio = &AudioSpec{
		Key:       "audio/music.mp3",
		Volume:    0.8,
		FadeIn:    1.0,
		FadeOut:   1.5,
		TrimStart: 2.0,
	}

	cfg := FFmpegJobConfig{
		Spec:             spec,
		ResolutionPreset: "720p",
		ImagePaths:       []string{"/tmp/img1.jpg"},
		AudioPath:        "/tmp/music.mp3",
		FontDir:          "/fonts",
		OutputPath:       "/tmp/out.mp4",
	}

	args, err := BuildFFmpegArgs(cfg)
	if err != nil {
		t.Fatalf("BuildFFmpegArgs failed: %v", err)
	}

	cmdStr := strings.Join(args, " ")
	if !strings.Contains(cmdStr, "drawtext=") {
		t.Errorf("Expected drawtext in command, got: %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "afade=t=in") {
		t.Errorf("Expected audio fade-in, got: %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "afade=t=out") {
		t.Errorf("Expected audio fade-out, got: %s", cmdStr)
	}
}
