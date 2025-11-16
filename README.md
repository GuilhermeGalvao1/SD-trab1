# SD-trab1 - Cliente Multi-Protocolo para Sistemas Distribuídos

Este projeto implementa um cliente em Go capaz de se comunicar com um servidor remoto utilizando três protocolos distintos: String, JSON e Protocol Buffers. Desenvolvido como trabalho da disciplina de Sistemas Distribuídos da UFC.

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Protocolos Suportados](#protocolos-suportados)
- [Componentes](#componentes)
- [Requisitos](#requisitos)
- [Instalação](#instalação)
- [Uso](#uso)
- [Operações Disponíveis](#operações-disponíveis)

## 🎯 Visão Geral

O projeto consiste em uma aplicação cliente que se conecta a um servidor remoto (IP: 3.88.99.255) e executa uma sequência de operações utilizando diferentes protocolos de comunicação. Cada protocolo opera em uma porta específica e possui sua própria implementação de serialização/deserialização de mensagens.

## 📁 Estrutura do Projeto

```
SD-trab1/
├── main.go                 # Ponto de entrada da aplicação
├── go.mod                  # Dependências do módulo Go
├── client/                 # Implementações dos clientes
│   ├── client.go          # Interface e estruturas de dados
│   ├── base.go            # Lógica compartilhada de conexão TCP
│   ├── string.go          # Cliente para protocolo String
│   ├── json.go            # Cliente para protocolo JSON
│   └── proto.go           # Cliente para Protocol Buffers
└── proto/                 # Definições Protocol Buffers
    ├── client.proto       # Especificação do protocolo
    └── client.pb.go       # Código Go gerado automaticamente
```

## 🔌 Protocolos Suportados

### 1. **Protocolo String** (Porta 8080)
- Mensagens delimitadas por pipe (`|`)
- Formato: `COMANDO|param1=valor1|param2=valor2|FIM`
- Respostas: `OK|campo1|campo2|FIM` ou `ERROR|mensagem|FIM`

### 2. **Protocolo JSON** (Porta 8081)
- Mensagens em formato JSON
- Estruturas tipadas para requisições e respostas
- Campos em lowercase (convenção do servidor)

### 3. **Protocol Buffers** (Porta 8082)
- Serialização binária eficiente
- Mensagens com cabeçalho de 4 bytes (BigEndian) indicando tamanho
- Baseado na especificação `proto3`

## 🧩 Componentes

### `main.go`
**Responsabilidade**: Ponto de entrada da aplicação e orquestração dos testes.

**Funcionalidades**:
- Parse de argumentos de linha de comando (`-proto`, `-host`, `-id`)
- Seleção do cliente apropriado baseado no protocolo escolhido
- Execução da sequência completa de testes (9 passos)
- Gerenciamento de contexto e timeouts

### `client/client.go`
**Responsabilidade**: Definição da interface comum e estruturas de dados.

**Componentes**:
- **Interface `Client`**: Define os 9 métodos que todos os clientes devem implementar
- **Structs de resposta**: `AuthResponse`, `EchoResponse`, `SomaResponse`, `TimestampResponse`, `StatusResponse`, `HistoricoResponse`, `InfoResponse`, `OperacaoInfo`

### `client/base.go`
**Responsabilidade**: Lógica compartilhada de conexão TCP.

**Funcionalidades**:
- Gerenciamento de conexão via `net.Conn`
- `Connect()`: Estabelece conexão TCP com host e porta
- `Disconnect()`: Fecha a conexão de forma segura
- `setDeadline()`: Configura timeout baseado no contexto

### `client/string.go`
**Responsabilidade**: Implementação do protocolo String.

**Características**:
- Utiliza `bufio.Reader` e `bufio.Writer` para I/O eficiente
- `sendAndReceive()`: Envia mensagens delimitadas e processa respostas
- Parsing manual de strings com `strings.Split()`
- Validação de respostas (OK/ERROR)
- Implementa todas as 9 operações do protocolo

**Formato de Mensagens**:
```
Requisição: COMANDO|param=valor|timestamp=RFC3339|FIM
Resposta: OK|campo1=valor1|campo2=valor2|FIM
```

### `client/json.go`
**Responsabilidade**: Implementação do protocolo JSON.

**Características**:
- Utiliza `encoding/json` para serialização/deserialização
- Estruturas tipadas para cada tipo de requisição/resposta
- `sendAndReceive()`: Método genérico com type parameters
- `opRequestHelper()`: Abstração para operações que requerem token
- Campos em lowercase conforme convenção do servidor

**Estruturas Principais**:
- `jsonAuthRequest` / `jsonAuthResponse`
- `jsonOperationRequest` / `jsonOperationResponse`
- Structs específicos de parâmetros: `jsonEchoParams`, `jsonSomaParams`, etc.

### `client/proto.go`
**Responsabilidade**: Implementação do Protocol Buffers.

**Características**:
- Comunicação binária com framing de 4 bytes (BigEndian)
- Utiliza `google.golang.org/protobuf/proto` para marshaling
- `sendAndReceive()`: Envia/recebe mensagens binárias com cabeçalho de tamanho
- `opRequestHelper()`: Validação flexível (ignora campo `Sucesso`, valida por presença de dados)
- Conversão de timezone (UTC → Local) para timestamps
- Parsing de JSON Python-formatted (single quotes, True/False) no histórico

**Peculiaridades**:
- Autenticação valida por presença de token ao invés do campo `Sucesso`
- Histórico retorna JSON em formato Python que precisa ser convertido
- Timestamps UTC são convertidos para timezone local (-03)

### `proto/client.proto`
**Responsabilidade**: Especificação Protocol Buffers.

**Definições**:
- **Requisicao**: Oneof entre `Auth` e `Operacao`
- **Auth**: Contém `aluno_id` e `timestamp`
- **Operacao**: Contém `token`, `nome_operacao`, `parametros` (map), `timestamp`
- **Resposta**: Contém `OperacaoResponse`
- **OperacaoResponse**: Contém `sucesso`, `resultado` (map), `timestamp`

### `proto/client.pb.go`
**Responsabilidade**: Código Go gerado automaticamente pelo compilador `protoc`.

**Observações**:
- **NÃO EDITAR MANUALMENTE**
- Gerado via: `protoc --go_out=. proto/client.proto`
- Contém implementações de serialização/deserialização
- Define structs Go correspondentes às mensagens protobuf

## 📦 Requisitos

- **Go**: 1.21 ou superior
- **Protocol Buffers**: `protoc` e `protoc-gen-go`
- **Dependências**:
  - `google.golang.org/protobuf`

## 🚀 Instalação

1. Clone o repositório:
```bash
git clone https://github.com/GuilhermeGalvao1/SD-trab1.git
cd SD-trab1
```

2. Instale as dependências:
```bash
go mod download
```

3. (Opcional) Regenere o código Protocol Buffers:
```bash
protoc --go_out=. proto/client.proto
```

## 💻 Uso

### Sintaxe Básica
```bash
go run . -proto=[PROTOCOLO] -host=[IP] -id=[MATRICULA]
```

### Parâmetros
- `-proto`: Protocolo a usar (`string`, `json`, ou `proto`) - padrão: `json`
- `-host`: IP do servidor - padrão: `3.88.99.255`
- `-id`: Matrícula do aluno - padrão: `520402`

## 🔧 Operações Disponíveis

A aplicação executa uma sequência de 9 operações em ordem:

### 1. **Connect**
Estabelece conexão TCP com o servidor na porta específica do protocolo.

### 2. **Auth (Autenticação)**
- **Entrada**: ID do aluno
- **Saída**: Token de sessão, nome e matrícula
- **Uso**: Obtém credenciais para operações subsequentes

### 3. **OpEcho**
- **Entrada**: Token + mensagem
- **Saída**: Mensagem original, eco, timestamp, tamanho, hash MD5
- **Uso**: Teste básico de comunicação

### 4. **OpSoma**
- **Entrada**: Token + array de números
- **Saída**: Soma, média, máximo, mínimo, quantidade processada
- **Uso**: Processamento de dados numéricos

### 5. **OpTimestamp**
- **Entrada**: Token
- **Saída**: Timestamp formatado, timezone, informações temporais
- **Uso**: Sincronização de tempo servidor-cliente

### 6. **OpStatus**
- **Entrada**: Token + flag detalhado
- **Saída**: Status do servidor, operações processadas, estatísticas
- **Uso**: Monitoramento do servidor

### 7. **OpHistorico**
- **Entrada**: Token + limite de registros
- **Saída**: Lista de operações executadas, estatísticas
- **Uso**: Auditoria de operações da sessão

### 8. **Info**
- **Entrada**: Token + tipo de informação
- **Saída**: Descrição do servidor, protocolo ativo, capacidades
- **Uso**: Metadados do servidor

### 9. **Logout**
- **Entrada**: Token
- **Saída**: Confirmação de logout
- **Uso**: Encerra sessão e libera recursos

## 📊 Saída Esperada

```
2025/11/16 19:44:43 Iniciando teste com protocolo: proto
[PASSO 1/9] Conectando a 3.88.99.255 (protocolo: proto)... Conectado.
[PASSO 2/9] Autenticando com ID: 520402... Autenticado: GUILHERME GALVÃO SERRA SILVA (520402)
[PASSO 3/9] Testando OpEcho... Echo OK: Hash 4969d59a3b4e2da4a8b446d571e1233e
[PASSO 4/9] Testando OpSoma... Soma OK: Soma=6.00, Média=2.00, Max=3.00, Min=1.00
[PASSO 5/9] Testando OpTimestamp... Timestamp OK: 16/11/2025 19:44:44 (-03)
[PASSO 6/9] Testando OpStatus... Status OK: ATIVO | Ops Processadas: 206
[PASSO 7/9] Testando OpHistorico... Histórico OK: 5 operações retornadas.
[PASSO 8/9] Testando Info... Info OK: Servidor Servidor Protocol Buffers | Protocolo protobuf v3
[PASSO 9/9] Testando Logout... Logout OK.
--- TESTE CONCLUÍDO COM SUCESSO ---
```

## 🐛 Tratamento de Erros

Cada cliente implementa validação robusta:
- **Timeout de contexto**: 60 segundos para toda a sequência
- **Validação de respostas**: Verifica campos obrigatórios e status
- **Reconexão**: Não implementada (conexão única por execução)
- **Logs detalhados**: Indica em qual passo ocorreu a falha

## 👨‍💻 Autor

**Guilherme Galvão Serra Silva**  
Matrícula: 520402  
Universidade Federal do Ceará - UFC  
Disciplina: Sistemas Distribuídos

## 📄 Licença

Este projeto é parte de um trabalho acadêmico da UFC.
