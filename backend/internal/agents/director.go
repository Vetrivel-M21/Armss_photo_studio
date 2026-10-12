package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const directorSystemPrompt = `You are the Creative Director for ARMSS Poster Studio. Given a
creative brief and (optionally) a house style profile, produce exactly three genuinely
different big ideas for the poster — different concepts, not three colourways of the same
idea. Each concept needs its own visual metaphor and mood, while still fitting the brief's
occasion and the house style's tone.

FESTIVE & CULTURAL VIBRANCY RULE:
For festival and celebration occasions (e.g., Diwali, Pongal, Eid, New Year), every concept must celebrate genuine human warmth, festive joy, and cultural vitality. Incorporate tangible human presence (e.g. families sharing moments, traditional silk attire, lighting lamps, sparklers), authentic festive elements, and celebratory lighting — never strip away the festival spirit into sterile, lifeless tabletop still-lifes.

CRITICAL: You must invoke the "emit_concepts" tool where the "concepts" parameter is a raw JSON
array of exactly three distinct concept objects (NOT a string, NOT markdown, NOT stringified JSON):
[
  {
    "name": "Concept Title",
    "bigIdea": "Core creative thesis",
    "visualMetaphor": "Visual subject and depth",
    "mood": "Emotional tone",
    "palette": "Color palette"
  },
  ...
]`

func directorToolSchema() map[string]any {
	concept := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":           map[string]any{"type": "string", "description": "Short distinctive title for the concept"},
			"bigIdea":        map[string]any{"type": "string", "description": "Core creative thesis and narrative"},
			"visualMetaphor": map[string]any{"type": "string", "description": "Specific visual elements, subjects, and framing"},
			"mood":           map[string]any{"type": "string", "description": "Atmospheric and emotional feel"},
			"palette":        map[string]any{"type": "string", "description": "Key color tones and mood"},
		},
		"required": []string{"name", "bigIdea", "visualMetaphor", "mood", "palette"},
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"concepts": map[string]any{
				"type":        "array",
				"description": "Array of exactly three distinct concept objects. Do not stringify.",
				"items":       concept,
				"minItems":    3,
				"maxItems":    3,
			},
		},
		"required": []string{"concepts"},
	}
}

type conceptsInput struct {
	Concepts []Concept `json:"concepts"`
}

// RunCreativeDirector is agent 3: Brief + StyleProfile → exactly three Concepts.
func (c *Client) RunCreativeDirector(ctx context.Context, brief Brief, style *StyleProfile) ([]Concept, Usage, error) {
	briefJSON, _ := json.MarshalIndent(brief, "", "  ")
	prompt := fmt.Sprintf("Creative brief:\n%s\n\n%s", string(briefJSON), styleContext(style))

	result, err := c.CallTool(ctx, ToolCallRequest{
		System:   directorSystemPrompt,
		UserText: prompt,
		Tool: ToolSpec{
			Name:        "emit_concepts",
			Description: "Emit exactly three distinct creative concepts for this poster.",
			InputSchema: directorToolSchema(),
		},
		MaxTokens: MaxTokensDirector,
	})
	if err != nil {
		return nil, Usage{}, err
	}

	concepts := parseConcepts(result.Raw, brief, style)
	return concepts, Usage{TokensIn: result.TokensIn, TokensOut: result.TokensOut}, nil
}

