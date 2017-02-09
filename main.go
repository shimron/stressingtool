package main

import (
	"fmt"

	"github.com/shimron/stressingtool/job"
	"github.com/shimron/stressingtool/runner"
)

func main() {

	// queryRunner := runner.NewJobRunner("query_runner", 10, "127.0.0.1:7053")

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

	// queryRunner.Execute(ch)
	// <-queryRunner.NoEventChan
	// queryRunner.CollectStates()

	createUserRunner := runner.NewJobRunner("create_user_runner", 10, "127.0.0.1:7053")

	ch := make(chan *job.Job, 100)

	go func() {

		cmd := job.ChainCodeCommand{
			URL:      "http://localhost:7050/chaincode",
			CCID:     "0228fd806eb1fd984eea7208db229033be9a0e64fe4d9de7048bf8cf9958053cad5b457a85f5be6328780a9af16ac1129b53026cf83c786fc498f4c7872c7166",
			IsInvoke: true,
		}
		offset := 100
		for i := 1 + offset; i <= 10000+offset; i++ {
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
