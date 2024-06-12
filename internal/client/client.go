package client

type ServerClient interface {
	Post(string) (int64, error)
}
