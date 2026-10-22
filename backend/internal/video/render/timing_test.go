package render

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// ---- vector loader ---------------------------------------------------------

type timingVectors struct {
	TotalDuration      []tdVector      `json:"total_duration"`
	ClipStartTimes     []startVector   `json:"clip_start_times"`
	XFadeOffsets       []xfadeVector   `json:"xfade_offsets"`
	TransitionMaxAllow []maxVector     `json:"transition_max_allowed"`
	KenBurnsEasing     []easingVector  `json:"ken_burns_easing"`
	TextAlpha          []alphaVector   `json:"text_alpha"`
}

// minimal wire types for the JSON vectors (avoid importing extra deps)
type wireClip struct {
	Duration      float64          `json:"duration"`
	TransitionOut *wireTransition  `json:"transition_out"`
}
type wireTransition struct {
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

func wireToClips(wcs []wireClip) []ClipSpec {
	out := make([]ClipSpec, len(wcs))
	for i, wc := range wcs {
		out[i] = ClipSpec{Duration: wc.Duration}
		if wc.TransitionOut != nil {
			out[i].TransitionOut = &TransitionSpec{
				Type:     wc.TransitionOut.Type,
				Duration: wc.TransitionOut.Duration,
			}
		}
	}
	return out
}

type tdVector struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Clips       []wireClip `json:"clips"`
	ClipsRepeat *struct {
		Count        int     `json:"count"`
		Duration     float64 `json:"duration"`
		TransitionOut *wireTransition `json:"transition_out"`
	} `json:"clips_repeat"`
	Expected float64 `json:"expected"`
}

type startVector struct {
	ID             string     `json:"id"`
	Clips          []wireClip `json:"clips"`
	ExpectedStarts []float64  `json:"expected_starts"`
}

type xfadeVector struct {
	ID              string     `json:"id"`
	Clips           []wireClip `json:"clips"`
	ExpectedOffsets []float64  `json:"expected_offsets"`
}

type maxVector struct {
	ID           string     `json:"id"`
	Clips        []wireClip `json:"clips"`
	BetweenIndex int        `json:"between_index"`
	ExpectedMax  float64    `json:"expected_max"`
}

type easingVector struct {
	ID       string  `json:"id"`
	Easing   string  `json:"easing"`
	From     float64 `json:"from"`
	To       float64 `json:"to"`
	U        float64 `json:"u"`
	Expected float64 `json:"expected"`
}

type alphaTextWire struct {
	Start        float64 `json:"start"`
	End          float64 `json:"end"`
	Anim         string  `json:"anim"`
	AnimDuration float64 `json:"anim_duration"`
}

type alphaVector struct {
	ID            string        `json:"id"`
	ClipStart     float64       `json:"clip_start"`
	Text          alphaTextWire `json:"text"`
	T             float64       `json:"t"`
	ExpectedAlpha float64       `json:"expected_alpha"`
}

func loadVectors(t *testing.T) timingVectors {
	t.Helper()
	data, err := os.ReadFile("testdata/timing_vectors.json")
	if err != nil {
		t.Fatalf("load timing_vectors.json: %v", err)
	}
	var v timingVectors
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("parse timing_vectors.json: %v", err)
	}
	return v
}

const eps = 1e-6 // tolerance for float comparisons

func approxEq(a, b float64) bool {
	return math.Abs(a-b) < eps
}

// ---- tests -----------------------------------------------------------------

func TestTotalDuration_GoldenVectors(t *testing.T) {
	vecs := loadVectors(t)
	for _, v := range vecs.TotalDuration {
		t.Run(v.ID, func(t *testing.T) {
			var clips []ClipSpec
			if v.ClipsRepeat != nil {
				for i := 0; i < v.ClipsRepeat.Count; i++ {
					c := ClipSpec{Duration: v.ClipsRepeat.Duration}
					if v.ClipsRepeat.TransitionOut != nil {
						c.TransitionOut = &TransitionSpec{
							Type:     v.ClipsRepeat.TransitionOut.Type,
							Duration: v.ClipsRepeat.TransitionOut.Duration,
						}
					}
					clips = append(clips, c)
				}
			} else {
				clips = wireToClips(v.Clips)
			}
			got := TotalDuration(clips)
			if !approxEq(got, v.Expected) {
				t.Errorf("TotalDuration = %.6f, want %.6f", got, v.Expected)
			}
		})
	}
}

