package replayer

import (
	"strings"

	"github.com/tvc-org/tvc/internal/models"
)

// SensitiveHeaders that must be stripped or replaced before replay.
var SensitiveHeaders = []string{
	"Authorization",
	"Cookie",
	"Set-Cookie",
	"X-Api-Key",
	"X-Auth-Token",
	"Proxy-Authorization",
}

type HeaderReplacements map[string]string

type FilterConfig struct {
	StripSensitiveHeaders bool
	HeaderReplacements    HeaderReplacements
	ExcludePaths          []string
}

func FilterTrafficForReplay(logs []models.TrafficLog, cfg FilterConfig) []models.TrafficLog {
	filtered := make([]models.TrafficLog, 0, len(logs))

	for i := range logs {
		if shouldExclude(logs[i].Path, cfg.ExcludePaths) {
			continue
		}

		log := logs[i]

		if log.RequestHeaders != nil {
			sanitized := make(map[string]interface{}, len(log.RequestHeaders))
			for k, v := range log.RequestHeaders {
				sanitized[k] = v
			}

			if cfg.StripSensitiveHeaders {
				for _, h := range SensitiveHeaders {
					delete(sanitized, h)
					delete(sanitized, strings.ToLower(h))
				}
			}

			for k, v := range cfg.HeaderReplacements {
				if _, exists := sanitized[k]; exists {
					sanitized[k] = v
				}
				if _, exists := sanitized[strings.ToLower(k)]; exists {
					sanitized[strings.ToLower(k)] = v
				}
			}

			log.RequestHeaders = sanitized
		}

		filtered = append(filtered, log)
	}

	return filtered
}

func shouldExclude(path string, excludePaths []string) bool {
	for _, p := range excludePaths {
		if strings.HasSuffix(p, "*") {
			if strings.HasPrefix(path, strings.TrimSuffix(p, "*")) {
				return true
			}
		} else if path == p {
			return true
		}
	}
	return false
}
