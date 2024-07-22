package netaddr

import (
	"errors"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int64
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.FormatInt(a.Port, 10)
}

func (a *NetAddress) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

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
