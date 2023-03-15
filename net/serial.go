package net

import (
	"didaGatewayCenter/domain"
	"github.com/goburrow/serial"
	"time"
)

type serial1 struct {
	fd serial.Port
}

func (s *serial1) ReadTimeout(t time.Duration) ([]byte, error) {
	var buffer []byte
	nowTime := time.Now()
	if t <= 0 {
		t = time.Duration(3) * time.Second
	}
	for {
		temp := make([]byte, 1)
		n, err := s.fd.Read(temp)
		if err != nil {
			if len(buffer) == 0 {
				if time.Now().After(nowTime.Add(t)) {
					return nil, err
				}
			}
			return buffer, nil
		} else {
			buffer = append(buffer, temp[:n]...)
		}
	}
}

func (s *serial1) WriteTimeout(writeData []byte, t time.Duration) error {
	nowTime := time.Now()
	if t <= 0 {
		t = time.Duration(1) * time.Second
	}
	length := len(writeData)
	for {
		if n, err := s.fd.Write(writeData); err != nil {
			if len(writeData) == length {
				if time.Now().After(nowTime.Add(t)) {
					return err
				}
			}
		} else {
			if n != len(writeData) {
				writeData = writeData[n:]
			} else {
				return nil
			}
		}
	}
}

func (s *serial1) WriteReadTimeout(writeData []byte, t time.Duration) ([]byte, error) {
	_, _ = s.ReadTimeout(10)
	if err := s.WriteTimeout(writeData, t); err != nil {
		return nil, err
	}
	return s.ReadTimeout(t)
}

func New(fd serial.Port) domain.Software {
	return &serial1{
		fd: fd,
	}
}