// parseConcepts extracts Concepts from any output format returned by LLM function calls:
// handles direct struct matching, root JSON arrays, stringified JSON strings, markdown fences,
// or structured text descriptions, ensuring the poster pipeline never crashes.
func parseConcepts(raw []byte, brief Brief, style *StyleProfile) []Concept {
	cleaned := cleanRawJSON(raw)
	if len(cleaned) == 0 {
		return synthesizeFallbackConcepts(brief, style)
	}

	// 1. Direct match with conceptsInput: {"concepts": [...]}
	var directInput conceptsInput
	if err := json.Unmarshal(cleaned, &directInput); err == nil && len(directInput.Concepts) > 0 {
		return finalizeConcepts(directInput.Concepts, brief, style)
	}

	// 2. Direct array: [{...}, ...]
	var rootList []Concept
	if err := json.Unmarshal(cleaned, &rootList); err == nil && len(rootList) > 0 {
		return finalizeConcepts(rootList, brief, style)
	}

	// 3. Map lookup for concepts / items / result
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(cleaned, &obj); err == nil {
		for _, k := range []string{"concepts", "Concepts", "items", "data", "result", "options", "conceptsList"} {
			if val, ok := obj[k]; ok && len(val) > 0 {
				// Try direct array inside val
				var subList []Concept
				if err := json.Unmarshal(val, &subList); err == nil && len(subList) > 0 {
					return finalizeConcepts(subList, brief, style)
				}
				// Try string (stringified JSON or text)
				var strVal string
				if err := json.Unmarshal(val, &strVal); err == nil {
					cleanStr := cleanMarkdownString(strVal)
					var strList []Concept
					if err := json.Unmarshal([]byte(cleanStr), &strList); err == nil && len(strList) > 0 {
						return finalizeConcepts(strList, brief, style)
					}
					var strWrapped conceptsInput
					if err := json.Unmarshal([]byte(cleanStr), &strWrapped); err == nil && len(strWrapped.Concepts) > 0 {
						return finalizeConcepts(strWrapped.Concepts, brief, style)
					}
					fromText := extractConceptsFromText(cleanStr)
					if len(fromText) > 0 {
						return finalizeConcepts(fromText, brief, style)
					}
				}
			}
		}
	}

	// 4. Try extract concepts from raw string/text
	fromText := extractConceptsFromText(string(cleaned))
	if len(fromText) > 0 {
		return finalizeConcepts(fromText, brief, style)
	}

	// 5. Fail-safe fallback to synthesized concepts
	return synthesizeFallbackConcepts(brief, style)
}

func cleanRawJSON(raw []byte) []byte {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
		var unescaped string
		if err := json.Unmarshal(trimmed, &unescaped); err == nil {
			trimmed = bytes.TrimSpace([]byte(unescaped))
		}
	}
	if bytes.HasPrefix(trimmed, []byte("```")) {
		lines := bytes.Split(trimmed, []byte("\n"))
		if len(lines) >= 2 {
			lines = lines[1:]
			if len(lines) > 0 && bytes.HasPrefix(bytes.TrimSpace(lines[len(lines)-1]), []byte("```")) {
				lines = lines[:len(lines)-1]
			}
			trimmed = bytes.TrimSpace(bytes.Join(lines, []byte("\n")))
		}
	}
	return trimmed
}

func cleanMarkdownString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 2 {
			lines = lines[1:]
			if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
				lines = lines[:len(lines)-1]
			}
			s = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
	return s
}

func extractConceptsFromText(text string) []Concept {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// Try splitting by Concept 1 / Concept 2 / Concept 3 or 1. / 2. / 3.
	splitRegex := regexp.MustCompile(`(?i)(?:^|\n)(?:concept\s*\d+[:.]?|\d+[\.\)])\s*`)
	indexes := splitRegex.FindAllStringIndex(text, -1)
	var sections []string
	if len(indexes) >= 2 {
		for i := 0; i < len(indexes); i++ {
			start := indexes[i][1]
			end := len(text)
			if i+1 < len(indexes) {
				end = indexes[i+1][0]
			}
			section := strings.TrimSpace(text[start:end])
			if section != "" {
				sections = append(sections, section)
			}
		}
	} else {
		// Split by double newline
		parts := strings.Split(text, "\n\n")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if len(p) > 20 {
				sections = append(sections, p)
			}
		}
	}

	var results []Concept
	for _, sec := range sections {
		c := parseSingleConceptSection(sec)
		if c.Name != "" || c.BigIdea != "" {
			results = append(results, c)
		}
	}
	return results
}

