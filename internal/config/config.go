package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig
	LLM     LLMConfig
	Browser BrowserConfig
	Agent   AgentConfig
	Logging LoggingConfig

	Env EnvConfig
}

type EnvConfig struct {
	Env                string
	ZaiAPIKey          string
	ZaiBaseURL         string
	BrowserUserDataDir string
	BrowserHeadless    bool
	BrowserSlowMoMs    int
}

type AppConfig struct {
	Env  string
	Name string
}

type LLMConfig struct {
	Provider    string  `mapstructure:"provider"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float32 `mapstructure:"temperature"`
}

type BrowserConfig struct {
	Engine   string
	Viewport struct {
		Width  int
		Height int
	}
	TimeoutMs int
}

type AgentConfig struct {
	MaxSteps        int
	AskConfirmation bool
	Memory          struct {
		ShortTermSteps  int
		MaxPageElements int
	}
}

type LoggingConfig struct {
	Level string
}

func Load(configPath string) (*Config, error) {
	_ = godotenv.Load() // .env optional

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.Env = loadEnv()

	return cfg, nil
}

func loadEnv() EnvConfig {
	dir := getEnv("BROWSER_USER_DATA_DIR", "./data/browser")

	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic("invalid BROWSER_USER_DATA_DIR")
	}

	return EnvConfig{
		Env:                getEnv("APP_ENV", "local"),
		ZaiAPIKey:          mustEnv("ZAI_API_KEY"),
		ZaiBaseURL:         getEnv("ZAI_BASE_URL", "https://api.z.ai/v1"),
		BrowserUserDataDir: absDir,
		BrowserHeadless:    getEnvBool("BROWSER_HEADLESS", false),
		BrowserSlowMoMs:    getEnvInt("BROWSER_SLOW_MO_MS", 0),
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return val
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		fmt.Sscanf(v, "%d", &i)
		return i
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return def
}
