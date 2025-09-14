package button

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rustyeddy/otto/device"
	"github.com/rustyeddy/otto/messanger"
)

var (
	gotit [2]bool
	wg    sync.WaitGroup
)

func TestButton(t *testing.T) {
	device.Mock(true)
	done := make(chan any)

	// c := messanger.GetMockClient()
	// m := messanger.SetMQTTClient(c)
	// err := m.Connect()
	// if err != nil {
	// 	t.Error("Failed to connect to MQTT broker: ", err)
	// }
	messanger.NewMessanger("local", "button")

	b := New("button", 23)
	//b.SetTopic(messanger.GetTopics().Control("button"))
	b.Subscribe(messanger.GetTopics().Control("button"), b.MsgHandler)
	go b.EventLoop(done, b.ReadPub)

	wg.Add(2)
	b.MockHWInput(0)
	b.MockHWInput(1)

	wg.Wait()
	time.Sleep(10 * time.Millisecond)
	b.Close()
	done <- true

	if !gotit[0] || !gotit[1] {
		t.Errorf("failed to get 0 and 1 got (%t) and (%t)", gotit[0], gotit[1])
	}

}

func (b *Button) MsgHandler(msg *messanger.Msg) {
	i, err := strconv.Atoi(msg.String())
	if err != nil {
		println("button return error")
		return
	}
	gotit[i] = true
	wg.Done()
	return
}
