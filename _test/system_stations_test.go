package system

import (
	"fmt"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/rustyeddy/otto/client"
	"github.com/rustyeddy/otto/station"
	"github.com/stretchr/testify/assert"
)

func (ts *systemTest) mockStations() {
	ts.StationManager = station.GetStationManager()
	for i := 1; i < config.StationCount; i++ {
		stname := fmt.Sprintf("station-%03d", i)
		st, err := ts.StationManager.Add(stname)
		if err != nil {
			panic(err)
		}
		st.Hostname = stname
		st.Local = false
		iface := &station.Iface{
			Name:    "eth0",
			MACAddr: fmt.Sprintf("22:33:44:55:66:%02x", i),
		}
		ipstr := fmt.Sprintf("10.77.1.%d", i)
		ipaddr, err := netip.ParseAddr(ipstr)
		if err != nil {
			panic(err)
		}

		iface.IPAddrs = append(iface.IPAddrs, ipaddr)
		st.Ifaces = append(st.Ifaces, iface)
		st.StartTicker(1 * time.Minute)
		st.SayHello()
	}
}

func (ts *systemTest) testStations(t *testing.T) {
	ts.mockStations()

	cli := client.NewClient(config.URL)
	assert.NotNil(t, cli, "expected a client got nil")

	stations, err := cli.GetStations()
	assert.NoError(t, err)
	assert.True(t, len(stations) == 11 || len(stations) == 10)

	st := ts.StationManager.Get("station-009")
	assert.NotNil(t, st)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()

		<-time.After(10 * time.Second)
		t.Logf("Stopping station-09")
		st.Stop()
		<-time.After(5 * time.Minute)
	}()
	wg.Wait()

	t.Logf("Checking to insure station-009 has been expired")
	assert.Equal(t, 10, len(stations))

	// reststations, err := cli.GetStations()
	// assert.NoError(t, err)
	// assert.Equal(t, 10, len(reststations))

	st = ts.StationManager.Get("station-009")
	assert.Nil(t, st)
}
