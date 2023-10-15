// +build linux windows

package netutil

import (
	"github.com/cakturk/go-netstat/netstat"
)

func UdpPortListeners(port int) (processes map[int]string, err error) {
	var binds []netstat.SockTabEntry
	processes = make(map[int]string)

	for _, socks := range []func(netstat.AcceptFn) ([]netstat.SockTabEntry, error){
		netstat.UDPSocks,
		netstat.UDP6Socks,
	} {
		binds, err = socks(func(se *netstat.SockTabEntry) bool {
			return int(se.LocalAddr.Port) == port
		})
		if err != nil {
			continue
		}

		for _, listener := range binds {
			processes[listener.Process.Pid] = listener.Process.Name
		}
	}
	return
}
