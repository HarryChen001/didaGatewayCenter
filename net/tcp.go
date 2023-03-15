package net

import (
	"didaGatewayCenter/domain"
	"net"
	"time"
)

type Tcp struct {
	Conn net.Conn
}

func (netTcp *Tcp) ReadTimeout(t time.Duration) ([]byte, error) {

	if t <= 0 {
		t = time.Duration(1) * time.Second
	}
	var buffer []byte
reRead:
	_ = netTcp.Conn.SetReadDeadline(time.Now().Add(t))
	temp := make([]byte, 100)
	if n, err := netTcp.Conn.Read(temp); err != nil {
		return buffer, err
	} else {
		buffer = append(buffer, temp[:n]...)
		if n == 100 {
			goto reRead
		}
		return buffer, err
	}
}
func (netTcp *Tcp) WriteTimeout(writeData []byte, t time.Duration) error {

	if t <= 0 {
		t = time.Duration(1) * time.Second
	}
	_ = netTcp.Conn.SetWriteDeadline(time.Now().Add(t))
reWrite:
	if n, err := netTcp.Conn.Write(writeData); err != nil {
		return err
	} else {
		if n < len(writeData) {
			writeData = writeData[n:]
			goto reWrite
		}
		return nil
	}
}
func (netTcp *Tcp) WriteReadTimeout(writeData []byte, t time.Duration) ([]byte, error) {
	_, _ = netTcp.ReadTimeout(time.Millisecond * 10)
	if err := netTcp.WriteTimeout(writeData, t); err != nil {
		return nil, err
	} else {
		return netTcp.ReadTimeout(t)
	}
}

func Dial(network string, address string) (domain.Software, error) {
	conn, err := net.Dial(network, address)
	return &Tcp{
		Conn: conn,
	}, err
}
