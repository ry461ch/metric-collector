// Module for addr parsing
package netaddr

import (
	"errors"
	"strconv"
	"strings"
)

// Модель хранения адреса
type NetAddress struct {
	Host string
	Port int64
}

// Строковое представление адреса
func (a NetAddress) String() string {
	return a.Host + ":" + strconv.FormatInt(a.Port, 10)
}

// Создание объекта тпа NetAddress из json
func (a *NetAddress) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

// Создание объекта тпа NetAddress из строки
func (a *NetAddress) Set(s string) error {
	if s == "" {
		return nil
	}

	hp := strings.Split(s, ":")
	switch len(hp) {
	case 1:
		a.Host = hp[0]
	case 2:
		port, err := strconv.ParseInt(hp[1], 10, 0)
		if err != nil {
			return err
		}
		a.Host = hp[0]
		a.Port = port
	default:
		return errors.New("need address in a form host:port")
	}
	return nil
}
