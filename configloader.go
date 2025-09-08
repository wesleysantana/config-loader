package configloader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// LoadOptions configura o comportamento do carregamento
type LoadOptions struct {
	EnvFiles  []string // Caminhos para arquivos .env (opcional)
	UseSystem bool     // Usar variáveis do sistema (default: true)
}

// Load carrega configurações com opções customizadas
func Load(config any, opts ...LoadOptions) error {
	options := LoadOptions{
		UseSystem: true,
	}

	if len(opts) > 0 {
		options = opts[0]
	}

	// Carrega arquivos .env se especificados
	if len(options.EnvFiles) > 0 {
		if err := godotenv.Load(options.EnvFiles...); err != nil {
			return fmt.Errorf("error loading .env files: %w", err)
		}
	} else {
		// Tenta carregar .env na raiz, mas não falha se não existir
		godotenv.Load()
	}

	return loadFromEnv(config, options.UseSystem)
}

// MustLoad loads environment variables and panics if any required field is missing
func MustLoad(config any) {
	if err := Load(config); err != nil {
		panic(err)
	}
}

// LoadFromEnv loads only from environment variables (ignores .env file)
func LoadFromEnv(config any) error {
	return loadFromEnv(config, true)
}

// LoadFromFile carrega de um arquivo .env específico
func LoadFromFile(config any, envFile string) error {
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	return loadFromEnv(config, true)
}

// LoadFromFiles carrega de múltiplos arquivos .env
func LoadFromFiles(config any, envFiles ...string) error {
	if err := godotenv.Load(envFiles...); err != nil {
		return fmt.Errorf("error loading .env files: %w", err)
	}
	return loadFromEnv(config, true)
}

// FindAndLoad procura por arquivos .env em locais comuns
func FindAndLoad(config any) error {
	possiblePaths := []string{
		".env",
		"./.env",
		"../.env",
		"../../.env",
		"./config/.env",
		"./env/.env",
		os.Getenv("ENV_FILE"), // Permite override por variável de ambiente
	}

	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				break
			}
		}
	}

	// Continua com variáveis de sistema mesmo se não encontrou arquivo
	return loadFromEnv(config, true)
}

func loadFromEnv(config interface{}, useSystem bool) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	var validationErrors []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		parts := parseEnvTag(envTag)
		envName := parts[0]

		value := ""
		if useSystem {
			value = os.Getenv(envName)
		}

		// Lógica de default/required
		defaultValue := ""
		if len(parts) > 1 {
			defaultValue = parts[1]
		}

		if value == "" {
			if defaultValue == "required" {
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", envName))
				continue
			} else if defaultValue != "" {
				value = defaultValue
			}
		}

		if value != "" && v.Field(i).CanSet() {
			if err := setFieldValue(v.Field(i), value); err != nil {
				return fmt.Errorf("error setting field %s: %w", field.Name, err)
			}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func parseEnvTag(tag string) []string {
	// Divide a tag em partes, mas preserva o valor default completo
	parts := strings.SplitN(tag, ",", 2)
	if len(parts) == 1 {
		return parts
	}

	// Retorna apenas [nome, valor_default]
	return []string{parts[0], parts[1]}
}
func setFieldValue(field reflect.Value, value string) error {
	// Verifica primeiro se é time.Duration (que é um tipo alias de int64)
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration value '%s': %w", value, err)
		}
		field.SetInt(int64(duration))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s': %w", value, err)
		}
		field.SetInt(intValue)

	case reflect.Bool:
		boolValue, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value '%s': %w", value, err)
		}
		field.SetBool(boolValue)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			sliceValue := parseStringSlice(value)
			field.Set(reflect.ValueOf(sliceValue))
		} else {
			return fmt.Errorf("unsupported slice type: %s", field.Type().Elem().Kind())
		}

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value '%s': %w", value, err)
		}
		field.SetFloat(floatValue)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on", "t":
		return true, nil
	case "false", "0", "no", "off", "f", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}

	// Divide por vírgula e remove espaços em branco
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}

	return result
}

// SPrint returns a string representation of the environment configuration
func SPrint(config any) string {
	var result strings.Builder
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	result.WriteString("Environment Configuration:\n")
	result.WriteString("==========================\n")

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envName := strings.Split(envTag, ",")[0]
		fieldValue := v.Field(i)

		// Esconde valores sensíveis
		displayValue := fieldValue.Interface()
		if shouldMaskField(field.Name) {
			displayValue = "***MASKED***"
		}

		result.WriteString(fmt.Sprintf("%-20s: %v\n", envName, displayValue))
	}

	return result.String()
}

func shouldMaskField(fieldName string) bool {
	maskedKeywords := []string{
		"password", "secret", "key", "token", "credential",
		"auth", "pass", "pwd", "access", "private",
	}

	lowerName := strings.ToLower(fieldName)
	for _, keyword := range maskedKeywords {
		if strings.Contains(lowerName, keyword) {
			return true
		}
	}
	return false
}
