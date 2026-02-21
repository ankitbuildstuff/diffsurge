package pii

type RedactionMode string

const (
	ModeRedact RedactionMode = "redact"
	ModeHash   RedactionMode = "hash"
	ModeMask   RedactionMode = "mask"
)

type Config struct {
	Enabled          bool            `mapstructure:"enabled"`
	Mode             RedactionMode   `mapstructure:"mode"`
	Patterns         PatternConfig   `mapstructure:"patterns"`
	CustomPatterns   []CustomPattern `mapstructure:"custom_patterns"`
	ScanHeaders      bool            `mapstructure:"scan_headers"`
	ScanQueryParams  bool            `mapstructure:"scan_query_params"`
	ScanURLPath      bool            `mapstructure:"scan_url_path"`
	ScanRequestBody  bool            `mapstructure:"scan_request_body"`
	ScanResponseBody bool            `mapstructure:"scan_response_body"`
}

type PatternConfig struct {
	Email      bool `mapstructure:"email"`
	Phone      bool `mapstructure:"phone"`
	CreditCard bool `mapstructure:"credit_card"`
	SSN        bool `mapstructure:"ssn"`
	APIKey     bool `mapstructure:"api_key"`
	JWT        bool `mapstructure:"jwt"`
	IPAddress  bool `mapstructure:"ip_address"`
	AWSKey     bool `mapstructure:"aws_key"`
	DOB        bool `mapstructure:"dob"`
}

type CustomPattern struct {
	Name        string `mapstructure:"name"`
	Regex       string `mapstructure:"regex"`
	Replacement string `mapstructure:"replacement"`
}

func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Mode:    ModeRedact,
		Patterns: PatternConfig{
			Email:      true,
			Phone:      true,
			CreditCard: true,
			SSN:        true,
			APIKey:     true,
			JWT:        true,
			IPAddress:  false,
			AWSKey:     true,
			DOB:        true,
		},
		ScanHeaders:      true,
		ScanQueryParams:  true,
		ScanURLPath:      true,
		ScanRequestBody:  true,
		ScanResponseBody: true,
	}
}

func (c *Config) enabledPatternTypes() map[PatternType]bool {
	return map[PatternType]bool{
		PatternEmail:      c.Patterns.Email,
		PatternPhone:      c.Patterns.Phone,
		PatternCreditCard: c.Patterns.CreditCard,
		PatternSSN:        c.Patterns.SSN,
		PatternAPIKey:     c.Patterns.APIKey,
		PatternJWT:        c.Patterns.JWT,
		PatternIPv4:       c.Patterns.IPAddress,
		PatternIPv6:       c.Patterns.IPAddress,
		PatternAWSKey:     c.Patterns.AWSKey,
		PatternDOB:        c.Patterns.DOB,
	}
}
