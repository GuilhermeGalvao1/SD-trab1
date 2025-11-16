package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	pb "github.com/GuilhermeGalvao1/SD-trab1/proto"

	"google.golang.org/protobuf/proto"
)

type ProtoClient struct {
	baseClient
}

func NewProtoClient() *ProtoClient {
	return &ProtoClient{}
}

func (c *ProtoClient) Connect(ctx context.Context, host string) error {
	return c.baseClient.Connect(ctx, host, "8082")
}

func (c *ProtoClient) sendAndReceive(ctx context.Context, req *pb.Requisicao) (*pb.Resposta, error) {
	if err := c.setDeadline(ctx); err != nil {
		return nil, err
	}

	payload, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("proto: falha ao serializar requisição: %w", err)
	}

	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))

	if _, err := c.conn.Write(append(hdr[:], payload...)); err != nil {
		return nil, fmt.Errorf("proto: falha ao enviar mensagem: %w", err)
	}

	if _, err := io.ReadFull(c.conn, hdr[:]); err != nil {
		return nil, fmt.Errorf("proto: falha ao ler cabeçalho da resposta: %w", err)
	}
	size := binary.BigEndian.Uint32(hdr[:])

	respPayload := make([]byte, size)
	if _, err := io.ReadFull(c.conn, respPayload); err != nil {
		return nil, fmt.Errorf("proto: falha ao ler payload da resposta: %w", err)
	}

	var resp pb.Resposta
	if err := proto.Unmarshal(respPayload, &resp); err != nil {
		return nil, fmt.Errorf("proto: falha ao desserializar resposta: %w", err)
	}

	return &resp, nil
}

func (c *ProtoClient) opRequestHelper(ctx context.Context, token, opName string, params map[string]string) (map[string]string, error) {
	req := &pb.Requisicao{
		Conteudo: &pb.Requisicao_Operacao{
			Operacao: &pb.Operacao{
				Token:        token,
				NomeOperacao: opName,
				Parametros:   params,
				Timestamp:    time.Now().UTC().Format(time.RFC3339Nano),
			},
		},
	}

	resp, err := c.sendAndReceive(ctx, req)
	if err != nil {
		return nil, err
	}

	opResp := resp.GetOperacao()
	if opResp == nil {
		return nil, fmt.Errorf("proto: resposta de operação inválida (nula)")
	}

	if errMsg, ok := opResp.Resultado["erro"]; ok && errMsg != "" {
		return nil, fmt.Errorf("proto: erro do servidor: %s", errMsg)
	}
	if errMsg, ok := opResp.Resultado["mensagem"]; ok && errMsg != "" {
		if opResp.Resultado["erro"] != "" || len(opResp.Resultado) == 1 {
			return nil, fmt.Errorf("proto: erro do servidor: %s", errMsg)
		}
	}
	if errMsg, ok := opResp.Resultado["error"]; ok && errMsg != "" {
		return nil, fmt.Errorf("proto: erro do servidor: %s", errMsg)
	}

	if len(opResp.Resultado) == 0 {
		return nil, fmt.Errorf("proto: operação falhou - sem dados retornados")
	}

	return opResp.Resultado, nil
}

func (c *ProtoClient) Auth(ctx context.Context, alunoID string) (*AuthResponse, error) {
	req := &pb.Requisicao{
		Conteudo: &pb.Requisicao_Auth{
			Auth: &pb.Auth{
				AlunoId:   alunoID,
				Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			},
		},
	}

	resp, err := c.sendAndReceive(ctx, req)
	if err != nil {
		return nil, err
	}

	opResp := resp.GetOperacao()
	if opResp == nil {
		return nil, fmt.Errorf("proto: resposta de autenticação inválida (nula)")
	}

	r := opResp.Resultado

	if r["token"] == "" {
		for key, value := range r {
			fmt.Printf("Campo '%s': %s\n", key, value)
		}

		if errMsg, ok := r["erro"]; ok {
			return nil, fmt.Errorf("proto: falha na autenticação: %s", errMsg)
		}
		if errMsg, ok := r["mensagem"]; ok {
			return nil, fmt.Errorf("proto: falha na autenticação: %s", errMsg)
		}
		if errMsg, ok := r["error"]; ok {
			return nil, fmt.Errorf("proto: falha na autenticação: %s", errMsg)
		}

		return nil, fmt.Errorf("proto: falha na autenticação - sem token retornado")
	}

	return &AuthResponse{
		Token:     r["token"],
		Nome:      r["nome"],
		Matricula: r["matricula"],
	}, nil
}

func (c *ProtoClient) OpEcho(ctx context.Context, token, msg string) (*EchoResponse, error) {
	params := map[string]string{"mensagem": msg}
	r, err := c.opRequestHelper(ctx, token, "echo", params)
	if err != nil {
		return nil, err
	}

	t, _ := strconv.Atoi(r["tamanho_mensagem"])
	return &EchoResponse{
		MensagemOriginal: r["mensagem_original"],
		Eco:              r["mensagem_eco"],
		Timestamp:        r["timestamp_servidor"],
		Tamanho:          t,
		HashMD5:          r["hash_md5"],
	}, nil
}