func TestClipStartTimes_GoldenVectors(t *testing.T) {
	vecs := loadVectors(t)
	for _, v := range vecs.ClipStartTimes {
		t.Run(v.ID, func(t *testing.T) {
			clips := wireToClips(v.Clips)
			for i, want := range v.ExpectedStarts {
				got := ClipStartTime(clips, i)
				if !approxEq(got, want) {
					t.Errorf("ClipStartTime[%d] = %.6f, want %.6f", i, got, want)
				}
			}
		})
	}
}

func TestXFadeOffsets_GoldenVectors(t *testing.T) {
	vecs := loadVectors(t)
	for _, v := range vecs.XFadeOffsets {
		t.Run(v.ID, func(t *testing.T) {
			clips := wireToClips(v.Clips)
			offIdx := 0
			for i, c := range clips {
				if c.TransitionOut == nil {
					continue
				}
				want := v.ExpectedOffsets[offIdx]
				got := XFadeOffset(clips, i)
				if !approxEq(got, want) {
					t.Errorf("XFadeOffset[%d] = %.6f, want %.6f", i, got, want)
				}
				offIdx++
			}
		})
	}
}

func TestTransitionMaxAllowed_GoldenVectors(t *testing.T) {
	vecs := loadVectors(t)
	for _, v := range vecs.TransitionMaxAllow {
		t.Run(v.ID, func(t *testing.T) {
			clips := wireToClips(v.Clips)
			_, maxAllowed := ValidateTransitionDuration(clips, v.BetweenIndex, 0)
			if !approxEq(maxAllowed, v.ExpectedMax) {
				t.Errorf("maxAllowed = %.6f, want %.6f", maxAllowed, v.ExpectedMax)
			}
		})
	}
}

func TestKenBurnsEasing_GoldenVectors(t *testing.T) {
	vecs := loadVectors(t)
	for _, v := range vecs.KenBurnsEasing {
		t.Run(v.ID, func(t *testing.T) {
			got := KenBurnsValue(v.From, v.To, v.U, v.Easing)
			if !approxEq(got, v.Expected) {
				t.Errorf("KenBurnsValue(from=%.2f, to=%.2f, u=%.4f, easing=%q) = %.8f, want %.8f",
					v.From, v.To, v.U, v.Easing, got, v.Expected)
			}
		})
	}
}

func TestTextAlpha_GoldenVectors(t *testing.T) {
	vecs := loadVectors(t)
	for _, v := range vecs.TextAlpha {
		t.Run(v.ID, func(t *testing.T) {
			spec := TextSpec{
				Start:        v.Text.Start,
				End:          v.Text.End,
				Anim:         v.Text.Anim,
				AnimDuration: v.Text.AnimDuration,
			}
			got := TextAlpha(v.ClipStart, spec, v.T)
			if !approxEq(got, v.ExpectedAlpha) {
				t.Errorf("TextAlpha(clipStart=%.2f, t=%.2f) = %.6f, want %.6f",
					v.ClipStart, v.T, got, v.ExpectedAlpha)
			}
		})
	}
}

// ---- extra unit tests (not in vectors, logic-only) -------------------------

func TestClipLocalTime_ActiveWindow(t *testing.T) {
	clips := []ClipSpec{
		{Duration: 3.0, TransitionOut: &TransitionSpec{Duration: 0.5}},
		{Duration: 4.0},
	}
	// clip[1] starts at 3.0−0.5 = 2.5
	// within [2.5, 6.5)
	u := ClipLocalTime(clips, 1, 4.5) // 2s into clip[1] of 4s → u=0.5
	if !approxEq(u, 0.5) {
		t.Errorf("ClipLocalTime = %.6f, want 0.5", u)
	}
}

func TestClipLocalTime_OutsideWindow(t *testing.T) {
	clips := []ClipSpec{{Duration: 3.0}}
	if u := ClipLocalTime(clips, 0, 5.0); u != -1 {
		t.Errorf("expected -1 for t outside window, got %.6f", u)
	}
}

func TestTotalDuration_NoClips(t *testing.T) {
	if d := TotalDuration(nil); d != 0 {
		t.Errorf("expected 0 for empty clips, got %.6f", d)
	}
}
