package event

import (
	"fmt"

	"github.com/hyperledger/fabric/events/consumer"

	"time"

	"os"

	pb "github.com/hyperledger/fabric/protos"
)

//EventConsumer ...
type EventConsumer struct {
	Notify   chan *pb.Event_Block
	Rejected chan *pb.Event_Rejection
}

func (ec *EventConsumer) GetInterestedEvents() ([]*pb.Interest, error) {
	return []*pb.Interest{{EventType: pb.EventType_BLOCK}, {EventType: pb.EventType_REJECTION}}, nil
}

func (ec *EventConsumer) Recv(msg *pb.Event) (bool, error) {
	if e, ok := msg.Event.(*pb.Event_Block); ok {
		ec.Notify <- e
		return true, nil
	}
	if e, ok := msg.Event.(*pb.Event_Rejection); ok {
		ec.Rejected <- e
		return true, nil
	}
	return false, fmt.Errorf("receive unknown event type:%v", msg)
}

func (ec *EventConsumer) Disconnected(err error) {
	fmt.Printf("receive error:%v\n", err)
	os.Exit(-1)
}

func NewEventClient(addr string) *EventConsumer {
	done := make(chan *pb.Event_Block, 10000)
	reject := make(chan *pb.Event_Rejection, 10000)
	adapter := &EventConsumer{Notify: done, Rejected: reject}
	obcEHClient, _ := consumer.NewEventsClient(addr, 30*time.Second, adapter)
	if err := obcEHClient.Start(); err != nil {
		fmt.Printf("could not start chat:%v\n", err)
		obcEHClient.Stop()
		return nil
	}
	fmt.Println("block listener is now serving...")
	return adapter
}
