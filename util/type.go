package util

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type TxResult struct {
	TxHash string `json:"txhash"`
	Logs   []struct {
		Events []Event `json:"events"`
	} `json:"logs"`
}
