package model

type Response struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}
type FinalResponse struct {
	Status string
	Info   string
	Data   interface{}
}

func (r Response) Error() string {
	return r.Info
}
