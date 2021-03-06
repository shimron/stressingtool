package job

import (
	"time"

	"github.com/satori/go.uuid"
	"github.com/shimron/stressingtool/chaincode"
)

type Job struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	SubmitTime time.Time        `json:"submit_time"`
	Command    ChainCodeCommand `json:"command"`
}

type ChainCodeCommand struct {
	URL      string   `json:"rest_url"`
	CCID     string   `json:"chaincode_id"`
	Args     []string `json:"args"`
	IsInvoke bool     `json:"is_invoke"`
}

func NewJob(name string, cmd ChainCodeCommand) *Job {
	return &Job{
		ID:      uuid.NewV1().String(),
		Name:    name,
		Command: cmd,
	}
}

func (j *Job) Run() *JobStat {
	j.SubmitTime = time.Now()
	txid, err := chaincode.QueryOrInvoke(j.Command.URL, j.Command.CCID, j.Command.Args, j.Command.IsInvoke)

	var isSuccess = false
	if err == nil && !j.Command.IsInvoke {
		isSuccess = true
	}
	var isDone = true
	if j.Command.IsInvoke && err == nil {
		isDone = false
	}

	var msg string
	if err != nil {
		msg = err.Error()
	}

	return &JobStat{
		JobID:        j.ID,
		Name:         j.Name,
		TXID:         txid,
		SubmitTime:   j.SubmitTime,
		ExecutedTime: time.Now(),
		IsDone:       isDone,
		IsSuccess:    isSuccess,
		ErrorMsg:     msg,
	}
}
