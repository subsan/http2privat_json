package connector

type JsonEntity struct {
	Method           string            `json:"method"`
	Step             int               `json:"step"`
	Params           map[string]string `json:"params"`
	Error            bool              `json:"error"`
	ErrorDescription string            `json:"errorDescription"`
}
