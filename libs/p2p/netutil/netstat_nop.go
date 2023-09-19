// +build !linux,!windows

package netutil

func UdpPortListeners(port int) (processes map[int]string, err error) {
	processes = make(map[int]string)

	return
}
