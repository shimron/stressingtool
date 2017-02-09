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

	jr.once.Do(
		func() {
			go jr.listenBlock(jr.EventAddr)
			time.Sleep(2 * time.Second)

			jr.StartTime = time.Now()

			if jr.ConcurrencyNum > 0 {
				ticks := make(chan struct{}, jr.ConcurrencyNum)
				for i := 0; i < jr.ConcurrencyNum; i++ {
					ticks <- struct{}{}
				}
			loop1:
				for {
					select {
					case jb, ok := <-jobChan:
						if !ok {
							fmt.Println("chan was closed")
							break loop1
						}
						<-ticks
						fmt.Printf("receive new job:%s\n", jb.Name)
						go func(jb *job.Job) {
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
						break loop1
					}
					runtime.Gosched()
				}

			} else {

			loop2:
				for {
					select {
					case jb, ok := <-jobChan:
						if !ok {
							fmt.Println("chan was closed")
							break loop2
						}
						fmt.Printf("receive new job:%s\n", jb.Name)
						go func(jb *job.Job) {
							js := jb.Run()
							jr.States.Set(js)
						}(jb)
					case <-jr.StopChan:
						fmt.Printf("stopping job runner...")
						break loop2
					}
					runtime.Gosched()
				}
			}
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
	for {
		select {
		case b := <-ec.Notify:

			go func(b *pb.Event_Block) {
				if len(b.Block.Transactions) != 0 {

					blockTime := time.Unix(b.Block.GetNonHashData().GetLocalLedgerCommitTimestamp().Seconds, int64(b.Block.GetNonHashData().GetLocalLedgerCommitTimestamp().Nanos))

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
			go func(r *pb.Event_Rejection) {
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
			jr.NoEventChan <- struct{}{}

		}
		runtime.Gosched()
	}
}

//Stop stop job runner
func (jr *JobRunner) Stop() {
	if jr.IsStopped {
		fmt.Printf("runner has been stopped")
		return
	}
	close(jr.StopChan)
}

//CollectStates caculate summary info
func (jr *JobRunner) CollectStates() {

	totalCost := jr.StopTime.Sub(jr.StartTime).Nanoseconds()
	jobCount := len(jr.States.JobStats)
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

		txStat := jr.TxStats.Get(jb.TXID)
		if txStat == nil {
			failedCount++
			continue
		}

		finishedCount++

		if txStat.IsSuccess {
			successCount++
		} else {
			failedCount++
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
	fmt.Printf("total time cost:%ds\n", totalCost*1.0/1000000000)
	fmt.Printf("finished job count:%d\n", finishedCount)
	fmt.Printf("successful job count:%d\n", successCount)
	fmt.Printf("failed job count:%d\n", failedCount)
	fmt.Printf("min execution cost:%fs\n", minExecutionCost/1000000000)
	fmt.Printf("max execution cost:%fs\n", maxExecutionCost/1000000000)
	fmt.Printf("avg execution cost:%fs\n", avgExecutionCost/1000000000)
	fmt.Printf("min confirm cost:%fs\n", minConfirmCost/1000000000)
	fmt.Printf("max confirm cost:%fs\n", maxConfirmCost/1000000000)
	fmt.Printf("avg confirm cost:%fs\n", avgConfirmCost/1000000000)
}
