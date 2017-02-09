package runner

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shimron/stressingtool/cache"
	"github.com/shimron/stressingtool/event"
	"github.com/shimron/stressingtool/job"

	pb "github.com/hyperledger/fabric/protos"
)

//JobRunner ...
type JobRunner struct {
	Name           string
	EventAddr      string
	States         *cache.JobStatMap
	TxStats        *cache.TxStatMap
	ConcurrencyNum int
	StopChan       chan struct{}
	IsStopped      bool
	StartTime      time.Time
	StopTime       time.Time
	NoEventChan    chan struct{}
	once           sync.Once
}

//NewJobRunner create a new JobRunner
func NewJobRunner(name string, concurrencyNum int, eventAddr string) *JobRunner {

	if concurrencyNum <= 0 {
		concurrencyNum = 10
	}

	return &JobRunner{
		Name:           name,
		ConcurrencyNum: concurrencyNum,
		EventAddr:      eventAddr,
		StopChan:       make(chan struct{}),
		NoEventChan:    make(chan struct{}),
		States:         cache.NewJobStatMap(),
		TxStats:        cache.NewTxStatMap(),
		once:           sync.Once{},
	}
}

//Execute execute jobs from job channel
func (jr *JobRunner) Execute(jobChan <-chan *job.Job) {

	jr.once.Do(func() {
		go jr.listenBlock(jr.EventAddr)
		time.Sleep(1 * time.Second)

		jr.StartTime = time.Now()

		ticks := make(chan struct{}, jr.ConcurrencyNum)
		for i := 0; i < jr.ConcurrencyNum; i++ {
			ticks <- struct{}{}
		}

		var wg sync.WaitGroup
	loop:
		for {
			select {
			case jb, ok := <-jobChan:
				if !ok {
					fmt.Println("chan was closed")
					break loop
				}
				wg.Add(1)
				<-ticks
				fmt.Printf("receive new job:%s\n", jb.Name)
				go func(jb *job.Job) {
					defer wg.Done()
					js := jb.Run()
					fmt.Printf("%s has done\n", jb.Name)
					err := jr.States.Set(js)
					if err != nil {
						fmt.Printf("fail to set jobstat:%v\n", err)
					}
					ticks <- struct{}{}
				}(jb)
			case <-jr.StopChan:
				fmt.Printf("stopping job runner")
				break loop
			}
			runtime.Gosched()
		}
		fmt.Println("waiting for jobs to be done...")
		wg.Wait()
		fmt.Println("all jobs were executed")
		jr.StopTime = time.Now()
	},
	)

}

func (jr *JobRunner) listenBlock(url string) {
	ec := event.NewEventClient(url)
	if ec == nil {
		fmt.Printf("fail to create new event client")
		os.Exit(-1)
	}
	var wg sync.WaitGroup
loop:
	for {
		select {
		case b := <-ec.Notify:
			wg.Add(1)
			go func(b *pb.Event_Block) {
				defer wg.Done()
				if len(b.Block.Transactions) != 0 {

					blockTimestamp := b.Block.GetNonHashData().GetLocalLedgerCommitTimestamp()
					blockTime := time.Unix(blockTimestamp.Seconds, int64(blockTimestamp.Nanos))

					for _, tx := range b.Block.Transactions {
						fmt.Printf("%s was written to ledger\n", tx.Txid)
						//	time.Sleep(2 * time.Second)
						js := jr.States.GetJobStatByTXID(tx.Txid)
						if js == nil {
							fmt.Printf("jobstat not found for %s\n", tx.Txid)
							continue
						}
						js.IsDone = true
						js.IsSuccess = true
						js.TXConfirmedTime = blockTime
						jr.TxStats.Set(js)
					}
				}
			}(b)

		case r := <-ec.Rejected:
			wg.Add(1)
			go func(r *pb.Event_Rejection) {
				defer wg.Done()
				fmt.Printf("%s was rejected\n", r.Rejection.Tx.Txid)
				//	time.Sleep(2 * time.Second)
				js := jr.States.GetJobStatByTXID(r.Rejection.Tx.Txid)
				if js == nil {
					fmt.Printf("jobstat not found for %s\n", r.Rejection.Tx.Txid)
					return
				}
				js.IsSuccess = false
				js.IsDone = true
				js.TXConfirmedTime = time.Now()
				js.ErrorMsg = r.Rejection.ErrorMsg
				jr.TxStats.Set(js)

			}(r)

		case <-time.After(20 * time.Second):
			break loop
		}
		runtime.Gosched()
	}
	wg.Wait()
	jr.NoEventChan <- struct{}{}
}

