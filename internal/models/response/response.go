package response

type ResponseEmptyObject struct{}

type ResponseErrorObject struct {
	Detail string `json:"detail,omitempty"`
}
