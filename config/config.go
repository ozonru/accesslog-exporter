package config

import (
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

const (
	defaultUserAgentCacheSize int = 100000
	defaultExportWorkers      int = 100
)

// Config contains all config of application
type Config struct {
	Global  Global   `yaml:"global"`
	Sources []Source `yaml:"sources"`
}

// Global contains global config settings
type Global struct {
	InternalSubnets    []string `yaml:"internal_subnets"`
	UserAgentCacheSize int      `yaml:"user_agent_cache_size"`
	ExportWorkers      int      `yaml:"export_workers"`

	UserAgentReplacementSettingsRaw []struct {
		MatchRe      string               `yaml:"match_re"`
		Match        string               `yaml:"match"`
		Replacements UserAgentReplacement `yaml:"replacements"`
	} `yaml:"user_agents"`
	RequestURIReplacementSettingsRaw []struct {
		MatchRe      string                `yaml:"match_re"`
		MatchMethod  string                `yaml:"match_method"`
		Replacements RequestURIReplacement `yaml:"replacements"`
	} `yaml:"request_uris"`

	Hosts []Host `yaml:"hosts"`

	// compiled settings
	UserAgentReplacementSettings  []UserAgentReplacementSetting
	RequestURIReplacementSettings []RequestURIReplacementSetting
}

// UserAgentReplacementSetting is a set of settings to replace user agent with custom value
type UserAgentReplacementSetting struct {
	MatchRe      *regexp.Regexp
	Match        string
	Replacements UserAgentReplacement
}

// RequestURIReplacementSetting is a set of settings to replace request URI with custom value
type RequestURIReplacementSetting struct {
	Method       string
	Regexp       *regexp.Regexp
	Replacements RequestURIReplacement
}

// UserAgentReplacement contains device, os name and user agent for replacement
type UserAgentReplacement struct {
	Device    string `yaml:"device"`
	Os        string `yaml:"os"`
	UserAgent string `yaml:"user_agent"`
}

// RequestURIReplacement contains request uri to replace
type RequestURIReplacement struct {
	RequestURI string `yaml:"request_uri"`
}

// Source contains host and log format for parsing log lines
type Source struct {
	Host      string `yaml:"host"`
	LogFormat string `yaml:"log_format"`
}

// Host contains replacements for host label
type Host struct {
	Match       string `yaml:"match"`
	Replacement string `yaml:"replacement"`
}

// MakeConfigFromFile loads file and makes config
func MakeConfigFromFile(path string) (*Config, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{Global: Global{UserAgentCacheSize: defaultUserAgentCacheSize, ExportWorkers: defaultExportWorkers}}
	err = yaml.Unmarshal(raw, cfg)
	if err != nil {
		return nil, err
	}

	for _, rep := range cfg.Global.UserAgentReplacementSettingsRaw {
		if rep.MatchRe != "" {
			cfg.Global.UserAgentReplacementSettings = append(cfg.Global.UserAgentReplacementSettings, UserAgentReplacementSetting{
				MatchRe:      regexp.MustCompile(rep.MatchRe),
				Replacements: rep.Replacements,
			})
		} else if rep.Match != "" {
			cfg.Global.UserAgentReplacementSettings = append(cfg.Global.UserAgentReplacementSettings, UserAgentReplacementSetting{
				Match:        rep.Match,
				Replacements: rep.Replacements,
			})
		}
	}

	for _, rep := range cfg.Global.RequestURIReplacementSettingsRaw {
		cfg.Global.RequestURIReplacementSettings = append(cfg.Global.RequestURIReplacementSettings, RequestURIReplacementSetting{
			Method:       rep.MatchMethod,
			Regexp:       regexp.MustCompile(rep.MatchRe),
			Replacements: rep.Replacements,
		})
	}

	return cfg, err
}
