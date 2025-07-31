package cipherx

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/tjfoc/gmsm/sm4"
	"io"
	"testing"
)

func TestCBC(t *testing.T) {
	key := []byte("airiot1234567890")
	t.Log("key", string(key))
	data := []byte("dell123")

	// 加密示例
	encrypted, err := Encrypt(data, key, AES, CBC)
	if err != nil {
		panic(err)
	}
	t.Log("AES cbc", encrypted)
	// 解密示例
	decrypted, err := Decrypt(encrypted, key, AES, CBC)
	if err != nil {
		panic(err)
	}

	t.Log("AES cbc", string(decrypted))
	// 加密示例
	encrypted, err = Encrypt(data, key, SM4, CBC)
	if err != nil {
		panic(err)
	}
	t.Log("SM4 cbc", encrypted)
	// 解密示例
	decrypted, err = Decrypt(encrypted, key, SM4, CBC)
	if err != nil {
		panic(err)
	}

	t.Log("SM4 cbc", string(decrypted))
}

func TestSM4GCM(t *testing.T) {
	key := []byte("1234567890abcdef")
	data := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10}
	fmt.Printf("data = %x\n", data)
	IV := make([]byte, sm4.BlockSize)
	testA := [][]byte{ // the length of the A can be random
		[]byte{},
		[]byte{0x01, 0x23, 0x45, 0x67, 0x89},
		[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10},
	}
	for _, A := range testA {
		gcmMsg, T, err := sm4.Sm4GCM(key, IV, data, A, true)
		if err != nil {
			t.Errorf("sm4 enc error:%s", err)
		}
		fmt.Printf("gcmMsg = %x\n", gcmMsg)
		gcmDec, T_, err := sm4.Sm4GCM(key, IV, gcmMsg, A, false)
		if err != nil {
			t.Errorf("sm4 dec error:%s", err)
		}
		fmt.Printf("gcmDec = %x\n", gcmDec)
		if bytes.Compare(T, T_) == 0 {
			fmt.Println("authentication successed")
		}

	}

}

func Test_Sm4gcm(t *testing.T) {
	key := []byte("airiot1234567890")
	//data := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10}
	str := "dell123"
	data := []byte(str)
	IV := make([]byte, sm4.BlockSize)
	if _, err := io.ReadFull(rand.Reader, IV); err != nil {
		t.Fatal(err)
	}
	data = pkcs7Padding(data, sm4.BlockSize)
	gcmMsg, T, err := sm4.Sm4GCM(key, IV, data, nil, true)
	if err != nil {
		t.Errorf("sm4 enc error:%s", err)
	}

	fmt.Printf("gcmMsg = %x\n", gcmMsg)
	fmt.Printf("gcmMsg T = %x\n", T)
	gcmDec, T_, err := sm4.Sm4GCM(key, IV, gcmMsg, nil, false)
	if err != nil {
		t.Errorf("sm4 dec error:%s", err)
	}
	gcmDec, err = pkcs7UnPadding(gcmDec)
	if err != nil {
		t.Errorf("sm4 pkcs7UnPadding error:%s", err)
	}
	fmt.Printf("gcmDec = %x\n", gcmDec)
	fmt.Printf("gcmDec = %s\n", string(gcmDec))
	fmt.Printf("gcmDec T= %x\n", T_)
}

func TestGCM(t *testing.T) {
	key := []byte("airiot1234567890")
	t.Log("key", string(key))
	data := []byte("dell123")

	// 加密示例
	encrypted, err := Encrypt(data, key, AES, GCM)
	if err != nil {
		panic(err)
	}
	t.Log("AES GCM", encrypted)
	// 解密示例
	decrypted, err := Decrypt(encrypted, key, AES, GCM)
	if err != nil {
		panic(err)
	}

	t.Log("AES GCM", string(decrypted))
	// 加密示例
	encrypted, err = Encrypt(data, key, SM4, GCM)
	if err != nil {
		panic(err)
	}
	t.Log("SM4 GCM", encrypted)
	// 解密示例
	decrypted, err = Decrypt(encrypted, key, SM4, GCM)
	if err != nil {
		panic(err)
	}

	t.Log("SM4 GCM", string(decrypted))
}

func TestCTR(t *testing.T) {
	key := []byte("airiot1234567890")
	t.Log("key", string(key))
	data := []byte("dell123")

	// 加密示例
	encrypted, err := Encrypt(data, key, AES, CTR)
	if err != nil {
		panic(err)
	}
	t.Log("AES CTR", encrypted)
	// 解密示例
	decrypted, err := Decrypt(encrypted, key, AES, CTR)
	if err != nil {
		panic(err)
	}

	t.Log("AES CTR", string(decrypted))
	// 加密示例
	encrypted, err = Encrypt(data, key, SM4, CTR)
	if err != nil {
		panic(err)
	}
	t.Log("SM4 CTR", encrypted)
	// 解密示例
	decrypted, err = Decrypt(encrypted, key, SM4, CTR)
	if err != nil {
		panic(err)
	}

	t.Log("SM4 CTR", string(decrypted))
}

func TestOFB(t *testing.T) {
	key := []byte("airiot1234567890")
	t.Log("key", string(key))
	data := []byte("dell123")

	// 加密示例
	encrypted, err := Encrypt(data, key, AES, OFB)
	if err != nil {
		panic(err)
	}
	t.Log("AES OFB", encrypted)
	// 解密示例
	decrypted, err := Decrypt(encrypted, key, AES, OFB)
	if err != nil {
		panic(err)
	}

	t.Log("AES OFB", string(decrypted))
	// 加密示例
	encrypted, err = Encrypt(data, key, SM4, OFB)
	if err != nil {
		panic(err)
	}
	t.Log("SM4 OFB", encrypted)
	// 解密示例
	decrypted, err = Decrypt(encrypted, key, SM4, OFB)
	if err != nil {
		panic(err)
	}

	t.Log("SM4 OFB", string(decrypted))
}

func TestCFB(t *testing.T) {
	key := []byte("airiot1234567890")
	t.Log("key", string(key))
	data := []byte("dell123")

	// 加密示例
	encrypted, err := Encrypt(data, key, AES, CFB)
	if err != nil {
		panic(err)
	}
	t.Log("AES CFB", encrypted)
	// 解密示例
	decrypted, err := Decrypt(encrypted, key, AES, CFB)
	if err != nil {
		panic(err)
	}

	t.Log("AES CFB", string(decrypted))
	// 加密示例
	encrypted, err = Encrypt(data, key, SM4, CFB)
	if err != nil {
		panic(err)
	}
	t.Log("SM4 CFB", encrypted)
	// 解密示例
	decrypted, err = Decrypt(encrypted, key, SM4, CFB)
	if err != nil {
		panic(err)
	}

	t.Log("SM4 CFB", string(decrypted))
}
