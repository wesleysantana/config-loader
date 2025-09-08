package configloader

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestConfig struct para testes
type TestConfig struct {
	ServerPort   string        `env:"SERVER_PORT,8080"`
	DBHost       string        `env:"DB_HOST,localhost"`
	DBPassword   string        `env:"DB_PASSWORD,required"`
	DebugMode    bool          `env:"DEBUG_MODE,false"`
	MaxUsers     int           `env:"MAX_USERS,100"`
	Timeout      time.Duration `env:"TIMEOUT,30s"`
	AllowedHosts []string      `env:"ALLOWED_HOSTS,localhost,127.0.0.1"`
	APIKey       string        `env:"API_KEY"`
	FloatValue   float64       `env:"FLOAT_VALUE,3.14"`
}

// TestConfigWithoutTags struct para testar campos sem tags
type TestConfigWithoutTags struct {
	ServerPort string
	DBHost     string `env:"DB_HOST"`
}

// TestLoad_Basic testa o carregamento básico
func TestLoad_Basic(t *testing.T) {
	// Configura variáveis de ambiente para o teste
	os.Setenv("DB_PASSWORD", "test123")
	defer os.Unsetenv("DB_PASSWORD")

	var cfg TestConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("Expected ServerPort 8080, got %s", cfg.ServerPort)
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("Expected DBHost localhost, got %s", cfg.DBHost)
	}

	if cfg.DBPassword != "test123" {
		t.Errorf("Expected DBPassword test123, got %s", cfg.DBPassword)
	}

	if cfg.DebugMode != false {
		t.Errorf("Expected DebugMode false, got %v", cfg.DebugMode)
	}

	if cfg.MaxUsers != 100 {
		t.Errorf("Expected MaxUsers 100, got %d", cfg.MaxUsers)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", cfg.Timeout)
	}

	expectedHosts := []string{"localhost", "127.0.0.1"}
	if len(cfg.AllowedHosts) != len(expectedHosts) {
		t.Errorf("Expected %d hosts, got %d", len(expectedHosts), len(cfg.AllowedHosts))
	}

	for i, host := range expectedHosts {
		if cfg.AllowedHosts[i] != host {
			t.Errorf("Expected host %s, got %s", host, cfg.AllowedHosts[i])
		}
	}

	for i, host := range expectedHosts {
		if cfg.AllowedHosts[i] != host {
			t.Errorf("Expected host %s, got %s", host, cfg.AllowedHosts[i])
		}
	}
}

// TestLoad_SliceValues testa valores de slice
func TestLoad_SliceValues(t *testing.T) {
	type SliceConfig struct {
		Hosts       []string `env:"HOSTS,localhost,127.0.0.1"`
		HostsSpaces []string `env:"HOSTS_SPACES, localhost , 127.0.0.1 "`
		HostsEmpty  []string `env:"HOSTS_EMPTY,"`
		HostsSingle []string `env:"HOSTS_SINGLE,localhost"`
	}

	var cfg SliceConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Teste HOSTS
	expectedHosts := []string{"localhost", "127.0.0.1"}
	if len(cfg.Hosts) != len(expectedHosts) {
		t.Errorf("Expected %d hosts, got %d", len(expectedHosts), len(cfg.Hosts))
	}
	for i, host := range expectedHosts {
		if cfg.Hosts[i] != host {
			t.Errorf("Expected host %s, got %s", host, cfg.Hosts[i])
		}
	}

	// Teste HOSTS_SPACES (com espaços)
	expectedHostsSpaces := []string{"localhost", "127.0.0.1"}
	if len(cfg.HostsSpaces) != len(expectedHostsSpaces) {
		t.Errorf("Expected %d hosts with spaces, got %d", len(expectedHostsSpaces), len(cfg.HostsSpaces))
	}
	for i, host := range expectedHostsSpaces {
		if cfg.HostsSpaces[i] != host {
			t.Errorf("Expected host %s, got %s", host, cfg.HostsSpaces[i])
		}
	}

	// Teste HOSTS_EMPTY
	if len(cfg.HostsEmpty) != 0 {
		t.Errorf("Expected empty slice, got %v", cfg.HostsEmpty)
	}

	// Teste HOSTS_SINGLE
	if len(cfg.HostsSingle) != 1 || cfg.HostsSingle[0] != "localhost" {
		t.Errorf("Expected single host localhost, got %v", cfg.HostsSingle)
	}
}

