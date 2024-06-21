package response

type EmptyObject struct {}

type ErrorObject struct {
	Detail string `json:"detail,omitempty"`
}
