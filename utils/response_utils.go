package utils

type Response struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}
type finalResponse struct {
	status string
	info   string
	data   interface{}
}

func (r Response) Error() string {
	return r.Info
}
