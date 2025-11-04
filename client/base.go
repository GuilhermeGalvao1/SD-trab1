package client

import (
	"context"
	"fmt"
	"net"
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
