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

// LoadOptions configura o comportamento do carregamento de variáveis de ambiente.
// Use esta struct para personalizar como as variáveis são carregadas.
type LoadOptions struct {
	// EnvFiles especifica os caminhos para arquivos .env a serem carregados.
	// Se vazio, a biblioteca tentará carregar de locais comuns.
	EnvFiles []string

	// UseSystem determina se variáveis de ambiente do sistema devem ser usadas.
	// Padrão: true. Se false, apenas arquivos .env serão considerados.
	UseSystem bool
}

// Load carrega configurações a partir de variáveis de ambiente e arquivos .env.
// Esta função é a principal entrada da biblioteca e oferece flexibilidade para
// diferentes cenários de carregamento.
//
// Parâmetros:
//   - config: Ponteiro para uma struct com tags `env` para mapeamento
//   - opts: Opções de carregamento (opcional)
//
// Exemplo:
//
//	type Config struct {
//	    Port string `env:"PORT,8080"`
//	    Host string `env:"HOST,localhost"`
//	}
//
//	var cfg Config
//	err := Load(&cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Retorna:
//   - error: Erro se a validação falhar ou se ocorrer problema no carregamento
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

// MustLoad carrega configurações e entra em panic se qualquer campo required estiver faltando.
// Use esta função quando quiser garantir que todas as configurações obrigatórias estejam presentes.
//
// Parâmetros:
//   - config: Ponteiro para uma struct com tags `env`
//
// Panics:
//   - Se campos marcados como "required" não estiverem presentes
//
// Exemplo:
//
//	var cfg Config
//	MustLoad(&cfg) // Panic se faltar DB_PASSWORD
func MustLoad(config any) {
	if err := Load(config); err != nil {
		panic(err)
	}
}

// LoadFromEnv carrega configurações apenas a partir de variáveis de ambiente do sistema,
// ignorando completamente arquivos .env.
//
// Parâmetros:
//   - config: Ponteiro para uma struct com tags `env`
//
// Retorna:
//   - error: Erro se a validação falhar
//
// Exemplo:
//
//	err := LoadFromEnv(&cfg) // Apenas variáveis de sistema
func LoadFromEnv(config any) error {
	return loadFromEnv(config, true)
}

// LoadFromFile carrega configurações a partir de um arquivo .env específico.
// Ideal para ambientes específicos como production.env, development.env, etc.
//
// Parâmetros:
//   - config: Ponteiro para uma struct com tags `env`
//   - envFile: Caminho para o arquivo .env
//
// Retorna:
//   - error: Erro se o arquivo não existir ou se a validação falhar
//
// Exemplo:
//
//	err := LoadFromFile(&cfg, "config/production.env")
func LoadFromFile(config any, envFile string) error {
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	return loadFromEnv(config, true)
}

// LoadFromFiles carrega configurações a partir de múltiplos arquivos .env.
// Útil para separar configurações em diferentes arquivos (ex: .env.base, .env.secrets).
// O último arquivo na lista tem precedência (override).
//
// Parâmetros:
//   - config: Ponteiro para uma struct com tags `env`
//   - envFiles: Lista de caminhos para arquivos .env
//
// Retorna:
//   - error: Erro se algum arquivo não existir ou se a validação falhar
//
// Exemplo:
//
//	err := LoadFromFiles(&cfg, ".env.defaults", ".env.local")
func LoadFromFiles(config any, envFiles ...string) error {
	if err := godotenv.Load(envFiles...); err != nil {
		return fmt.Errorf("error loading .env files: %w", err)
	}
	return loadFromEnv(config, true)
}

// FindAndLoad procura automaticamente por arquivos .env em locais comuns
// e carrega as configurações. Útil quando não se sabe o local exato do arquivo.
//
// Locais pesquisados:
//   - .env (raiz do projeto)
//   - ./.env
//   - ../.env (um nível acima)
//   - ../../.env (dois níveis acima)
//   - ./config/.env
//   - ./env/.env
//   - caminho da variável de ambiente ENV_FILE
//
// Parâmetros:
//   - config: Ponteiro para uma struct com tags `env`
//
// Retorna:
//   - error: Erro se a validação falhar
//
// Exemplo:
//
//	err := FindAndLoad(&cfg) // Busca automática
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

// SPrint retorna uma representação string formatada das configurações carregadas.
// Campos sensíveis (com palavras como password, secret, key) são mascarados.
//
// Parâmetros:
//   - config: Struct com as configurações carregadas
//
// Retorna:
//   - string: Configurações formatadas para visualização
//
// Exemplo:
//
//	fmt.Println(SPrint(cfg))
//	// Output:
//	// Environment Configuration:
//	// ==========================
//	// SERVER_PORT: 8080
//	// DB_PASSWORD: ***MASKED***
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

// loadFromEnv é a função interna que realiza o carregamento das variáveis de ambiente
// para a struct configurada.
func loadFromEnv(config any, useSystem bool) error {
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

		// Lógica de default/required - agora parts[1] contém o valor completo
		if value == "" && len(parts) > 1 {
			defaultValue := parts[1]
			if defaultValue == "required" {
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", envName))
			} else {
				value = defaultValue // Usa o valor default completo
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

// parseEnvTag parseia a tag `env` extraindo o nome da variável e valores default.
// Suporta formatos: "VAR_NAME", "VAR_NAME,default", "VAR_NAME,required"
// Usa SplitN com limite 2 para dividir apenas na primeira vírgula
func parseEnvTag(tag string) []string {
	if tag == "" {
		return []string{""}
	}

	// Divide a tag em no máximo 2 partes: nome e valor default
	// Usa SplitN com limite 2 para preservar vírgulas no valor default
	parts := strings.SplitN(tag, ",", 2)
	return parts
}

// setFieldValue define o valor de um campo baseado no seu tipo e no valor string fornecido.
// Suporta: string, int, bool, []string, time.Duration, float64
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

// parseBool converte uma string para valor booleano.
// Aceita: "true", "1", "yes", "on", "t" → true
//
//	"false", "0", "no", "off", "f", "" → false
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

// parseStringSlice converte uma string separada por vírgulas em slice de strings.
// Remove espaços em branco e ignora valores vazios.
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

// shouldMaskField determina se um campo deve ser mascarado na exibição.
// Campos com nomes contendo: password, secret, key, token, credential, auth, pass, pwd, access, private
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
