package main

import (
	"fmt"

	"github.com/shimron/stressingtool/job"
	"github.com/shimron/stressingtool/runner"
)

func main() {

	// queryRunner := runner.NewJobRunner("query_runner", 100)

	// ch := make(chan *job.Job, 100)
	// go func() {
	// 	cmd := job.ChainCodeCommand{
	// 		URL:      "http://122.224.6.212:7050/chaincode",
	// 		CCID:     "12bf1d21203f9b695e993f2ce27e65de8cee15fabf4612a50eab31985dfec66dfdb5bcfe88533f2cb9bafa6c6a8f86072bbf07dbae6aee22d220653cce20b361",
	// 		Args:     []string{"getUser", `{"userEmail":"test5@test.com"}`},
	// 		IsInvoke: false,
	// 	}
	// 	for i := 1; i <= 100000; i++ {
	// 		jb := job.NewJob(fmt.Sprintf("query_job_%d", i), cmd)
	// 		ch <- jb
	// 	}
	// 	close(ch)
	// }()
	// go queryRunner.ListenBlock("122.224.6.212:7053")
	// queryRunner.Execute(ch)

	// time.Sleep(5 * time.Second)
	// queryRunner.CollectStates()

	createUserRunner := runner.NewJobRunner("create_user_runner", 10, "127.0.0.1:7053")

	ch := make(chan *job.Job, 100)

	go func() {

		cmd := job.ChainCodeCommand{
			URL:      "http://localhost:7050/chaincode",
			CCID:     "dd49c619f32b5d8c0168f0a1c15c6bd556839c0d59ad8aba759430068071b27b0905bc6d6e284d0a9a8fde60ad418c5f169d40ae5ce70e3bd723e47b7567b7eb",
			IsInvoke: true,
		}
		for i := 1; i <= 10000; i++ {
			args := []string{"createUser", fmt.Sprintf(`{"userEmail":"test@test%d.com","userName":"test82_%d","userMobile":"test_%d","userIdentityID":"teyst2_%d","userPassword":"1232424"}`, i, i, i, i)}
			cmd.Args = args

			jb := job.NewJob(fmt.Sprintf("create_user_job_%d", i), cmd)
			ch <- jb
		}
		close(ch)
	}()

	createUserRunner.Execute(ch)
	<-createUserRunner.NoEventChan
	createUserRunner.CollectStates()
}
