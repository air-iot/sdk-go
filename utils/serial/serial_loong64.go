//go:build loong64

package serial

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func GetSerialPorts() ([]string, error) {
	// Linux系统串口设备目录
	const devDir = "/dev"
	files, err := ioutil.ReadDir(devDir)
	if err != nil {
		return nil, err
	}

	var ports []string
	for _, file := range files {
		name := file.Name()
		// 识别常见的串口设备前缀
		if strings.HasPrefix(name, "ttyS") || // 标准串口
			strings.HasPrefix(name, "ttyUSB") || // USB转串口
			strings.HasPrefix(name, "ttyAMA") || // ARM串口(龙芯兼容)
			strings.HasPrefix(name, "ttyACM") { // CDC ACM设备
			ports = append(ports, filepath.Join(devDir, name))
		}
	}
	return ports, nil
}
