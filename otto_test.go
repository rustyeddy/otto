package otto

// import (
// 	"testing"
// 	"time"

// 	"github.com/rustyeddy/otto/messanger"
// 	"github.com/rustyeddy/otto/server"
// 	"github.com/rustyeddy/otto/station"
// )

// // Mock implementations for external dependencies

// type mockStation struct {
// 	initialized bool
// }

// func (m *mockStation) Init()                       { m.initialized = true }
// func (m *mockStation) Start() error                { return nil }
// func (m *mockStation) Stop()                       {}
// func (m *mockStation) MsgHandler(_ *messanger.Msg) {}

// type mockStationManager struct {
// 	addedName string
// }

// func (m *mockStationManager) Add(name string) (*station.Station, error) {
// 	m.addedName = name
// 	return &station.Station{}, nil
// }
// func (m *mockStationManager) Start() {}
// func (m *mockStationManager) Stop()  {}

// type mockServer struct {
// 	started bool
// 	closed  bool
// }

// func (m *mockServer) Start(done chan any) {
// 	m.started = true
// 	go func() {
// 		time.Sleep(10 * time.Millisecond)
// 		done <- struct{}{}
// 	}()
// }
// func (m *mockServer) Close() error {
// 	m.closed = true
// 	return nil
// }

// func TestOttO_Init(t *testing.T) {
// 	o := &OttO{Name: "test-otto"}

// 	// Test when done is already initialized
// 	o.done = make(chan any)
// 	o.Init() // Should print "Server has already been started"
// 	if o.done == nil {
// 		t.Errorf("done should not be nil after Init called with done set")
// 	}
// 	// Reset done for further tests
// 	o.done = nil

// 	// Test Mock mode
// 	o.Mock = true
// 	o.Init()
// 	if o.done == nil {
// 		t.Errorf("Init should initialize the done channel")
// 	}
// }

// func TestOttO_Done(t *testing.T) {
// 	o := &OttO{}
// 	o.done = make(chan any)
// 	ch := o.Done()
// 	if ch != o.done {
// 		t.Errorf("Done() should return the done channel")
// 	}
// }

// func TestOttO_StartAndStop(t *testing.T) {
// 	o := &OttO{Name: "test-otto"}
// 	// Use mock server and station manager
// 	mockSrv := &mockServer{}
// 	o.Server = &server.Server{}
// 	// We'll override the Start behavior below

// 	// Patch server.GetServer to return our mock
// 	origGetServer := server.GetServer
// 	server.GetServer = func() *server.Server {
// 		return o.Server
// 	}
// 	defer func() { server.GetServer = origGetServer }()

// 	// Patch Close to our mock
// 	origClose := o.Server.Close
// 	o.Server.Close = func() error {
// 		mockSrv.closed = true
// 		return nil
// 	}
// 	// Patch Start to just close the channel after a short delay
// 	origStart := o.Server.Start
// 	o.Server.Start = func(done chan any) {
// 		go func() {
// 			time.Sleep(10 * time.Millisecond)
// 			done <- struct{}{}
// 		}()
// 	}

// 	o.done = make(chan any)
// 	o.StationManager = &station.StationManager{}
// 	// Patch StationManager Start to do nothing
// 	origSMStart := o.StationManager.Start
// 	o.StationManager.Start = func() {}

// 	go func() {
// 		time.Sleep(20 * time.Millisecond)
// 		o.done <- struct{}{}
// 	}()
// 	if err := o.Start(); err != nil {
// 		t.Errorf("Start() should not return error, got: %v", err)
// 	}
// 	// Restore original functions
// 	o.Server.Close = origClose
// 	o.Server.Start = origStart
// 	o.StationManager.Start = origSMStart
// }

// func TestOttO_Stop(t *testing.T) {
// 	o := &OttO{Name: "test-otto"}
// 	o.done = make(chan any, 1)
// 	o.done <- struct{}{}

// 	mockSrv := &mockServer{}
// 	o.Server = &server.Server{}
// 	origGetServer := server.GetServer
// 	server.GetServer = func() *server.Server {
// 		return o.Server
// 	}
// 	defer func() { server.GetServer = origGetServer }()

// 	origClose := o.Server.Close
// 	o.Server.Close = func() error {
// 		mockSrv.closed = true
// 		return nil
// 	}

// 	// Messanger Close
// 	var closed bool
// 	o.Messanger = &messanger.Messanger{}
// 	origMsgClose := o.Messanger.Close
// 	o.Messanger.Close = func() { closed = true }

// 	o.Stop()
// 	if !mockSrv.closed {
// 		t.Errorf("Server should be closed on Stop()")
// 	}
// 	if !closed {
// 		t.Errorf("Messanger should be closed on Stop()")
// 	}

// 	// Restore
// 	o.Server.Close = origClose
// 	o.Messanger.Close = origMsgClose
// }

// func TestVersionInit(t *testing.T) {
// 	if Version != "0.0.9" {
// 		t.Errorf("Expected Version to be '0.0.9', got %q", Version)
// 	}
// }

// func TestOttO_Init(t *testing.T) {
// 	o := &OttO{Name: "test"}

// 	o.Init()

// 	assert.NotNil(t, o.done, "done channel should be initialized")
// 	assert.NotNil(t, o.Messanger, "Messanger should be initialized")
// 	assert.NotNil(t, o.StationManager, "StationManager should be initialized")
// 	assert.NotNil(t, o.Station, "Station should be initialized")
// 	assert.NotNil(t, o.DataManager, "DataManager should be initialized")
// 	assert.NotNil(t, o.Server, "Server should be initialized")
// }

// func TestOttO_Start(t *testing.T) {
// 	o := &OttO{
// 		Name:           "test",
// 		done:           make(chan any),
// 		StationManager: station.GetStationManager(),
// 		Server:         server.GetServer(),
// 	}

// 	go func() {
// 		time.Sleep(1 * time.Second)
// 		close(o.done)
// 	}()

// 	err := o.Start()
// 	assert.NoError(t, err, "Start should not return an error")
// }

// func TestOttO_Stop(t *testing.T) {
// 	o := &OttO{
// 		Name:      "test",
// 		done:      make(chan any),
// 		Messanger: messanger.NewMessanger("test", ""),
// 		Server:    server.GetServer(),
// 	}

// 	go func() {
// 		time.Sleep(1 * time.Second)
// 		close(o.done)
// 	}()

// 	o.Stop()
// }

// func TestOttO_Done(t *testing.T) {
// 	o := &OttO{}
// 	o.Init()
// 	done := o.Done()
// 	assert.NotNil(t, done, "Done channel should not be nil")
// }
