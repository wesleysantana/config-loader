Requisitos de Versão
Go 1.18 ou superior é necessário para utilizar esta biblioteca.

📦 Instalação
go get github.com/wesleysantana/config-loader

🚀 Uso Rápido
```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/wesleysantana/config-loader"
)

type Config struct {
    ServerPort  string        `env:"SERVER_PORT,5000"`
    DBHost      string        `env:"DB_HOST,localhost"`
    DBPassword  string        `env:"DB_PASSWORD,required"`
    DebugMode   bool          `env:"DEBUG_MODE,false"`
    MaxUsers    int           `env:"MAX_USERS,100"`
    Timeout     time.Duration `env:"TIMEOUT,30s"`
    AllowedHosts []string     `env:"ALLOWED_HOSTS,localhost,127.0.0.1"`
}

func main() {
    var cfg Config
    
    // Carrega automaticamente (procura .env em locais comuns)
    if err := envconfig.Load(&cfg); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Server running on port %s\n", cfg.ServerPort)
    fmt.Println(envconfig.SPrint(cfg))
}
```
✨ Características
- Tipo Seguro: Suporte nativo para strings, integers, booleans, slices e time.Duration
- Flexível: Múltiplas estratégias de carregamento
- Validação: Campos required e validações customizáveis
- Extensível: Interface simples para adicionar novos tipos
- Multi-ambiente: Suporte a diferentes arquivos .env por ambiente
- Segurança: Mascaramento automático de campos sensíveis
- Time Support: Suporte nativo a time.Duration

📚 Métodos de Carregamento
- Carregamento Automático
```go
    // Procura por .env em locais comuns (raiz, config/, ../, etc)
    err := envconfig.Load(&cfg)
```	
- Arquivo Específico
```go
    // Carrega de um arquivo específico
    err := envconfig.LoadFromFile(&cfg, "./config/production.env")
```
- Múltiplos Arquivos 
```go
    // Carrega de múltiplos arquivos
    err := envconfig.LoadFromFiles(&cfg, "./config/base.env", "./config/production.env")
```
- Apenas Variáveis de Sistema
```go
    // Ignora arquivos .env, usa apenas variáveis de sistema
    err := envconfig.LoadFromSystem(&cfg)
```
- Busca Inteligente
```go
    // Procura em locais comuns automaticamente
    err := envconfig.FindAndLoad(&cfg)
```
- Carregamento com Opções
```go
    err := envconfig.Load(&cfg, envconfig.LoadOptions{
        EnvFiles:  []string{"./config/.env", "./config/secrets.env"},
        UseSystem: true,
    })
```	
- Carregamento com Panic
```go
    // Panic se houver erro de validação
    envconfig.MustLoad(&cfg)
```
🏷️ Tag Syntax
```go
    type Config struct {
        // Valor padrão "3000"
        Port string `env:"PORT,3000"`
        
        // Campo obrigatório
        APIKey string `env:"API_KEY,required"`
        
        // Tipo booleano com default
        Debug bool `env:"DEBUG,false"`
        
        // Tipo inteiro
        Timeout int `env:"TIMEOUT,30"`
        
        // Slice de strings (separados por vírgula)
        Hosts []string `env:"HOSTS,localhost:8080,api.example.com"`
        
        // Duração com suporte a unidades (s, m, h)
        ShutdownDelay time.Duration `env:"SHUTDOWN_DELAY,10s"`
        
        // Sem tag - ignorado
        Internal string
    }
``` 
🛡️ Validação
A biblioteca valida automaticamente campos marcados como required:
```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL,required"`
}

// Se DATABASE_URL não estiver definida:
// Error: validation errors: DATABASE_URL is required
```
🌳 Hierarquia de Valores
1. Variáveis de ambiente do sistema (mais alta precedência)
2. Arquivos .env (carregados na ordem especificada)
3. Valores default da tag env (menor precedência)

🔧 Tipos Suportados
* string - Valores textuais
* int, int8, int16, int32, int64 - Números inteiros
* bool - Valores booleanos (true, 1, yes, on, false, 0, no, off)
* []string - Slices (separados por vírgula)
* time.Duration - Durações (ex: "30s", "5m", "1h")

🔒 Mascaramento de Campos Sensíveis
A função SPrint() mascara automaticamente campos que contenham palavras sensíveis:
```go
    // No output isso aparecerá como "***MASKED***"
    DBPassword string `env:"DB_PASSWORD,required"`
    APIKey     string `env:"API_KEY,secret123"`
```	
Palavras-chave que ativam o mascaramento: password, secret, key, token, credential, auth, pass, pwd, access, private.

📊 Visualização de Configuração
```go
    // Exibe todas as variáveis de ambiente configuradas
    fmt.Println(envconfig.SPrint(cfg))

    // Output:
    // Environment Configuration:
    // ==========================
    // SERVER_PORT         : 5000
    // DB_HOST             : localhost
    // DB_PASSWORD         : ***MASKED***
    // DEBUG_MODE          : false
    // MAX_USERS           : 100
    // TIMEOUT             : 30s
    // ALLOWED_HOSTS       : [localhost 127.0.0.1]
```
