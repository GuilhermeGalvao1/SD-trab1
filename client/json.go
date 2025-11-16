package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

type jsonOperationRequest struct {
	Tipo       string `json:"tipo"`
	Token      string `json:"token"`
	Operacao   string `json:"operacao"`
	Parametros any    `json:"parametros"`
	Timestamp  string `json:"timestamp"`
}
type jsonAuthRequest struct {
	Tipo      string `json:"tipo"`
	AlunoID   string `json:"aluno_id"`
	Timestamp string `json:"timestamp"`
}
type jsonEchoParams struct {
	Mensagem string `json:"mensagem"`
}
type jsonSomaParams struct {
	Numeros []int `json:"numeros"`
}
type jsonStatusParams struct {
	Detalhado bool `json:"detalhado"`
}
type jsonHistoricoParams struct {
	Limite int `json:"limite"`
}

type jsonBaseResponse struct {
	Sucesso   bool   `json:"sucesso"`
	Erro      string `json:"erro"`
	Mensagem  string `json:"mensagem"`
	Timestamp string `json:"timestamp"`
}

type jsonAuthResponse struct {
	Sucesso    bool   `json:"sucesso"`
	Erro       string `json:"erro"`
	Token      string `json:"token"`
	DadosAluno struct {
		Nome      string `json:"nome"`
		Matricula string `json:"matricula"`
	} `json:"dados_aluno"`
	Mensagem  string `json:"mensagem"`
	Timestamp string `json:"timestamp"`
}

type jsonOperationResponse struct {
	Sucesso   bool           `json:"sucesso"`
	Erro      string         `json:"erro"`
	Resultado map[string]any `json:"resultado"`
	Mensagem  string         `json:"mensagem"`
	Timestamp string         `json:"timestamp"`
}

