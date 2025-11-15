package client

import "context"

type AuthResponse struct {
	Token     string
	Nome      string
	Matricula string
}

type EchoResponse struct {
	MensagemOriginal string
	Eco              string
	Timestamp        string
	Tamanho          int
	HashMD5          string
}

type SomaResponse struct {
	Soma               float64
	Media              float64
	Maximo             float64
	Minimo             float64
	NumerosProcessados int
}

type TimestampResponse struct {
	TimestampFormatado   string
	Timezone             string
	InformacoesTemporais string
}

type StatusResponse struct {
	Status               string
	OperacoesProcessadas int
	Estatisticas         map[string]any
}

type OperacaoInfo struct {
	Comando   string
	Timestamp string
	Sucesso   bool
}

type HistoricoResponse struct {
	Operacoes    []OperacaoInfo
	Estatisticas map[string]any
}

type InfoResponse struct {
	DescricaoServidor string
	ProtocoloAtivo    string
	Capacidades       []string
}

type Client interface {
	Connect(ctx context.Context, host string) error

	Disconnect() error

	Auth(ctx context.Context, alunoID string) (*AuthResponse, error)

	OpEcho(ctx context.Context, token, msg string) (*EchoResponse, error)

	OpSoma(ctx context.Context, token string, numeros []string) (*SomaResponse, error)

	OpTimestamp(ctx context.Context, token string) (*TimestampResponse, error)

	OpStatus(ctx context.Context, token string, detalhado bool) (*StatusResponse, error)

	OpHistorico(ctx context.Context, token string, limite int) (*HistoricoResponse, error)

	Info(ctx context.Context, token, tipo string) (*InfoResponse, error)

	Logout(ctx context.Context, token string) error
}
