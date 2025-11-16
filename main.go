package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/GuilhermeGalvao1/SD-trab1/client"
)

func main() {
	proto := flag.String("proto", "json", "Protocolo a ser usado (string, json, proto)")
	host := flag.String("host", "3.88.99.255", "IP do servidor (fornecido pelo professor)")
	id := flag.String("id", "520402", "Matrícula do aluno para teste (ex: 202301001)")
	flag.Parse()

	if *host == "" {
		log.Fatalf("Erro: O IP do host é obrigatório. Use -host=[IP_PUBLICO]")
	}

	var c client.Client
	log.Printf("Iniciando teste com protocolo: %s\n", *proto)

	switch *proto {
	case "string":
		c = client.NewStringClient()
	case "json":
		c = client.NewJsonClient()
	case "proto":
		c = client.NewProtoClient()
	default:
		log.Fatalf("Protocolo '%s' desconhecido. Use 'string', 'json' ou 'proto'.", *proto)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := runTestSequence(ctx, c, *host, *id, *proto)
	if err != nil {
		log.Fatalf("\n--- TESTE FALHOU ---\n%v\n--------------------", err)
	}

	log.Println("\n--- TESTE CONCLUÍDO COM SUCESSO ---")
}
func runTestSequence(ctx context.Context, c client.Client, host, alunoID, protoName string) error {
	var token string

	log.Printf("[PASSO 1/9] Conectando a %s (protocolo: %s)...", host, protoName)
	if err := c.Connect(ctx, host); err != nil {
		return fmt.Errorf("falha ao conectar: %w", err)
	}
	defer c.Disconnect()
	log.Println("... Conectado.")

	log.Printf("[PASSO 2/9] Autenticando com ID: %s...", alunoID)
	authResp, err := c.Auth(ctx, alunoID)
	if err != nil {
		return fmt.Errorf("falha no Auth: %w", err)
	}
	token = authResp.Token
	log.Printf("... Autenticado: %s (%s)", authResp.Nome, authResp.Matricula)

	log.Println("[PASSO 3/9] Testando OpEcho...")
	echoMsg := "Ola-Mundo-SD-Go"
	echoResp, err := c.OpEcho(ctx, token, echoMsg)
	if err != nil {
		return fmt.Errorf("falha no OpEcho: %w", err)
	}
	log.Printf("... Echo OK: Hash %s", echoResp.HashMD5)

	log.Println("[PASSO 4/9] Testando OpSoma...")
	numeros := []string{"1", "2", "3"}
	somaResp, err := c.OpSoma(ctx, token, numeros)
	if err != nil {
		return fmt.Errorf("falha no OpSoma: %w", err)
	}
	log.Printf("... Soma OK: Soma=%.2f, Média=%.2f, Max=%.2f, Min=%.2f",
		somaResp.Soma, somaResp.Media, somaResp.Maximo, somaResp.Minimo)

	log.Println("[PASSO 5/9] Testando OpTimestamp...")
	tsResp, err := c.OpTimestamp(ctx, token)
	if err != nil {
		return fmt.Errorf("falha no OpTimestamp: %w", err)
	}
	log.Printf("... Timestamp OK: %s (%s)", tsResp.TimestampFormatado, tsResp.Timezone)

	log.Println("[PASSO 6/9] Testando OpStatus (detalhado)...")
	statusResp, err := c.OpStatus(ctx, token, true)
	if err != nil {
		return fmt.Errorf("falha no OpStatus: %w", err)
	}
	log.Printf("... Status OK: %s | Ops Processadas: %d",
		statusResp.Status, statusResp.OperacoesProcessadas)
	if statusResp.Estatisticas != nil {
		log.Printf("... Estatísticas do Status: %v", statusResp.Estatisticas)
	}

	log.Println("[PASSO 7/9] Testando OpHistorico (limite 5)...")
	histResp, err := c.OpHistorico(ctx, token, 5)
	if err != nil {
		return fmt.Errorf("falha no OpHistorico: %w", err)
	}
	log.Printf("... Histórico OK: %d operações retornadas.", len(histResp.Operacoes))

	log.Println("[PASSO 8/9] Testando Info (detalhado)...")
	infoResp, err := c.Info(ctx, token, "detalhado")
	if err != nil {
		return fmt.Errorf("falha no Info: %w", err)
	}
	log.Printf("... Info OK: Servidor %s | Protocolo %s",
		infoResp.DescricaoServidor, infoResp.ProtocoloAtivo)

	log.Println("[PASSO 9/9] Testando Logout...")
	if err := c.Logout(ctx, token); err != nil {
		return fmt.Errorf("falha no Logout: %w", err)
	}
	log.Println("... Logout OK.")
	return nil
}