type JsonClient struct {
	baseClient
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewJsonClient() *JsonClient {
	return &JsonClient{}
}

func (c *JsonClient) Connect(ctx context.Context, host string) error {
	if err := c.baseClient.Connect(ctx, host, "8081"); err != nil {
		return err
	}
	c.encoder = json.NewEncoder(c.conn)
	c.decoder = json.NewDecoder(c.conn)
	return nil
}

func (c *JsonClient) sendAndReceive(ctx context.Context, request any, responseDest any) error {
	if err := c.setDeadline(ctx); err != nil {
		return err
	}
	debugBytes, _ := json.Marshal(request)
	log.Printf("[DEBUG JSON] Enviando: %s\n", string(debugBytes))

	if err := c.encoder.Encode(request); err != nil {
		return fmt.Errorf("falha ao enviar JSON: %w", err)
	}
	if err := c.decoder.Decode(responseDest); err != nil {
		return fmt.Errorf("falha ao ler/decodificar resposta JSON: %w", err)
	}
	debugRespBytes, _ := json.Marshal(responseDest)
	log.Printf("[DEBUG JSON] Recebido: %s\n", string(debugRespBytes))
	return nil
}

func (c *JsonClient) Auth(ctx context.Context, alunoID string) (*AuthResponse, error) {
	req := jsonAuthRequest{
		Tipo:      "autenticar",
		AlunoID:   alunoID,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	var resp jsonAuthResponse
	if err := c.sendAndReceive(ctx, req, &resp); err != nil {
		return nil, err
	}

	if !resp.Sucesso {
		if resp.Erro != "" {
			return nil, fmt.Errorf("falha na autenticação: %s", resp.Erro)
		}
		if resp.Mensagem != "" {
			return nil, fmt.Errorf("falha na autenticação: %s", resp.Mensagem)
		}
		return nil, fmt.Errorf("falha na autenticação: (status não OK e sem mensagem de erro)")
	}

	return &AuthResponse{
		Token:     resp.Token,
		Nome:      resp.DadosAluno.Nome,
		Matricula: alunoID,
	}, nil
}

func (c *JsonClient) opRequestHelper(ctx context.Context, token, opName string, params any) (*jsonOperationResponse, error) {
	req := jsonOperationRequest{
		Tipo:       "operacao",
		Token:      token,
		Operacao:   opName,
		Parametros: params,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	var resp jsonOperationResponse
	if err := c.sendAndReceive(ctx, req, &resp); err != nil {
		return nil, err
	}

	if !resp.Sucesso {
		if resp.Erro != "" {
			return nil, fmt.Errorf("erro na operação '%s': %s", opName, resp.Erro)
		}
		if resp.Mensagem != "" {
			return nil, fmt.Errorf("erro na operação '%s': %s", opName, resp.Mensagem)
		}
		return nil, fmt.Errorf("erro na operação '%s': (status não OK e sem mensagem de erro)", opName)
	}

	return &resp, nil
}

func (c *JsonClient) OpEcho(ctx context.Context, token, msg string) (*EchoResponse, error) {
	params := jsonEchoParams{Mensagem: msg}
	resp, err := c.opRequestHelper(ctx, token, "echo", params)
	if err != nil {
		return nil, err
	}

	r := resp.Resultado
	return &EchoResponse{
		MensagemOriginal: r["mensagem_original"].(string),
		Eco:              r["mensagem_eco"].(string),
		Timestamp:        r["timestamp_servidor"].(string),
		Tamanho:          int(r["tamanho_mensagem"].(float64)),
		HashMD5:          r["hash_md5"].(string),
	}, nil
}

func (c *JsonClient) OpSoma(ctx context.Context, token string, numeros []string) (*SomaResponse, error) {
	var numsInt []int
	for _, s := range numeros {
		n, err := strconv.Atoi(s)
		if err == nil {
			numsInt = append(numsInt, n)
		} else {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				numsInt = append(numsInt, int(f))
			}
		}
	}

	params := jsonSomaParams{Numeros: numsInt}
	resp, err := c.opRequestHelper(ctx, token, "soma", params)
	if err != nil {
		return nil, err
	}

	r := resp.Resultado
	return &SomaResponse{
		Soma:               r["soma"].(float64),
		Media:              r["media"].(float64),
		Maximo:             r["maximo"].(float64),
		Minimo:             r["minimo"].(float64),
		NumerosProcessados: int(r["quantidade"].(float64)),
	}, nil
}

func (c *JsonClient) OpTimestamp(ctx context.Context, token string) (*TimestampResponse, error) {
	params := make(map[string]any)
	resp, err := c.opRequestHelper(ctx, token, "timestamp", params)
	if err != nil {
		return nil, err
	}

	r := resp.Resultado
	return &TimestampResponse{
		TimestampFormatado:   r["timestamp_formatado"].(string),
		Timezone:             "N/A",
		InformacoesTemporais: r["timestamp_iso"].(string),
	}, nil
}

func (c *JsonClient) OpStatus(ctx context.Context, token string, detalhado bool) (*StatusResponse, error) {
	params := jsonStatusParams{Detalhado: detalhado}
	resp, err := c.opRequestHelper(ctx, token, "status", params)
	if err != nil {
		return nil, err
	}

	r := resp.Resultado
	statsMap, _ := r["estatisticas_banco"].(map[string]any)

	return &StatusResponse{
		Status:               r["status"].(string),
		OperacoesProcessadas: int(r["operacoes_processadas"].(float64)),
		Estatisticas:         statsMap,
	}, nil
}

func (c *JsonClient) OpHistorico(ctx context.Context, token string, limite int) (*HistoricoResponse, error) {
	params := jsonHistoricoParams{Limite: limite}
	resp, err := c.opRequestHelper(ctx, token, "historico", params)
	if err != nil {
		return nil, err
	}

	r := resp.Resultado
	var operacoes []OperacaoInfo

	if ops, ok := r["historico"].([]any); ok {
		for _, opAny := range ops {
			if opMap, ok := opAny.(map[string]any); ok {
				var opNome, opTs string
				var opSuc bool
				if v, ok := opMap["operacao"].(string); ok {
					opNome = v
				}
				if v, ok := opMap["timestamp"].(string); ok {
					opTs = v
				}
				if v, ok := opMap["sucesso"].(bool); ok {
					opSuc = v
				}

				operacoes = append(operacoes, OperacaoInfo{
					Comando:   opNome,
					Timestamp: opTs,
					Sucesso:   opSuc,
				})
			}
		}
	}

	statsMap, _ := r["estatisticas"].(map[string]any)

	return &HistoricoResponse{
		Operacoes:    operacoes,
		Estatisticas: statsMap,
	}, nil
}

func (c *JsonClient) Info(ctx context.Context, token, tipo string) (*InfoResponse, error) {
	req := map[string]any{
		"tipo":      "info",
		"token":     token,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	var resp jsonOperationResponse
	if err := c.sendAndReceive(ctx, req, &resp); err != nil {
		log.Printf("Falha no 'tipo'=\"info\", tentando 'operacao'=\"info\"...")
		return c.InfoAsOperation(ctx, token, tipo)
	}

	r := resp.Resultado
	var desc, proto string

	if d, ok := r["nome"].(string); ok {
		desc = d
	}
	if p, ok := r["versao"].(string); ok {
		proto = p
	}

	return &InfoResponse{
		DescricaoServidor: desc,
		ProtocoloAtivo:    proto,
	}, nil
}

func (c *JsonClient) InfoAsOperation(ctx context.Context, token, tipo string) (*InfoResponse, error) {
	params := map[string]any{"tipo": tipo}
	resp, err := c.opRequestHelper(ctx, token, "info", params)
	if err != nil {
		return nil, err
	}

	r := resp.Resultado
	var desc, proto string
	if d, ok := r["nome"].(string); ok {
		desc = d
	}
	if p, ok := r["versao"].(string); ok {
		proto = p
	}

	return &InfoResponse{
		DescricaoServidor: desc,
		ProtocoloAtivo:    proto,
	}, nil
}

func (c *JsonClient) Logout(ctx context.Context, token string) error {
	req := map[string]any{
		"tipo":      "logout",
		"token":     token,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	var resp jsonBaseResponse
	if err := c.sendAndReceive(ctx, req, &resp); err != nil {
		return err
	}

	if !resp.Sucesso {
		if resp.Erro != "" {
			return fmt.Errorf("falha no logout: %s", resp.Erro)
		}
		return fmt.Errorf("falha no logout: (status não OK e sem mensagem de erro)")
	}

	return nil
}
