package main

import (
	"fmt"
	"mania/constant"
	"sync"
	"time"

	"mania/tcpx"

	"mania/apis/protocol"
	"net"
)

func main() {
	var mutex sync.Mutex
	conn, _ := net.Dial("tcp", "127.0.0.1:7170")
	//conn, _ := net.Dial("tcp", "192.9.139.115:80")
	go func() {
		for {
			buf, _ := tcpx.FirstBlockOf(conn)
			fmt.Println(string(buf))
			time.Sleep(time.Second)
			mutex.Unlock()
		}
	}()

	mutex.Lock()
	buf, e := tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: constant.LocalOnline,
		Body: protocol.ReqUserLogin{
			Credential: "",
			Duid:       "1625040488_lihenan_85",
			Package:    "",
			PID:        1,
			P:          1,
		},
	}, tcpx.JsonMarshaller{})
	if e != nil {
		panic(e)
	}

	_, _ = conn.Write(buf)

	mutex.Lock()
	buf, e = tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: constant.SnakeLadder,
		Body:      protocol.ReqSnakeLadder{Action: 1},
	}, tcpx.JsonMarshaller{})
	if e != nil {
		panic(e)
	}

	_, _ = conn.Write(buf)
	//
	//for i := 0; i < 50; i++ {
	//	mutex.Lock()
	//	buf, e = tcpx.PackWithMarshaller(tcpx.Message{
	//		MessageID: constant.SNAKE_LADDER,
	//		Body:      protocol.ReqSnakeLadder{Action: 2},
	//	}, tcpx.JsonMarshaller{})
	//	if e != nil {
	//		panic(e)
	//	}
	//
	//	conn.Write(buf)
	//}
	//
	//mutex.Lock()
	//buf, e = tcpx.PackWithMarshaller(tcpx.Message{
	//	MessageID: constant.SNAKE_LADDER,
	//	Body:      protocol.ReqSnakeLadder{Action: 3},
	//}, tcpx.JsonMarshaller{})
	//if e != nil {
	//	panic(e)
	//}
	//
	//conn.Write(buf)

	time.Sleep(200 * time.Second)
}
