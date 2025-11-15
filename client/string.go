package client

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StringClient struct {
	baseClient
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewStringClient() *StringClient {
	return &StringClient{}
}

func (c *StringClient) Connect(ctx context.Context, host string) error {
	if err := c.baseClient.Connect(ctx, host, "8080"); err != nil {
		return err
	}
	c.reader = bufio.NewReader(c.conn)
	c.writer = bufio.NewWriter(c.conn)
	return nil
}

func (c *StringClient) getTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

func (c *StringClient) sendAndReceive(ctx context.Context, requestBody string) ([]string, error) {
	if err := c.setDeadline(ctx); err != nil {
		return nil, err
	}

	ts := c.getTimestamp()
	msg := fmt.Sprintf("%s|timestamp=%s|FIM\n", requestBody, ts)

	if _, err := c.writer.WriteString(msg); err != nil {
		return nil, fmt.Errorf("falha ao escrever: %w", err)
	}
	if err := c.writer.Flush(); err != nil {
		return nil, fmt.Errorf("falha ao dar flush: %w", err)
	}

	resp, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("falha ao ler resposta: %w", err)
	}

	resp = strings.TrimSpace(resp)
	parts := strings.Split(resp, "|")

	if len(parts) == 0 {
		return nil, fmt.Errorf("resposta vazia ou inválida do servidor")
	}

	if parts[0] == "ERROR" {
		if len(parts) > 1 {
			return nil, fmt.Errorf("erro do servidor: %s", strings.Join(parts[1:], "|"))
		}
		return nil, fmt.Errorf("erro desconhecido do servidor")
	}

	if parts[0] != "OK" {
		return nil, fmt.Errorf("resposta inesperada do servidor: %s", resp)
	}

	if len(parts) > 0 && parts[len(parts)-1] == "FIM" {
		parts = parts[:len(parts)-1]
	}

	return parts[1:], nil
}

func splitVal(kv string) string {
	s := strings.SplitN(kv, "=", 2)
	if len(s) == 2 {
		return s[1]
	}
	return kv
}

func (c *StringClient) Auth(ctx context.Context, alunoID string) (*AuthResponse, error) {
	reqBody := "AUTH|aluno_id=" + alunoID
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("resposta de AUTH incompleta. Esperado 3 campos, recebido %d", len(parts))
	}

	return &AuthResponse{
		Token:     splitVal(parts[0]),
		Nome:      splitVal(parts[1]),
		Matricula: splitVal(parts[2]),
	}, nil
}

func (c *StringClient) OpEcho(ctx context.Context, token, msg string) (*EchoResponse, error) {
	reqBody := fmt.Sprintf("OP|operacao=echo|mensagem=%s|token=%s", msg, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 5 {
		return nil, fmt.Errorf("resposta de ECHO incompleta. Esperado 5 campos, recebido %d", len(parts))
	}

	tamanho, _ := strconv.Atoi(splitVal(parts[3]))
	return &EchoResponse{
		MensagemOriginal: splitVal(parts[0]),
		Eco:              splitVal(parts[1]),
		Timestamp:        splitVal(parts[2]),
		Tamanho:          tamanho,
		HashMD5:          splitVal(parts[4]),
	}, nil
}

func (c *StringClient) OpSoma(ctx context.Context, token string, numeros []string) (*SomaResponse, error) {
	numListStr := strings.Join(numeros, ",")

	reqBody := fmt.Sprintf(`OP|operacao=soma|nums=%s|token=%s`, numListStr, token)

	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 6 {
		return nil, fmt.Errorf("resposta de SOMA incompleta. Esperado 6 campos, recebido %d", len(parts))
	}

	soma, _ := strconv.ParseFloat(splitVal(parts[2]), 64)
	media, _ := strconv.ParseFloat(splitVal(parts[3]), 64)
	maximo, _ := strconv.ParseFloat(splitVal(parts[4]), 64)
	minimo, _ := strconv.ParseFloat(splitVal(parts[5]), 64)
	count, _ := strconv.Atoi(splitVal(parts[1]))

	return &SomaResponse{
		Soma:               soma,
		Media:              media,
		Maximo:             maximo,
		Minimo:             minimo,
		NumerosProcessados: count,
	}, nil
}

func (c *StringClient) OpTimestamp(ctx context.Context, token string) (*TimestampResponse, error) {
	reqBody := "OP|operacao=timestamp|token=" + token
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("resposta de TIMESTAMP incompleta. Esperado 3 campos, recebido %d", len(parts))
	}

	return &TimestampResponse{
		TimestampFormatado:   splitVal(parts[0]),
		Timezone:             splitVal(parts[1]),
		InformacoesTemporais: splitVal(parts[2]),
	}, nil
}

func (c *StringClient) OpStatus(ctx context.Context, token string, detalhado bool) (*StatusResponse, error) {
	reqBody := fmt.Sprintf("OP|operacao=status|detalhado=%t|token=%s", detalhado, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 2 {
		return nil, fmt.Errorf("resposta de STATUS incompleta. Esperado 2+ campos, recebido %d", len(parts))
	}

	resp := &StatusResponse{
		Status: splitVal(parts[0]),
	}

	if detalhado && len(parts) > 2 {
		opCount, _ := strconv.Atoi(splitVal(parts[2]))
		resp.OperacoesProcessadas = opCount
		resp.Estatisticas = map[string]any{
			"raw_stats": strings.Join(parts[2:], "|"),
		}
	} else if len(parts) > 1 {
		opCount, _ := strconv.Atoi(splitVal(parts[1]))
		resp.OperacoesProcessadas = opCount
	}

	return resp, nil
}

func (c *StringClient) OpHistorico(ctx context.Context, token string, limite int) (*HistoricoResponse, error) {
	reqBodyBase := "OP|operacao=historico"
	if limite > 0 {
		reqBodyBase = fmt.Sprintf("%s|limite=%d", reqBodyBase, limite)
	}
	reqBody := reqBodyBase + "|token=" + token

	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 2 {
		return nil, fmt.Errorf("resposta de HISTORICO incompleta. Esperado 2 campos, recebido %d", len(parts))
	}

	opListStr := splitVal(parts[0])
	opStrings := strings.Split(opListStr, ",")
	operacoes := make([]OperacaoInfo, len(opStrings))
	for i, opStr := range opStrings {
		operacoes[i] = OperacaoInfo{Comando: opStr, Timestamp: "N/A", Sucesso: true}
	}

	return &HistoricoResponse{
		Operacoes: operacoes,
		Estatisticas: map[string]any{
			"raw_stats": splitVal(parts[1]),
		},
	}, nil
}

func (c *StringClient) Info(ctx context.Context, token, tipo string) (*InfoResponse, error) {
	reqBody := fmt.Sprintf("INFO|tipo=%s|token=%s", tipo, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("resposta de INFO incompleta. Esperado 3 campos, recebido %d", len(parts))
	}

	return &InfoResponse{
		DescricaoServidor: splitVal(parts[0]),
		ProtocoloAtivo:    splitVal(parts[1]),
		Capacidades:       strings.Split(splitVal(parts[2]), ","),
	}, nil
}

func (c *StringClient) Logout(ctx context.Context, token string) error {
	reqBody := "LOGOUT|token=" + token
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return err
	}

	if len(parts) < 1 {
		return fmt.Errorf("resposta de LOGOUT inválida")
	}

	fmt.Printf("[Servidor String]: %s\n", splitVal(parts[0]))
	return nil
}
