package chaincode

type jsonrpcRequest struct {
	JSONRPC string       `json:"jsonrpc"` // 2.0
	Method  string       `json:"method"`  //query or invoke
	Params  jsonrpcParam `json:"params"`
	ID      int          `json:"id"` // 1
}

type jsonrpcParam struct {
	Type        int              `json:"type"` // 1
	ChaincodeID jsonrpcChaincode `json:"chaincodeID"`
	CtorMsg     jsonrcpCtorMsg   `json:"ctorMsg"`
}

type jsonrpcChaincode struct {
	Name string `json:"name"`
}

type jsonrcpCtorMsg struct {
	Args []string `json:"args"`
}

//newJSONRPCRequest return a new jsonrpc request
func newJSONRPCRequest(isInvoke bool, ccid string, args []string) *jsonrpcRequest {
	var method string
	if isInvoke {
		method = "invoke"
	} else {
		method = "query"
	}
	req := &jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		ID:      1,
		Params: jsonrpcParam{
			Type: 1,
			ChaincodeID: jsonrpcChaincode{
				Name: ccid,
			},
			CtorMsg: jsonrcpCtorMsg{
				Args: args,
			},
		},
	}
	return req
}
