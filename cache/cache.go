package cache

import "github.com/shimron/stressingtool/job"
import "sync"

type JobStatMap struct {
	JobStats map[string]*job.JobStat
	TXJobIds map[string]string //txid->jobId

	Lock sync.RWMutex
}

func NewJobStatMap() *JobStatMap {
	return &JobStatMap{
		JobStats: make(map[string]*job.JobStat),
		TXJobIds: make(map[string]string),
		Lock:     sync.RWMutex{},
	}
}

func (jsm *JobStatMap) Set(js *job.JobStat) error {
	jsm.Lock.Lock()
	defer jsm.Lock.Unlock()
	if js == nil {
		return nil
	}
	jsm.JobStats[js.JobID] = js
	if js.TXID != "" {
		jsm.TXJobIds[js.TXID] = js.JobID
	}
	return nil
}

func (jsm *JobStatMap) Get(jobId string) *job.JobStat {
	jsm.Lock.RLock()
	defer jsm.Lock.RUnlock()
	js := jsm.JobStats[jobId]
	return js
}

func (jsm *JobStatMap) GetJobStatByTXID(txid string) *job.JobStat {
	jsm.Lock.RLock()
	defer jsm.Lock.RUnlock()
	jobID, ok := jsm.TXJobIds[txid]
	if !ok || jobID == "" {
		return nil
	}
	return jsm.JobStats[jobID]
}

type TxStatMap struct {
	TxStats map[string]*job.JobStat
	Lock    sync.RWMutex
}

func NewTxStatMap() *TxStatMap {
	return &TxStatMap{
		TxStats: make(map[string]*job.JobStat),
		Lock:    sync.RWMutex{},
	}
}

func (tsm *TxStatMap) Set(tx *job.JobStat) error {
	tsm.Lock.Lock()
	defer tsm.Lock.Unlock()
	if tx == nil {
		return nil
	}
	tsm.TxStats[tx.TXID] = tx
	return nil
}

func (tsm *TxStatMap) Get(txid string) *job.JobStat {
	tsm.Lock.RLock()
	defer tsm.Lock.RUnlock()
	js := tsm.TxStats[txid]
	return js
}
