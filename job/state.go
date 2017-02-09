package job

import (
	"time"
)

//JobStat ...
type JobStat struct {
	JobID           string    `json:"job_id"`
	Name            string    `json:"name"`
	TXID            string    `json:"txid"`
	SubmitTime      time.Time `json:"submit_time"`
	ExecutedTime    time.Time `json:"executed_time"`
	TXConfirmedTime time.Time `json:"tx_confirmed_time"`
	IsSuccess       bool      `json:"is_success"`
	IsDone          bool      `json:"is_done"`
	ErrorMsg        string    `json:"error_msg"`
}
