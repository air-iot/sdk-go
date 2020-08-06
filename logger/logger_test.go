/**
 * @Author: ZhangQiang
 * @Description:
 * @File:  logger_test
 * @Version: 1.0.0
 * @Date: 2020/8/6 15:46
 */
package logger

// 导入日志包
import "testing"

func TestNewLogger(t *testing.T) {
	// 创建log
	l := NewLogger("debug")
	l.Debugln("debug")
	l.Infoln("info")
	l.Warnln("warn")
	l.Errorln("error")
}
