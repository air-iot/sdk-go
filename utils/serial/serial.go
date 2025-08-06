//go:build !loong64

package serial

import (
	"fmt"
	"go.bug.st/serial"
	"runtime"
	"strings"
)

func GetSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	// 只在 Windows 系统上处理 COM10+ 端口
	if runtime.GOOS == "windows" {
		var formattedPorts []string
		for _, port := range ports {
			// 检查是否是 COM10 或更高
			if strings.HasPrefix(strings.ToUpper(port), "COM") {
				// 提取 COM 后面的数字
				comNumber := port[3:]
				// 如果是 COM10 或更高，添加 \\.\ 前缀
				if len(comNumber) >= 2 { // COM10 及以上至少有 2 位数字
					formattedPorts = append(formattedPorts, fmt.Sprintf("\\\\.\\%s", port))
					continue
				}
			}
			// 其他情况保持不变
			formattedPorts = append(formattedPorts, port)
		}
		return formattedPorts, nil
	}

	// 非 Windows 系统直接返回
	return ports, nil
}
