package messanger

// type tclient struct {
// 	gotit bool
// 	topic string
// 	msg   string
// }

// func (t *tclient) Callback(msg *Msg) {
// 	if msg.Path[0] != "t" || msg.Path[1] != "test" {
// 		return
// 	}

// 	if msg.String() != "message" {
// 		return
// 	}
// 	t.gotit = true
// 	return
// }

// func TestSubscribe(t *testing.T) {
// 	c := GetMockClient()
// 	m := SetMQTTClient(c)
// 	err := m.Connect()
// 	if err != nil {
// 		t.Error("Failed to connect to MQTT broker: ", err)
// 	}

// 	tc := &tclient{
// 		gotit: false,
// 		topic: "t/test",
// 		msg:   "message",
// 	}

// 	m.Subscribe(tc.topic, tc.Callback)
// 	m.Publish(tc.topic, tc.msg)

// 	if tc.gotit == false {
// 		t.Error("Expected to recv message but did not")
// 	}
// }
