package main

/*
#include <string.h>
*/
import "C"
import (
	"unsafe"

	kcp "github.com/xtaci/kcp-go/v5"
)

var connections = map[int]*kcp.UDPSession{}
var connID = 0

//export KCPDial
func KCPDial(addr *C.char, key *C.char, dataShards C.int, parityShards C.int) C.int {
	block, err := kcp.NewAESBlockCrypt([]byte(C.GoString(key)))
	conn, err := kcp.DialWithOptions(C.GoString(addr), block, int(dataShards), int(parityShards))
	if err != nil {
		return -1
	}
	connID++
	connections[connID] = conn
	return C.int(connID)
}

//export KCPSend
func KCPSend(id C.int, data *C.char, length C.int) C.int {
	conn, ok := connections[int(id)]
	if !ok {
		return -1
	}
	buf := C.GoBytes(unsafe.Pointer(data), length)
	n, err := conn.Write(buf)
	if err != nil {
		return -1
	}
	return C.int(n)
}

//export KCPRecv
func KCPRecv(id C.int, buf *C.char, maxLen C.int) C.int {
	conn, ok := connections[int(id)]
	if !ok {
		return -1
	}
	b := make([]byte, int(maxLen))
	n, err := conn.Read(b)
	if err != nil {
		return -1
	}
	C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))
	return C.int(n)
}

//export KCPClose
func KCPClose(id C.int) {
	if conn, ok := connections[int(id)]; ok {
		conn.Close()
		delete(connections, int(id))
	}
}

func main() {} // required but unused