// TestLoad_RequiredField testa campo obrigatório
func TestLoad_RequiredField(t *testing.T) {
	// Garante que a variável não está definida
	os.Unsetenv("DB_PASSWORD")

	var cfg TestConfig
	err := Load(&cfg)

	if err == nil {
		t.Fatal("Expected error for required field, got nil")
	}

	if !strings.Contains(err.Error(), "DB_PASSWORD is required") {
		t.Errorf("Expected required field error, got: %v", err)
	}
}

// TestLoad_EnvironmentOverride testa override por variável de ambiente
func TestLoad_EnvironmentOverride(t *testing.T) {
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("DB_PASSWORD", "test456")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_PASSWORD")
	}()

	var cfg TestConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ServerPort != "3000" {
		t.Errorf("Expected ServerPort 3000 (from env), got %s", cfg.ServerPort)
	}

	if cfg.DBPassword != "test456" {
		t.Errorf("Expected DBPassword test456, got %s", cfg.DBPassword)
	}
}

// TestLoad_WithoutTags testa struct sem tags
func TestLoad_WithoutTags(t *testing.T) {
	var cfg TestConfigWithoutTags
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Campos sem tags devem permanecer com valores zero
	if cfg.ServerPort != "" {
		t.Errorf("Expected empty ServerPort, got %s", cfg.ServerPort)
	}
}

// TestLoadFromEnv_Only testa carregamento apenas de variáveis de ambiente
func TestLoadFromEnv_Only(t *testing.T) {
	os.Setenv("SERVER_PORT", "4000")
	os.Setenv("DB_PASSWORD", "env_only")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_PASSWORD")
	}()

	var cfg TestConfig
	err := LoadFromEnv(&cfg)
	if err != nil {
		t.Fatalf("LoadFromEnv failed: %v", err)
	}

	if cfg.ServerPort != "4000" {
		t.Errorf("Expected ServerPort 4000, got %s", cfg.ServerPort)
	}
}

// TestMustLoad testa MustLoad com panic
func TestMustLoad(t *testing.T) {
	// Deve panic sem DB_PASSWORD
	os.Unsetenv("DB_PASSWORD")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic from MustLoad, but none occurred")
		}
	}()

	var cfg TestConfig
	MustLoad(&cfg)
}

// TestSPrint testa a função de impressão
func TestSPrint(t *testing.T) {
	os.Setenv("DB_PASSWORD", "secret123")
	defer os.Unsetenv("DB_PASSWORD")

	var cfg TestConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	result := SPrint(cfg)

	// Verifica se campos sensíveis estão mascarados
	if !strings.Contains(result, "***MASKED***") {
		t.Error("Expected masked sensitive fields in SPrint output")
	}

	// Verifica se campos não sensíveis são mostrados
	if !strings.Contains(result, "8080") {
		t.Error("Expected non-sensitive fields to be visible in SPrint output")
	}
}

// TestLoad_InvalidTypes testa tipos inválidos
func TestLoad_InvalidTypes(t *testing.T) {
	type InvalidConfig struct {
		InvalidField complex64 `env:"INVALID_FIELD,100"`
	}

	os.Setenv("INVALID_FIELD", "100")
	defer os.Unsetenv("INVALID_FIELD")

	var cfg InvalidConfig
	err := Load(&cfg)

	if err == nil {
		t.Error("Expected error for unsupported type, got nil")
	}
}

