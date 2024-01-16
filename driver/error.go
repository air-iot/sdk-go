package driver

import (
	"errors"
	"io"
	"strings"

	"github.com/air-iot/logger"
)

type ErrorType int

const (
	UNKONWN                     ErrorType = 1
	TIMEOUT                     ErrorType = 2
	CONNECTION_FAIELD           ErrorType = 3
	CONNECTION_CLOSED           ErrorType = 4
	CONNECTION_EOF              ErrorType = 5
	MODBUS_TRANSACTION          ErrorType = 6
	MODBUS_ILLEGAL_DATA_ADDRESS ErrorType = 7
)

func TcpClientErrSuggest(err error) (ErrorType, error) {
	if strings.Contains(err.Error(), "timeout") {
		return TIMEOUT, logger.NewErrorFocusNotice("检查网络是否延时;检查服务端设备资源(CPU、内存等)占用是否过高资源不够(降低采集频率)", err)
	} else if strings.Contains(err.Error(), "An established connection was aborted by the software in your host machine") ||
		strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") ||
		strings.Contains(err.Error(), "No connection could be made because the target machine actively refused it") ||
		strings.Contains(err.Error(), "connection refused") {
		return CONNECTION_FAIELD, logger.NewErrorFocusNotice("检查服务端设备是否开机,网络端口是否通,防火墙端口是否开放", err)
	} else if strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "connection reset by peer") {
		return CONNECTION_CLOSED, logger.NewErrorFocusNotice("检查服务端设备是否超过了最大连接数", err)
	} else if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return CONNECTION_EOF, logger.NewErrorFocusNotice("检查服务端设备是否关闭了连接", err)
	}
	return UNKONWN, err
}

func ModbusErrSuggest(err error) (ErrorType, error) {
	if strings.Contains(err.Error(), "modbus: response transaction id") && strings.Contains(err.Error(), "does not match request") {
		return MODBUS_TRANSACTION, logger.NewErrorFocusNotice("检查服务端设备是否有多个连接在读写引起事务不一致", err)
	} else if strings.Contains(err.Error(), "illegal data address") {
		return MODBUS_ILLEGAL_DATA_ADDRESS, logger.NewErrorFocusNotice("检查站号和数据点地址是否配置正确", err)
	} else {
		return TcpClientErrSuggest(err)
	}
}
