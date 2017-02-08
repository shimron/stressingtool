package job

import (
	"time"
)

type JobStat struct {
	JobID           string    `json:"job_id"`
	TXID            string    `json:"txid"`
	CreateTime      time.Time `json:"create_time"`
	ExecutedTime    time.Time `json:"executed_time"`
	TXConfirmedTime time.Time `json:"tx_confirmed_time"`
	IsSuccess       bool      `json:"is_success"`
	IsDone          bool      `json:"is_done"`
	ErrorMsg        string    `json:"error_msg"`
}
