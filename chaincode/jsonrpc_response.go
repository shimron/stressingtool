package chaincode

type jsonrpcResponse struct {
	JSONRPC string         `json:"jsonrpc"` // 2.0
	Result  *jsonrpcResult `json:"result"`
	Error   *jsonrpcError  `json:"error"`
	ID      int            `json:"id"` // same with request
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type jsonrpcResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
