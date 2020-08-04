package tcp

import (
	"fmt"
	"testing"
	"time"
)

func Test_Read(t *testing.T) {
	conn, err := NewConn("tcp", "localhost", 8001)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		for i := 0; i < 100; i++ {
			str := []byte(fmt.Sprintf("%d", i))
			//写给服务器
			_, err := conn.Write(str)
			if err != nil {
				t.Log("客户端发送数据失败：", err.Error())
			} else {
				t.Log("客户端发出的数据：", i)
			}
			time.Sleep(time.Second * 4)
		}
	}()

	// 回显服务器回发的数据
	buf := make([]byte, 4096)
	j := 0
	for {
		j++
		n, err := conn.Read(buf)
		if n == 0 {
			t.Log("检查到服务器关闭!")
			time.Sleep(time.Second * 2)
			continue
			//return
		}
		if err != nil {
			t.Log("conn.Read err:", err)
			time.Sleep(time.Second * 2)
			//return
		} else {
			t.Log("客户端读到服务器回发：", string(buf[:n]))
		}
		// 测试 Close()
		//if j == 5 {
		//	conn.Close()
		//}

	}
}