func parseSingleConceptSection(sec string) Concept {
	lines := strings.Split(sec, "\n")
	var c Concept
	for _, l := range lines {
		line := strings.TrimSpace(l)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "name:") || strings.HasPrefix(lower, "title:"):
			c.Name = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		case strings.HasPrefix(lower, "big idea:") || strings.HasPrefix(lower, "bigidea:") || strings.HasPrefix(lower, "idea:"):
			c.BigIdea = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		case strings.HasPrefix(lower, "visual metaphor:") || strings.HasPrefix(lower, "visualmetaphor:") || strings.HasPrefix(lower, "visual:"):
			c.VisualMetaphor = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		case strings.HasPrefix(lower, "mood:") || strings.HasPrefix(lower, "tone:"):
			c.Mood = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		case strings.HasPrefix(lower, "palette:") || strings.HasPrefix(lower, "color palette:") || strings.HasPrefix(lower, "colors:"):
			c.Palette = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		default:
			if c.Name == "" {
				c.Name = line
			} else if c.BigIdea == "" {
				c.BigIdea = line
			}
		}
	}
	return c
}

func synthesizeFallbackConcepts(brief Brief, style *StyleProfile) []Concept {
	occasion := brief.Occasion
	if occasion == "" {
		occasion = "Celebration Event"
	}
	msg := brief.KeyMessage
	if msg == "" {
		msg = brief.Objective
	}
	paletteStr := strings.Join(brief.Palette, ", ")
	if paletteStr == "" {
		paletteStr = "Deep navy, gold, and warm amber"
	}
	moodStr := strings.Join(brief.Tone, ", ")
	if moodStr == "" {
		moodStr = "Festive, prestigious, and uplifting"
	}

	return []Concept{
		{
			Name:           "Cinematic Hero",
			BigIdea:        fmt.Sprintf("Direct focal celebration of %s emphasizing %s.", occasion, msg),
			VisualMetaphor: fmt.Sprintf("Prominent central illuminated hero motifs with atmospheric motivated lighting representing %s.", occasion),
			Mood:           moodStr,
			Palette:        paletteStr,
		},
		{
			Name:           "Modern Graphic Minimalist",
			BigIdea:        fmt.Sprintf("Clean contemporary focus on %s with bold typography and open composition.", occasion),
			VisualMetaphor: "Geometric balance with deep negative space and clean edge illumination.",
			Mood:           "Crisp, elegant, and sophisticated",
			Palette:        paletteStr,
		},
		{
			Name:           "Vibrant Cultural Radiance",
			BigIdea:        fmt.Sprintf("Dynamic festive energy capturing %s with rich layered depth.", occasion),
			VisualMetaphor: "Multi-layered foreground and background depth with celebratory elements and warm rim light.",
			Mood:           "Energetic, joyous, and welcoming",
			Palette:        paletteStr,
		},
	}
}

func finalizeConcepts(concepts []Concept, brief Brief, style *StyleProfile) []Concept {
	fallback := synthesizeFallbackConcepts(brief, style)
	var cleaned []Concept
	for i, c := range concepts {
		fb := fallback[i%len(fallback)]
		if strings.TrimSpace(c.Name) == "" {
			c.Name = fb.Name
		}
		if strings.TrimSpace(c.BigIdea) == "" {
			c.BigIdea = fb.BigIdea
		}
		if strings.TrimSpace(c.VisualMetaphor) == "" {
			c.VisualMetaphor = fb.VisualMetaphor
		}
		if strings.TrimSpace(c.Mood) == "" {
			c.Mood = fb.Mood
		}
		if strings.TrimSpace(c.Palette) == "" {
			c.Palette = fb.Palette
		}
		cleaned = append(cleaned, c)
	}

	// Ensure at least 3 concepts
	for len(cleaned) < 3 {
		cleaned = append(cleaned, fallback[len(cleaned)%len(fallback)])
	}
	if len(cleaned) > 3 {
		cleaned = cleaned[:3]
	}
	return cleaned
}
