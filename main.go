package main

/*
#include <string.h>
#include <stdlib.h>
*/
import "C"
import (
	"crypto/sha256"
	"fmt"
	"time"
	"unsafe"

	kcp "github.com/xtaci/kcp-go/v5"
)

var connections = map[int]*kcp.UDPSession{}
var connID = 0

func dialKCP(addr, key string, dataShards, parityShards int) int {
	h := sha256.Sum256([]byte(key))
	block, err := kcp.NewAESBlockCrypt(h[:])
	if err != nil {
		return -1
	}
	conn, err := kcp.DialWithOptions(addr, block, dataShards, parityShards)
	if err != nil {
		return -1
	}
	connID++
	connections[connID] = conn
	return connID
}

//export KCPDial
func KCPDial(addr *C.char, ckey *C.char, dataShards C.int, parityShards C.int) C.int {
	return C.int(dialKCP(C.GoString(addr), C.GoString(ckey), int(dataShards), int(parityShards)))
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

func main() {
	secretKey := "meow"

	fmt.Println("Server starting...")
	key := sha256.Sum256([]byte(secretKey))
	encryption, err := kcp.NewAESBlockCrypt(key[:])
	if err != nil {
		panic(err)
	}

	listener, err := kcp.ListenWithOptions(":6969", encryption, 10, 3)
	if err != nil {
		panic(err)
	}

	go testLib()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		go func(c *kcp.UDPSession) {
			defer c.Close()
			buf := make([]byte, 1024)

			// 1. Read from client
			n, err := c.Read(buf)
			if err != nil {
				return
			}
			fmt.Printf("[Server] Received: %s\n", string(buf[:n]))

			// 2. Echo back / Respond
			c.Write([]byte("Hello from Server!"))
		}(conn.(*kcp.UDPSession))
	}
}

func testLib() {
	// Wait for server to be ready
	time.Sleep(time.Second)

	// --- 1. Dial ---
	cAddr := C.CString("127.0.0.1:6969")
	cKey := C.CString("meow")
	defer C.free(unsafe.Pointer(cAddr))
	defer C.free(unsafe.Pointer(cKey))

	id := KCPDial(cAddr, cKey, 10, 3)
	if id < 0 {
		fmt.Println("[Client] Dial failed")
		return
	}
	fmt.Printf("[Client] Connected with ID: %d\n", int(id))

	// --- 2. Send ---
	payload := C.CString("Hello from C-Land!")
	defer C.free(unsafe.Pointer(payload))

	sent := KCPSend(id, payload, C.int(len("Hello from C-Land!")))
	fmt.Printf("[Client] Sent %d bytes\n", int(sent))

	// --- 3. Receive ---
	// Pre-allocate a buffer in C-style (simulating how a C caller would do it)
	cBuf := (*C.char)(C.malloc(1024))
	defer C.free(unsafe.Pointer(cBuf))

	fmt.Println("[Client] Waiting for response...")
	recvd := KCPRecv(id, cBuf, 1024)

	if recvd > 0 {
		// Convert C buffer back to Go string for printing
		response := C.GoStringN(cBuf, recvd)
		fmt.Printf("[Client] Received: %s\n", response)
	} else {
		fmt.Println("[Client] Receive failed or connection closed")
	}

	// --- 4. Close ---
	KCPClose(id)
	fmt.Println("[Client] Connection closed")
}
