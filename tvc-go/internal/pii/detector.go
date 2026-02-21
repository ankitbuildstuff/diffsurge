package pii

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sync"
)

type Detection struct {
	Type     PatternType `json:"type"`
	Path     string      `json:"path"`
	Redacted bool        `json:"redacted"`
}

type ScanResult struct {
	Found      bool        `json:"found"`
	Detections []Detection `json:"detections,omitempty"`
}

type Detector struct {
	patterns       []Pattern
	mode           RedactionMode
	config         Config
	customPatterns []Pattern
	bufPool        sync.Pool
}

func NewDetector(cfg Config) *Detector {
	enabled := cfg.enabledPatternTypes()
	var active []Pattern
	for _, p := range defaultPatterns {
		if enabled[p.Type] {
			active = append(active, p)
		}
	}

	var custom []Pattern
	for _, cp := range cfg.CustomPatterns {
		re, err := regexp.Compile(cp.Regex)
		if err != nil {
			continue
		}
		custom = append(custom, Pattern{
			Type:        PatternType(cp.Name),
			Regex:       re,
			Replacement: cp.Replacement,
		})
	}

	return &Detector{
		patterns:       active,
		mode:           cfg.Mode,
		config:         cfg,
		customPatterns: custom,
		bufPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 4096)
			},
		},
	}
}

func (d *Detector) ScanAndRedact(data map[string]interface{}) ScanResult {
	result := ScanResult{}
	d.walkAndRedact(data, "", &result)
	return result
}

func (d *Detector) ScanString(s string) []Detection {
	var detections []Detection
	allPatterns := append(d.patterns, d.customPatterns...)

	for _, p := range allPatterns {
		matches := p.Regex.FindAllString(s, -1)
		for _, match := range matches {
			if p.Validate != nil && !p.Validate(match) {
				continue
			}
			detections = append(detections, Detection{
				Type:     p.Type,
				Redacted: true,
			})
		}
	}
	return detections
}

func (d *Detector) RedactString(s string) (string, bool) {
	redacted := false
	result := s
	allPatterns := append(d.patterns, d.customPatterns...)

	for _, p := range allPatterns {
		indices := p.Regex.FindAllStringIndex(result, -1)
		if len(indices) == 0 {
			continue
		}

		validIndices := indices
		if p.Validate != nil {
			validIndices = nil
			for _, idx := range indices {
				match := result[idx[0]:idx[1]]
				if p.Validate(match) {
					validIndices = append(validIndices, idx)
				}
			}
		}

		if len(validIndices) == 0 {
			continue
		}

		redacted = true
		for i := len(validIndices) - 1; i >= 0; i-- {
			idx := validIndices[i]
			match := result[idx[0]:idx[1]]
			replacement := d.replacementFor(p, match)
			result = result[:idx[0]] + replacement + result[idx[1]:]
		}
	}
	return result, redacted
}

func (d *Detector) replacementFor(p Pattern, match string) string {
	switch d.mode {
	case ModeHash:
		hash := sha256.Sum256([]byte(match))
		return fmt.Sprintf("[%s:%x]", p.Type, hash[:8])
	case ModeMask:
		if p.Mask != nil {
			return p.Mask(match)
		}
		return maskGeneric(match)
	default:
		return p.Replacement
	}
}

func (d *Detector) walkAndRedact(data map[string]interface{}, prefix string, result *ScanResult) {
	for key, val := range data {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		redactedKey, keyRedacted := d.RedactString(key)
		if keyRedacted {
			result.Found = true
			result.Detections = append(result.Detections, Detection{
				Type:     "key",
				Path:     path,
				Redacted: true,
			})
			delete(data, key)
			data[redactedKey] = val
			key = redactedKey
		}

		switch v := val.(type) {
		case string:
			redactedVal, wasRedacted := d.RedactString(v)
			if wasRedacted {
				result.Found = true
				data[key] = redactedVal
				detections := d.ScanString(v)
				for i := range detections {
					detections[i].Path = path
				}
				result.Detections = append(result.Detections, detections...)
			}

		case map[string]interface{}:
			d.walkAndRedact(v, path, result)

		case []interface{}:
			d.walkSlice(v, path, result)
		}
	}
}

func (d *Detector) walkSlice(data []interface{}, prefix string, result *ScanResult) {
	for i, val := range data {
		path := fmt.Sprintf("%s[%d]", prefix, i)

		switch v := val.(type) {
		case string:
			redactedVal, wasRedacted := d.RedactString(v)
			if wasRedacted {
				result.Found = true
				data[i] = redactedVal
				detections := d.ScanString(v)
				for j := range detections {
					detections[j].Path = path
				}
				result.Detections = append(result.Detections, detections...)
			}

		case map[string]interface{}:
			d.walkAndRedact(v, path, result)

		case []interface{}:
			d.walkSlice(v, path, result)
		}
	}
}
