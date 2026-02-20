package diffing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Formatter interface {
	Format(diffs []Diff) (string, error)
}

// TextFormatter produces human-readable colored terminal output.
type TextFormatter struct {
	Colors bool
}

func NewTextFormatter(colors bool) *TextFormatter {
	return &TextFormatter{Colors: colors}
}

func (f *TextFormatter) Format(diffs []Diff) (string, error) {
	var buf bytes.Buffer

	if len(diffs) == 0 {
		buf.WriteString("No differences found.\n")
		return buf.String(), nil
	}

	buf.WriteString(fmt.Sprintf("Found %d difference(s):\n\n", len(diffs)))

	for _, diff := range diffs {
		symbol := f.symbol(diff.Type)
		color := f.color(diff.Severity)
		reset := "\033[0m"

		if f.Colors {
			buf.WriteString(color)
		}

		buf.WriteString(fmt.Sprintf("  %s %s (%s) [%s]\n", symbol, diff.Path, diff.Type, diff.Severity))

		if diff.OldValue != nil {
			buf.WriteString(fmt.Sprintf("    - Old: %v\n", formatValue(diff.OldValue)))
		}
		if diff.NewValue != nil {
			buf.WriteString(fmt.Sprintf("    + New: %v\n", formatValue(diff.NewValue)))
		}

		if f.Colors {
			buf.WriteString(reset)
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func (f *TextFormatter) symbol(t DiffType) string {
	switch t {
	case DiffTypeAdded:
		return "+"
	case DiffTypeRemoved:
		return "-"
	case DiffTypeModified:
		return "~"
	case DiffTypeTypeChanged:
		return "!"
	default:
		return "?"
	}
}

func (f *TextFormatter) color(s Severity) string {
	switch s {
	case SeverityInfo:
		return "\033[36m" // cyan
	case SeverityWarning:
		return "\033[33m" // yellow
	case SeverityError:
		return "\033[31m" // red
	case SeverityBreaking:
		return "\033[1;31m" // bold red
	default:
		return ""
	}
}

// JSONFormatter produces machine-readable JSON output.
type JSONFormatter struct {
	Pretty bool
}

func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{Pretty: pretty}
}

func (f *JSONFormatter) Format(diffs []Diff) (string, error) {
	var data []byte
	var err error

	output := struct {
		TotalChanges int    `json:"total_changes"`
		Diffs        []Diff `json:"diffs"`
		HasBreaking  bool   `json:"has_breaking"`
	}{
		TotalChanges: len(diffs),
		Diffs:        diffs,
		HasBreaking:  hasBreakingChanges(diffs),
	}

	if f.Pretty {
		data, err = json.MarshalIndent(output, "", "  ")
	} else {
		data, err = json.Marshal(output)
	}

	if err != nil {
		return "", fmt.Errorf("marshaling diff output: %w", err)
	}

	return string(data), nil
}

// BreakingChangeFormatter produces output focused on breaking changes only.
type BreakingChangeFormatter struct {
	Colors bool
}

func NewBreakingChangeFormatter(colors bool) *BreakingChangeFormatter {
	return &BreakingChangeFormatter{Colors: colors}
}

func (f *BreakingChangeFormatter) FormatBreakingChanges(changes []BreakingChange) (string, error) {
	var buf bytes.Buffer

	if len(changes) == 0 {
		buf.WriteString("No breaking changes detected.\n")
		return buf.String(), nil
	}

	buf.WriteString(fmt.Sprintf("⚠ Found %d breaking change(s):\n\n", len(changes)))

	for i, change := range changes {
		if f.Colors {
			buf.WriteString("\033[1;31m")
		}
		buf.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, strings.ToUpper(change.Type), change.Path))
		buf.WriteString(fmt.Sprintf("     %s\n", change.Description))
		if f.Colors {
			buf.WriteString("\033[0m")
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func hasBreakingChanges(diffs []Diff) bool {
	for _, d := range diffs {
		if d.Severity == SeverityBreaking {
			return true
		}
	}
	return false
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case map[string]interface{}, []interface{}:
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		if len(data) > 80 {
			return string(data[:77]) + "..."
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}