// TestLoad_BoolValues testa valores booleanos
func TestLoad_BoolValues(t *testing.T) {
	type BoolConfig struct {
		BoolTrue  bool `env:"BOOL_TRUE,true"`
		BoolFalse bool `env:"BOOL_FALSE,false"`
		BoolOne   bool `env:"BOOL_ONE,1"`
		BoolZero  bool `env:"BOOL_ZERO,0"`
	}

	var cfg BoolConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !cfg.BoolTrue {
		t.Error("Expected BoolTrue to be true")
	}

	if cfg.BoolFalse {
		t.Error("Expected BoolFalse to be false")
	}

	if !cfg.BoolOne {
		t.Error("Expected BoolOne to be true")
	}

	if cfg.BoolZero {
		t.Error("Expected BoolZero to be false")
	}
}

// TestLoad_DurationValues testa valores de duração
func TestLoad_DurationValues(t *testing.T) {
	type DurationConfig struct {
		Seconds time.Duration `env:"SECONDS,30s"`
		Minutes time.Duration `env:"MINUTES,5m"`
		Hours   time.Duration `env:"HOURS,2h"`
	}

	var cfg DurationConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Seconds != 30*time.Second {
		t.Errorf("Expected 30s, got %v", cfg.Seconds)
	}

	if cfg.Minutes != 5*time.Minute {
		t.Errorf("Expected 5m, got %v", cfg.Minutes)
	}

	if cfg.Hours != 2*time.Hour {
		t.Errorf("Expected 2h, got %v", cfg.Hours)
	}
}

// TestLoad_FloatValues testa valores float
func TestLoad_FloatValues(t *testing.T) {
	os.Setenv("FLOAT_ENV", "2.71")
	defer os.Unsetenv("FLOAT_ENV")

	type FloatConfig struct {
		FloatDefault float64 `env:"FLOAT_DEFAULT,3.14"`
		FloatEnv     float64 `env:"FLOAT_ENV"`
	}

	var cfg FloatConfig
	err := Load(&cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.FloatDefault != 3.14 {
		t.Errorf("Expected FloatDefault 3.14, got %f", cfg.FloatDefault)
	}

	if cfg.FloatEnv != 2.71 {
		t.Errorf("Expected FloatEnv 2.71, got %f", cfg.FloatEnv)
	}
}

// TestLoad_WithOptions testa carregamento com opções
func TestLoad_WithOptions(t *testing.T) {
	os.Setenv("TEST_VAR", "from_env")
	defer os.Unsetenv("TEST_VAR")

	type SimpleConfig struct {
		TestVar string `env:"TEST_VAR,default_value"`
	}

	var cfg SimpleConfig
	err := Load(&cfg, LoadOptions{
		UseSystem: true,
	})
	if err != nil {
		t.Fatalf("Load with options failed: %v", err)
	}

	if cfg.TestVar != "from_env" {
		t.Errorf("Expected from_env, got %s", cfg.TestVar)
	}
}

// TestLoad_NonPointer testa erro com não-pointer
func TestLoad_NonPointer(t *testing.T) {
	var cfg TestConfig
	err := Load(cfg) // Erro: não é pointer

	if err == nil {
		t.Error("Expected error for non-pointer, got nil")
	}

	if !strings.Contains(err.Error(), "pointer") {
		t.Errorf("Expected pointer error, got: %v", err)
	}
}

// TestLoad_NonStruct testa erro com não-struct
func TestLoad_NonStruct(t *testing.T) {
	var cfg string
	err := Load(&cfg) // Erro: não é struct

	if err == nil {
		t.Error("Expected error for non-struct, got nil")
	}

	if !strings.Contains(err.Error(), "struct") {
		t.Errorf("Expected struct error, got: %v", err)
	}
}

// TestParseEnvTag testa o parsing correto das tags
func TestParseEnvTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected []string
	}{
		{"Simple", "PORT,8080", []string{"PORT", "8080"}},
		{"Required", "PASSWORD,required", []string{"PASSWORD", "required"}},
		{"Slice", "HOSTS,localhost,127.0.0.1", []string{"HOSTS", "localhost,127.0.0.1"}},
		{"NoDefault", "DEBUG", []string{"DEBUG"}},
		{"Empty", "", []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEnvTag(tt.tag)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(result))
				return
			}
			for i, part := range tt.expected {
				if result[i] != part {
					t.Errorf("Part %d: expected %s, got %s", i, part, result[i])
				}
			}
		})
	}
}
