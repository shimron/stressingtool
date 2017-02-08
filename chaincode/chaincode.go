package chaincode

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

//QueryOrInvoke ...
func QueryOrInvoke(url string, ccid string, args []string, isInvoke bool) (string, error) {
	if isInvoke {
		return Invoke(url, ccid, args)
	}

	err := Query(url, ccid, args)
	return "", err
}

//Query ...
func Query(url string, ccid string, args []string) error {

	req := newJSONRPCRequest(false, ccid, args)
	resp, err := post(url, req)
	if err != nil {
		return nil
	}
	if resp.Error != nil {
		return errors.New(resp.Error.Message)
	}
	return nil
}

//Invoke ...
func Invoke(url string, ccid string, args []string) (string, error) {
	req := newJSONRPCRequest(true, ccid, args)
	resp, err := post(url, req)
	if err != nil {
		return "", nil
	}
	if resp.Error != nil {
		return "", errors.New(resp.Error.Message)
	}
	return resp.Result.Message, nil
}

func post(url string, req *jsonrpcRequest) (*jsonrpcResponse, error) {
	msg, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	body := strings.NewReader(string(msg))
	resp, err := http.DefaultClient.Post(url, "application/json", body)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	var res jsonrpcResponse
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
