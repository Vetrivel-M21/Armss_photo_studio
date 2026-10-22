package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/video/render"
	"armss-poster-studio/internal/video/store"
)

type Runner struct {
	Store      *store.Store
	Storage    *storage.Store
	FontDir    string
	FFmpegPath string
	Logger     *slog.Logger
}

func NewRunner(st *store.Store, storage *storage.Store, fontDir, ffmpegPath string, logger *slog.Logger) *Runner {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{
		Store:      st,
		Storage:    storage,
		FontDir:    fontDir,
		FFmpegPath: ffmpegPath,
		Logger:     logger,
	}
}

// EnqueueVideoJob starts an asynchronous worker goroutine to process the job.
func (r *Runner) EnqueueVideoJob(jobID, userID string) error {
	go func() {
		// Support up to 30-minute videos: allow 90 minutes for full FFmpeg encoding
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
		defer cancel()

		if err := r.processJob(ctx, jobID, userID); err != nil {
			r.Logger.Error("video_job_failed", "job_id", jobID, "user_id", userID, "error", err)
			_ = r.Store.UpdateJobStatus(context.Background(), jobID, "failed", 0, "", err.Error())
		}
	}()
	return nil
}

func (r *Runner) processJob(ctx context.Context, jobID, userID string) error {
	// 1. Fetch job spec
	job, err := r.Store.GetJob(ctx, jobID, userID)
	if err != nil {
		return fmt.Errorf("load job: %w", err)
	}
	if job.Spec == nil {
		return fmt.Errorf("job has empty spec")
	}

	// 2. Mark processing
	if err := r.Store.UpdateJobStatus(ctx, jobID, "processing", 10, "", ""); err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	// 3. Create temporary working directory for input files and rendered output
	tmpDir, err := os.MkdirTemp("", "video_job_"+jobID+"_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 4. Download / resolve images
	var imagePaths []string
	for i, clip := range job.Spec.Clips {
		imgPath := filepath.Join(tmpDir, fmt.Sprintf("clip_%d.jpg", i))
		if err := r.resolveMediaFile(ctx, clip.ImageKey, imgPath); err != nil {
			return fmt.Errorf("resolve image %d (%s): %w", i, clip.ImageKey, err)
		}
		imagePaths = append(imagePaths, imgPath)
	}

	_ = r.Store.UpdateJobStatus(ctx, jobID, "processing", 35, "", "")

	// 5. Download audio if present
	var audioPath string
	if job.Spec.Audio != nil && job.Spec.Audio.Key != "" {
		audioPath = filepath.Join(tmpDir, "audio_track.mp3")
		if err := r.resolveMediaFile(ctx, job.Spec.Audio.Key, audioPath); err != nil {
			return fmt.Errorf("resolve audio (%s): %w", job.Spec.Audio.Key, err)
		}
	}

	_ = r.Store.UpdateJobStatus(ctx, jobID, "processing", 50, "", "")

	// 6. Build FFmpeg command
	outMp4Path := filepath.Join(tmpDir, "output.mp4")
	cfg := render.FFmpegJobConfig{
		Spec:             *job.Spec,
		ResolutionPreset: job.Resolution,
		ImagePaths:       imagePaths,
		AudioPath:        audioPath,
		FontDir:          r.FontDir,
		OutputPath:       outMp4Path,
	}

	args, err := render.BuildFFmpegArgs(cfg)
	if err != nil {
		return fmt.Errorf("build ffmpeg args: %w", err)
	}

	_ = r.Store.UpdateJobStatus(ctx, jobID, "processing", 65, "", "")

	// 7. Execute FFmpeg
	cmd := exec.CommandContext(ctx, r.FFmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg execution failed (%w): %s", err, string(output))
	}

	_ = r.Store.UpdateJobStatus(ctx, jobID, "processing", 85, "", "")

	// 8. Upload final MP4 to storage
	mp4Bytes, err := os.ReadFile(outMp4Path)
	if err != nil {
		return fmt.Errorf("read rendered mp4: %w", err)
	}

	outputKey := fmt.Sprintf("video/%s/renders/%s.mp4", userID, jobID)
	if err := r.Storage.Put(ctx, outputKey, mp4Bytes, "video/mp4"); err != nil {
		return fmt.Errorf("store output mp4: %w", err)
	}

	// 9. Mark job completed
	if err := r.Store.UpdateJobStatus(ctx, jobID, "done", 100, outputKey, ""); err != nil {
		return fmt.Errorf("mark job done: %w", err)
	}

	r.Logger.Info("video_job_completed", "job_id", jobID, "user_id", userID, "output_key", outputKey)
	return nil
}

func (r *Runner) resolveMediaFile(ctx context.Context, keyOrURL, destPath string) error {
	// If it's a web URL
	if strings.HasPrefix(keyOrURL, "http://") || strings.HasPrefix(keyOrURL, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, keyOrURL, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http download failed: status %d", resp.StatusCode)
		}

		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
		return err
	}

	// Object storage key
	cleanKey := strings.TrimPrefix(keyOrURL, "/files/")
	data, err := r.Storage.Get(ctx, cleanKey)
	if err != nil {
		return fmt.Errorf("get object %q: %w", cleanKey, err)
	}
	return os.WriteFile(destPath, data, 0644)
}
