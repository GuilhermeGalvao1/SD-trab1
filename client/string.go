package client

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type baseClient struct {
	conn net.Conn
}

func (c *baseClient) Connect(ctx context.Context, host, port string) error {
	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return fmt.Errorf("falha ao conectar (%s:%s): %w", host, port, err)
	}
	c.conn = conn
	return nil
}

func (c *baseClient) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *baseClient) setDeadline(ctx context.Context) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}
	return c.conn.SetDeadline(deadline)
}

type StringClient struct {
	baseClient
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewStringClient() *StringClient {
	return &StringClient{}
}

func (c *StringClient) Connect(ctx context.Context, host string) error {
	if err := c.baseClient.Connect(ctx, host, "8081"); err != nil {
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

	msg := fmt.Sprintf("%s||%s\n", requestBody, c.getTimestamp())

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

	if parts[0] == "ERRO" {
		if len(parts) > 1 {
			return nil, fmt.Errorf("erro do servidor: %s", strings.Join(parts[1:], "|"))
		}
		return nil, fmt.Errorf("erro desconhecido do servidor")
	}

	if parts[0] == "OK" {
		return nil, fmt.Errorf("resposta inesperada do servidor (não OK): %s", resp)
	}

	return parts[1:], nil
}

func (c *StringClient) Auth(ctx context.Context, alunoID string) (*AuthResponse, error) {
	reqBody := fmt.Sprintf("AUTH|%s", alunoID)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("resposta de AUTH incompleta. Esperado 3 campos, recebido %d", len(parts))
	}

	return &AuthResponse{
		Token:     parts[0],
		Nome:      parts[1],
		Matricula: parts[2],
	}, nil
}

func (c *StringClient) OpEcho(ctx context.Context, token, msg string) (*EchoResponse, error) {
	param1 := fmt.Sprintf("echo mensagem %s", msg)
	reqBody := fmt.Sprintf("OP|%s|%s", param1, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 5 {
		return nil, fmt.Errorf("resposta de ECHO incompleta. Esperado 5 campos, recebido %d", len(parts))
	}

	tamanho, _ := strconv.Atoi(parts[3])
	return &EchoResponse{
		MensagemOriginal: parts[0],
		Eco:              parts[1],
		Timestamp:        parts[2],
		Tamanho:          tamanho,
		HashMD5:          parts[4],
	}, nil
}

func (c *StringClient) OpSoma(ctx context.Context, token string, numeros []float64) (*SomaResponse, error) {
	numStrs := make([]string, len(numeros))
	for i, n := range numeros {
		numStrs[i] = strconv.FormatFloat(n, 'f', -1, 64)
	}
	numListStr := strings.Join(numStrs, ",")

	param1 := fmt.Sprintf("soma numeros %s", numListStr)
	reqBody := fmt.Sprintf("OP|%s|%s", param1, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 5 {
		return nil, fmt.Errorf("resposta de SOMA incompleta. Esperado 5 campos, recebido %d", len(parts))
	}

	soma, _ := strconv.ParseFloat(parts[0], 64)
	media, _ := strconv.ParseFloat(parts[1], 64)
	maximo, _ := strconv.ParseFloat(parts[2], 64)
	minimo, _ := strconv.ParseFloat(parts[3], 64)
	count, _ := strconv.Atoi(parts[4])

	return &SomaResponse{
		Soma:               soma,
		Media:              media,
		Maximo:             maximo,
		Minimo:             minimo,
		NumerosProcessados: count,
	}, nil
}

func (c *StringClient) OpTimestamp(ctx context.Context, token string) (*TimestampResponse, error) {
	param1 := "timestamp"
	reqBody := fmt.Sprintf("OP|%s|%s", param1, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("resposta de TIMESTAMP incompleta. Esperado 3 campos, recebido %d", len(parts))
	}

	return &TimestampResponse{
		TimestampFormatado: parts[0],
		Timezone:           parts[1],
		InfoTemporais:      parts[2],
	}, nil
}

func (c *StringClient) OpStatus(ctx context.Context, token string, detalhado bool) (*StatusResponse, error) {
	param1 := "status"
	if detalhado {
		param1 = "status detalhado"
	}

	reqBody := fmt.Sprintf("OP|%s|%s", param1, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 2 {
		return nil, fmt.Errorf("resposta de STATUS incompleta. Esperado 2+ campos, recebido %d", len(parts))
	}

	opCount, _ := strconv.Atoi(parts[1])
	resp := &StatusResponse{
		Status:               parts[0],
		OperaçõesProcessadas: opCount,
	}

	if detalhado && len(parts) > 2 {
		resp.Estatisticas = map[string]any{
			"raw_stats": strings.Join(parts[2:], "|"),
		}
	}

	return resp, nil
}

func (c *StringClient) OpHistorico(ctx context.Context, token string, limite int) (*HistoricoResponse, error) {
	param1 := "historico"
	if limite > 0 {
		param1 = fmt.Sprintf("historico limite %d", limite)
	}

	reqBody := fmt.Sprintf("OP|%s|%s", param1, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 2 {
		return nil, fmt.Errorf("resposta de HISTORICO incompleta. Esperado 2 campos, recebido %d", len(parts))
	}

	opStrings := strings.Split(parts[0], ",")
	operacoes := make([]OperacaoInfo, len(opStrings))
	for i, opStr := range opStrings {
		operacoes[i] = OperacaoInfo{Comando: opStr, Timestamp: "N/A", Sucesso: true}
	}

	return &HistoricoResponse{
		Operacoes: operacoes,
		Estatisticas: map[string]any{
			"raw_stats": parts[1],
		},
	}, nil
}

func (c *StringClient) Info(ctx context.Context, token, tipo string) (*InfoResponse, error) {
	reqBody := fmt.Sprintf("INFO|%s|%s", tipo, token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("resposta de INFO incompleta. Esperado 3 campos, recebido %d", len(parts))
	}

	return &InfoResponse{
		DescricaoServidor: parts[0],
		ProtocoloAtivo:    parts[1],
		Capacidades:       strings.Split(parts[2], ","),
	}, nil
}

func (c *StringClient) Logout(ctx context.Context, token string) error {
	reqBody := fmt.Sprintf("LOGOUT|%s", token)
	parts, err := c.sendAndReceive(ctx, reqBody)
	if err != nil {
		return err
	}

	if len(parts) < 1 {
		return fmt.Errorf("resposta de LOGOUT inválida")
	}

	fmt.Printf("[Servidor String]: %s\n", parts[0])
	return nil
}