func (c *ProtoClient) OpSoma(ctx context.Context, token string, numeros []string) (*SomaResponse, error) {
	numListStr := strings.Join(numeros, ",")

	params := map[string]string{"nums": numListStr}

	r, err := c.opRequestHelper(ctx, token, "soma", params)
	if err != nil {
		return nil, err
	}

	soma, _ := strconv.ParseFloat(r["soma"], 64)
	media, _ := strconv.ParseFloat(r["media"], 64)
	max, _ := strconv.ParseFloat(r["maximo"], 64)
	min, _ := strconv.ParseFloat(r["minimo"], 64)
	count, _ := strconv.Atoi(r["quantidade"])

	return &SomaResponse{
		Soma:               soma,
		Media:              media,
		Maximo:             max,
		Minimo:             min,
		NumerosProcessados: count,
	}, nil
}

func (c *ProtoClient) OpTimestamp(ctx context.Context, token string) (*TimestampResponse, error) {
	params := map[string]string{}
	r, err := c.opRequestHelper(ctx, token, "timestamp", params)
	if err != nil {
		return nil, err
	}

	timestampFormatado := r["timestamp_formatado"]
	timestampISO := r["timestamp_iso"]
	tz := "Local"

	if timestampISO != "" {
		t, err := time.Parse("2006-01-02T15:04:05.999999", timestampISO)
		if err == nil {
			localTime := t.UTC().In(time.Local)
			timestampFormatado = localTime.Format("02/01/2006 15:04:05")

			zoneName, _ := localTime.Zone()
			tz = zoneName
		}
	}

	return &TimestampResponse{
		TimestampFormatado:   timestampFormatado,
		Timezone:             tz,
		InformacoesTemporais: timestampISO,
	}, nil
}

func (c *ProtoClient) OpStatus(ctx context.Context, token string, detalhado bool) (*StatusResponse, error) {
	params := map[string]string{"detalhado": strconv.FormatBool(detalhado)}
	r, err := c.opRequestHelper(ctx, token, "status", params)
	if err != nil {
		return nil, err
	}

	opCount, _ := strconv.Atoi(r["operacoes_processadas"])

	statsMap := make(map[string]any)
	if statsStr, ok := r["estatisticas_banco"]; ok {
		_ = json.Unmarshal([]byte(statsStr), &statsMap)
	}

	return &StatusResponse{
		Status:               r["status"],
		OperacoesProcessadas: opCount,
		Estatisticas:         statsMap,
	}, nil
}

func (c *ProtoClient) OpHistorico(ctx context.Context, token string, limite int) (*HistoricoResponse, error) {
	params := map[string]string{"limite": strconv.Itoa(limite)}
	r, err := c.opRequestHelper(ctx, token, "historico", params)
	if err != nil {
		return nil, err
	}

	var operacoes []OperacaoInfo
	if histStr, ok := r["historico"]; ok && histStr != "" {
		histStr = strings.ReplaceAll(histStr, "'", "\"")
		histStr = strings.ReplaceAll(histStr, "True", "true")
		histStr = strings.ReplaceAll(histStr, "False", "false")

		var rawOps []map[string]any
		if err := json.Unmarshal([]byte(histStr), &rawOps); err == nil {
			operacoes = make([]OperacaoInfo, len(rawOps))
			for i, op := range rawOps {
				operacoes[i] = OperacaoInfo{
					Comando:   fmt.Sprintf("%v", op["operacao"]),
					Timestamp: fmt.Sprintf("%v", op["timestamp"]),
					Sucesso:   fmt.Sprintf("%v", op["sucesso"]) == "true",
				}
			}
		}
	}

	statsMap := make(map[string]any)
	if statsStr, ok := r["estatisticas"]; ok && statsStr != "" {
		statsStr = strings.ReplaceAll(statsStr, "'", "\"")
		_ = json.Unmarshal([]byte(statsStr), &statsMap)
	}

	return &HistoricoResponse{
		Operacoes:    operacoes,
		Estatisticas: statsMap,
	}, nil
}

func (c *ProtoClient) Info(ctx context.Context, token, tipo string) (*InfoResponse, error) {
	params := map[string]string{"tipo": tipo}
	r, err := c.opRequestHelper(ctx, token, "info", params)
	if err != nil {
		return &InfoResponse{
			DescricaoServidor: "Servidor Protocol Buffers",
			ProtocoloAtivo:    "protobuf v3",
			Capacidades:       []string{"auth", "echo", "soma", "timestamp", "status", "historico", "logout"},
		}, nil
	}

	var capacidades []string
	if capStr, ok := r["capacidades"]; ok && capStr != "" {
		capacidades = strings.Split(capStr, ",")
	}

	return &InfoResponse{
		DescricaoServidor: r["nome"],
		ProtocoloAtivo:    r["versao"],
		Capacidades:       capacidades,
	}, nil
}

func (c *ProtoClient) Logout(ctx context.Context, token string) error {
	_, err := c.opRequestHelper(ctx, token, "logout", map[string]string{})
	if err != nil && err.Error() != "proto: resposta de operação inválida (nula)" {
		return err
	}
	return nil
}
