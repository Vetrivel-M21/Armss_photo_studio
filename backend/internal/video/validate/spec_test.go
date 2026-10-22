package validate

import (
	"fmt"
	"strings"
	"testing"

	"armss-poster-studio/internal/video/render"
)

var lim = DefaultLimits()
const userPrefix = "video/user-abc-123/"

// ---- helpers ---------------------------------------------------------------

func validSpec() render.VideoSpec {
	return render.VideoSpec{
		Version: render.SpecVersion,
		Output: render.OutputSpec{Aspect: "9:16", Fit: "cover", FPS: 30},
		Clips: []render.ClipSpec{
			{
				ID:       "clip-1",
				ImageKey: userPrefix + "images/a.jpg",
				Duration: 3.0,
				Crop:     render.CropSpec{X: 0, Y: 0, W: 1, H: 1},
				Filter:   "none",
				Adjust:   render.AdjustSpec{Brightness: 0, Contrast: 1, Saturation: 1},
			},
		},
	}
}

func expectValid(t *testing.T, spec render.VideoSpec) {
	t.Helper()
	if err := Spec(spec, lim, userPrefix); err != nil {
		t.Errorf("expected valid spec, got error: %v", err)
	}
}

func expectField(t *testing.T, spec render.VideoSpec, contains string) {
	t.Helper()
	err := Spec(spec, lim, userPrefix)
	if err == nil {
		t.Errorf("expected validation error containing %q, got nil", contains)
		return
	}
	if !strings.Contains(err.Error(), contains) {
		t.Errorf("expected error to contain %q, got: %v", contains, err)
	}
}

// ---- valid baseline --------------------------------------------------------

func TestSpec_Valid_Baseline(t *testing.T) {
	expectValid(t, validSpec())
}

// ---- version ---------------------------------------------------------------

func TestSpec_WrongVersion(t *testing.T) {
	s := validSpec()
	s.Version = 99
	expectField(t, s, "unsupported spec version")
}

// ---- output ----------------------------------------------------------------

func TestSpec_BadAspect(t *testing.T) {
	s := validSpec(); s.Output.Aspect = "4:3"
	expectField(t, s, "output.aspect")
}

func TestSpec_BadFit(t *testing.T) {
	s := validSpec(); s.Output.Fit = "stretch"
	expectField(t, s, "output.fit")
}

func TestSpec_BadFPS(t *testing.T) {
	s := validSpec(); s.Output.FPS = 60
	expectField(t, s, "output.fps")
}

// ---- clips -----------------------------------------------------------------

func TestSpec_NoClips(t *testing.T) {
	s := validSpec(); s.Clips = nil
	expectField(t, s, "at least one clip")
}

func TestSpec_TooManyClips(t *testing.T) {
	s := validSpec()
	for i := 0; i < lim.MaxClips; i++ {
		c := s.Clips[0]
		c.ID = fmt.Sprintf("clip-extra-%d", i)
		s.Clips = append(s.Clips, c)
	}
	expectField(t, s, fmt.Sprintf("max %d clips", lim.MaxClips))
}

func TestSpec_DuplicateClipID(t *testing.T) {
	s := validSpec()
	s.Clips = append(s.Clips, s.Clips[0]) // same ID
	expectField(t, s, "duplicate clip id")
}

func TestSpec_ForeignImageKey(t *testing.T) {
	s := validSpec()
	s.Clips[0].ImageKey = "video/user-other/images/hack.jpg"
	expectField(t, s, "does not belong")
}

func TestSpec_PresetKeyAllowed(t *testing.T) {
	// Preset keys are always accessible regardless of user
	s := validSpec()
	s.Clips[0].ImageKey = "video/presets/sample.jpg"
	expectValid(t, s)
}

func TestSpec_ClipDurationTooShort(t *testing.T) {
	s := validSpec(); s.Clips[0].Duration = 0.5
	expectField(t, s, "duration")
}

func TestSpec_ClipDurationTooLong(t *testing.T) {
	s := validSpec(); s.Clips[0].Duration = lim.MaxClipDurationS + 1
	expectField(t, s, "duration")
}

func TestSpec_CropOutOfBounds(t *testing.T) {
	s := validSpec(); s.Clips[0].Crop = render.CropSpec{X: 0.8, Y: 0, W: 0.5, H: 1}
	expectField(t, s, "crop")
}

func TestSpec_InvalidRotation(t *testing.T) {
	s := validSpec(); s.Clips[0].Rotate = 45
	expectField(t, s, "rotate")
}

func TestSpec_BadFilter(t *testing.T) {
	s := validSpec(); s.Clips[0].Filter = "hdr-boost"
	expectField(t, s, "filter")
}

func TestSpec_AdjustBrightnessOutOfRange(t *testing.T) {
	s := validSpec(); s.Clips[0].Adjust.Brightness = 1.5
	expectField(t, s, "brightness")
}

func TestSpec_AdjustContrastOutOfRange(t *testing.T) {
	s := validSpec(); s.Clips[0].Adjust.Contrast = 0.1
	expectField(t, s, "contrast")
}

// ---- transition ------------------------------------------------------------

func TestSpec_LastClipHasTransition(t *testing.T) {
	s := validSpec()
	s.Clips[0].TransitionOut = &render.TransitionSpec{Type: "fade", Duration: 0.5}
	// Only one clip → it's the last → invalid
	expectField(t, s, "last clip must not have a transition")
}