//Stop stop job runner
func (jr *JobRunner) Stop() {
	if jr.IsStopped {
		fmt.Printf("runner has been stopped")
		return
	}
	jr.IsStopped = true
	close(jr.StopChan)
}

//CollectStates caculate summary info
func (jr *JobRunner) CollectStates() {
	totalTimeCost := time.Now().Sub(jr.StartTime).Nanoseconds()
	totalSubmitTimeCost := jr.StopTime.Sub(jr.StartTime).Nanoseconds()
	jobCount := len(jr.States.JobStats)
	//save 10 failed job name ( only used to  validate  transactions were failed exactly )
	var failedJobs = make([]string, 0, 10)
	var successCount int
	var failedCount int
	var finishedCount int
	var avgExecutionCost float64
	var avgConfirmCost float64
	var minExecutionCost float64
	var minConfirmCost float64
	var maxExecutionCost float64
	var maxConfirmCost float64

	var totalExecutionCost int64
	var totalConfirmCost int64
	for _, jb := range jr.States.JobStats {
		executionCost := jb.ExecutedTime.Sub(jb.SubmitTime).Nanoseconds()
		if executionCost > 0 {
			totalExecutionCost += executionCost
			if minExecutionCost != 0 {
				minExecutionCost = math.Min(float64(minExecutionCost), float64(executionCost))
			} else {
				minExecutionCost = float64(executionCost)
			}
			if maxExecutionCost != 0 {
				maxExecutionCost = math.Max(float64(maxExecutionCost), float64(executionCost))
			} else {
				maxExecutionCost = float64(executionCost)
			}
		}

		if len(jb.TXID) == 0 {
			successCount++
			finishedCount++
			continue
		}
		//未找到对应的txid对应的job stat，认为任务失败
		txStat := jr.TxStats.Get(jb.TXID)
		if txStat == nil {
			failedCount++
			if len(failedJobs) < cap(failedJobs) {
				failedJobs = append(failedJobs, jb.Name)
			}
			continue
		}

		finishedCount++
		//收到tx的block event认定为成功，收到rejection event认定为失败
		if txStat.IsSuccess {
			successCount++
		} else {
			failedCount++
			if len(failedJobs) < cap(failedJobs) {
				failedJobs = append(failedJobs, jb.Name)
			}
			continue
		}
		//仅计算写入ledger的交易确认时间
		confirmCost := txStat.TXConfirmedTime.Sub(jb.ExecutedTime).Nanoseconds()
		if confirmCost > 0 {
			totalConfirmCost += confirmCost

			if minConfirmCost != 0 {
				minConfirmCost = math.Min(float64(minConfirmCost), float64(confirmCost))
			} else {
				minConfirmCost = float64(confirmCost)
			}

			if maxConfirmCost != 0 {
				maxConfirmCost = math.Max(float64(maxConfirmCost), float64(confirmCost))
			} else {
				maxConfirmCost = float64(confirmCost)
			}
		}
	}

	avgExecutionCost = float64(totalExecutionCost) / float64(jobCount)
	avgConfirmCost = float64(totalConfirmCost) / float64(successCount)

	fmt.Println("********Summary*******")
	fmt.Printf("total job count:%d\n", jobCount)
	fmt.Printf("total time cost:%fs\n", float64(totalTimeCost)/1000000000)
	fmt.Printf("total job execution time cost:%fs\n", float64(totalSubmitTimeCost)/1000000000)
	fmt.Printf("finished job count:%d\n", finishedCount)
	fmt.Printf("successful job count:%d\n", successCount)
	fmt.Printf("failed job count:%d\n", failedCount)
	fmt.Printf("min execution cost:%fs\n", minExecutionCost/1000000000)
	fmt.Printf("max execution cost:%fs\n", maxExecutionCost/1000000000)
	fmt.Printf("avg execution cost:%fs\n", avgExecutionCost/1000000000)
	fmt.Printf("min confirm cost:%fs\n", minConfirmCost/1000000000)
	fmt.Printf("max confirm cost:%fs\n", maxConfirmCost/1000000000)
	fmt.Printf("avg confirm cost:%fs\n", avgConfirmCost/1000000000)
	fmt.Printf("first 10 failed job names:%v\n", failedJobs)
}
