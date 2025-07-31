package decrypt

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/air-iot/sdk-go/v4/utils/cipherx"
	"github.com/duke-git/lancet/v2/cryptor"
)

const (
	ENV_KEY = "AIRIOT_CIPHER_KEY" // 配置文件加密密钥
)

func Decode() {

	key := []byte(os.Getenv(ENV_KEY))
	keys := viper.AllKeys()
	for _, k := range keys {
		v := viper.Get(k)
		vStr, ok := v.(string)
		if ok {
			viper.Set(k, decode(key, vStr))
		}
	}
}

func decode(key []byte, val string) string {
	// 处理 AES_模式(...) 格式
	if strings.HasPrefix(val, "AES_") && strings.Contains(val, "(") && strings.HasSuffix(val, ")") {
		// 提取模式部分
		modeEnd := strings.Index(val, "(")
		if modeEnd > 5 { // AES_ 是4个字符，至少有一个模式字符
			mode := val[4:modeEnd]
			encryptedStr := val[modeEnd+1 : len(val)-1]

			decrypted, err := cipherx.Decrypt(encryptedStr, key, cipherx.AES, cipherx.Mode(mode))
			if err != nil {
				panic(fmt.Sprintf("解密错误: %v (输入: %s)", err, encryptedStr))
			}
			return string(decrypted)
		}
	}

	// 处理 SM4_模式(...) 格式
	if strings.HasPrefix(val, "SM4_") && strings.Contains(val, "(") && strings.HasSuffix(val, ")") {
		// 提取模式部分
		modeEnd := strings.Index(val, "(")
		if modeEnd > 5 { // SM4_ 是4个字符，至少有一个模式字符
			mode := val[4:modeEnd]
			encryptedStr := val[modeEnd+1 : len(val)-1]

			decrypted, err := cipherx.Decrypt(encryptedStr, key, cipherx.SM4, cipherx.Mode(mode))
			if err != nil {
				panic(fmt.Sprintf("解密错误: %v (输入: %s)", err, encryptedStr))
			}
			return string(decrypted)
		}
	}

	// 兼容旧格式
	if strings.HasPrefix(val, "ENC(") && strings.HasSuffix(val, ")") {
		vStr := strings.TrimSuffix(strings.TrimPrefix(val, "ENC("), ")")
		return cryptor.Base64StdDecode(vStr)
	} else if strings.HasPrefix(val, "AES(") && strings.HasSuffix(val, ")") {
		vStr := strings.TrimSuffix(strings.TrimPrefix(val, "AES("), ")")
		desVal, err := cipherx.Decrypt(vStr, key, cipherx.AES, cipherx.GCM)
		if err != nil {
			panic(fmt.Sprintf("解析密码错误: %v", err))
		}
		return string(desVal)
	}
	return val
}