func TestSpec_TransitionTooLong(t *testing.T) {
	s := validSpec()
	s.Clips = append(s.Clips, render.ClipSpec{
		ID: "clip-2", ImageKey: userPrefix + "images/b.jpg",
		Duration: 2.0, Filter: "none",
		Adjust: render.AdjustSpec{Contrast: 1, Saturation: 1},
		Crop: render.CropSpec{W: 1, H: 1},
	})
	// 0.5×min(3,2) = 1.0; set 1.1 → should fail
	s.Clips[0].TransitionOut = &render.TransitionSpec{Type: "fade", Duration: 1.1}
	expectField(t, s, "exceeds allowed max")
}

func TestSpec_TransitionBadType(t *testing.T) {
	s := validSpec()
	s.Clips = append(s.Clips, render.ClipSpec{
		ID: "clip-2", ImageKey: userPrefix + "images/b.jpg",
		Duration: 3.0, Filter: "none",
		Adjust: render.AdjustSpec{Contrast: 1, Saturation: 1},
		Crop: render.CropSpec{W: 1, H: 1},
	})
	s.Clips[0].TransitionOut = &render.TransitionSpec{Type: "wipe", Duration: 0.5}
	expectField(t, s, "unknown transition")
}

// ---- ken burns -------------------------------------------------------------

func TestSpec_KenBurnsScaleOutOfRange(t *testing.T) {
	s := validSpec()
	s.Clips[0].KenBurns = &render.KenBurnsSpec{
		From:   render.KenBurnsKeyframe{Scale: 2.5, X: 0.5, Y: 0.5},
		To:     render.KenBurnsKeyframe{Scale: 1.0, X: 0.5, Y: 0.5},
		Easing: "linear",
	}
	expectField(t, s, "ken_burns")
}

// ---- text overlays ---------------------------------------------------------

func TestSpec_TextBadFont(t *testing.T) {
	s := validSpec()
	s.Clips[0].Texts = []render.TextSpec{{
		Text: "Hello", Font: "Comic Sans", Size: 32, Color: "#ffffff",
		Opacity: 1, X: 0.5, Y: 0.8, Start: 0, End: 2,
		Anim: "fade-in", AnimDuration: 0.3,
	}}
	expectField(t, s, "unknown font")
}

func TestSpec_TextTooLong(t *testing.T) {
	s := validSpec()
	s.Clips[0].Texts = []render.TextSpec{{
		Text: strings.Repeat("a", lim.MaxTextLength+1),
		Font: "Inter", Size: 32, Color: "#ffffff",
		Opacity: 1, X: 0.5, Y: 0.8, Start: 0, End: 2,
		Anim: "none", AnimDuration: 0.3,
	}}
	expectField(t, s, "max 200 chars")
}

func TestSpec_TextBadColor(t *testing.T) {
	s := validSpec()
	s.Clips[0].Texts = []render.TextSpec{{
		Text: "Hello", Font: "Inter", Size: 32, Color: "red",
		Opacity: 1, X: 0.5, Y: 0.8, Start: 0, End: 2,
		Anim: "none", AnimDuration: 0.3,
	}}
	expectField(t, s, "hex")
}

func TestSpec_TextEndBeforeStart(t *testing.T) {
	s := validSpec()
	s.Clips[0].Texts = []render.TextSpec{{
		Text: "Hi", Font: "Inter", Size: 32, Color: "#ffffff",
		Opacity: 1, X: 0.5, Y: 0.8, Start: 1.5, End: 1.0,
		Anim: "none", AnimDuration: 0.3,
	}}
	expectField(t, s, "end: must be > start")
}

// ---- audio -----------------------------------------------------------------

func TestSpec_AudioForeignKey(t *testing.T) {
	s := validSpec()
	s.Audio = &render.AudioSpec{
		Key: "video/other-user/audio/track.mp3",
		Volume: 1.0,
	}
	expectField(t, s, "audio.key")
}

func TestSpec_AudioVolumeOutOfRange(t *testing.T) {
	s := validSpec()
	s.Audio = &render.AudioSpec{
		Key:    userPrefix + "audio/track.mp3",
		Volume: 3.0,
	}
	expectField(t, s, "audio.volume")
}

func TestSpec_AudioTrimEndBeforeStart(t *testing.T) {
	s := validSpec()
	s.Audio = &render.AudioSpec{
		Key: userPrefix + "audio/track.mp3",
		TrimStart: 5.0, TrimEnd: 3.0, Volume: 1.0,
	}
	expectField(t, s, "trim_end")
}

// ---- total duration --------------------------------------------------------

func TestSpec_TotalDurationExceedsMax(t *testing.T) {
	s := validSpec()
	s.Clips[0].Duration = lim.MaxClipDurationS
	s.Clips = append(s.Clips, render.ClipSpec{
		ID:       "clip-extra-x",
		ImageKey: userPrefix + "images/x.jpg",
		Duration: 10,
		Filter:   "none",
		Adjust:   render.AdjustSpec{Contrast: 1, Saturation: 1},
		Crop:     render.CropSpec{W: 1, H: 1},
	})
	// Total duration = 1800 + 10 = 1810s > 1800s
	expectField(t, s, "exceeds max")
}

// ---- resolution ------------------------------------------------------------

func TestResolution_Valid(t *testing.T) {
	for _, r := range []string{"720p", "1080p"} {
		if err := Resolution(r); err != nil {
			t.Errorf("expected valid resolution %q, got: %v", r, err)
		}
	}
}

func TestResolution_Invalid(t *testing.T) {
	if err := Resolution("4K"); err == nil {
		t.Error("expected error for 4K resolution")
	}
}
