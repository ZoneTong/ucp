package multiple

// mtu=1500 whole IP layer
// -20 ip header
// -8  udp header
// = 1472
func Fragment(data []byte, mtu int) (payload [][]byte) {
	total := len(data)
	for total > mtu {
		payload = append(payload, data[:mtu])
		data = data[mtu:]
		total -= mtu
	}
	if total > 0 {
		payload = append(payload, data)
	}

	return
}
