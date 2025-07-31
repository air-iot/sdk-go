package cipherx

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/tjfoc/gmsm/sm4"
)

// CipherType 加密算法类型
type CipherType string

const (
	AES CipherType = "AES"
	SM4 CipherType = "SM4"
)

// Mode 加密模式类型
type Mode string

const (
	CBC Mode = "CBC"
	ECB Mode = "ECB"
	CTR Mode = "CTR"
	GCM Mode = "GCM"
	OFB Mode = "OFB"
	CFB Mode = "CFB"
)

// Encrypt 通用加密函数
func Encrypt(data []byte, key []byte, cipherType CipherType, mode Mode) (string, error) {
	var block cipher.Block
	var err error

	// 根据加密算法类型创建对应的块
	switch cipherType {
	case AES:
		block, err = aes.NewCipher(key)
	case SM4:
		block, err = sm4.NewCipher(key)
	default:
		return "", errors.New("unsupported cipher type")
	}
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()

	switch mode {
	case ECB:
		// ECB 模式不需要 IV
		data = pkcs7Padding(data, blockSize)
		crypted := make([]byte, len(data))
		for bs, be := 0, blockSize; bs < len(data); bs, be = bs+blockSize, be+blockSize {
			block.Encrypt(crypted[bs:be], data[bs:be])
		}
		return base64.StdEncoding.EncodeToString(crypted), nil

	case CBC:
		data = pkcs7Padding(data, blockSize)
		crypted := make([]byte, blockSize+len(data))
		iv := crypted[:blockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return "", err
		}

		cbcMode := cipher.NewCBCEncrypter(block, iv)
		cbcMode.CryptBlocks(crypted[blockSize:], data)
		return base64.StdEncoding.EncodeToString(crypted), nil

	case CTR:
		crypted := make([]byte, blockSize+len(data))
		iv := crypted[:blockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return "", err
		}

		var stream cipher.Stream
		if cipherType == SM4 {
			// SM4 的 CTR 模式需要特殊处理
			stream = NewSM4CTR(block, iv)
		} else {
			stream = cipher.NewCTR(block, iv)
		}
		stream.XORKeyStream(crypted[blockSize:], data)
		return base64.StdEncoding.EncodeToString(crypted), nil

	case GCM:
		if cipherType == SM4 {
			nonce := make([]byte, sm4.BlockSize)
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				return "", err
			}
			data = pkcs7Padding(data, sm4.BlockSize)
			ciphertext, _, err := sm4.Sm4GCM(key, nonce, data, nil, true)
			if err != nil {
				return "", err
			}
			// 组合 nonce + ciphertext 用于传输
			encryptedData := append(nonce, ciphertext...)
			return base64.StdEncoding.EncodeToString(encryptedData), nil
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			return "", err
		}

		crypted := gcm.Seal(nonce, nonce, data, nil)
		return base64.StdEncoding.EncodeToString(crypted), nil

	case OFB:
		crypted := make([]byte, blockSize+len(data))
		iv := crypted[:blockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return "", err
		}

		stream := cipher.NewOFB(block, iv)
		stream.XORKeyStream(crypted[blockSize:], data)
		return base64.StdEncoding.EncodeToString(crypted), nil

	case CFB:
		crypted := make([]byte, blockSize+len(data))
		iv := crypted[:blockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return "", err
		}

		stream := cipher.NewCFBEncrypter(block, iv)
		stream.XORKeyStream(crypted[blockSize:], data)
		return base64.StdEncoding.EncodeToString(crypted), nil

	default:
		return "", errors.New("unsupported encryption mode")
	}
}

// Decrypt 通用解密函数
func Decrypt(encrypted string, key []byte, cipherType CipherType, mode Mode) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	var block cipher.Block
	switch cipherType {
	case AES:
		block, err = aes.NewCipher(key)
	case SM4:
		block, err = sm4.NewCipher(key)
	default:
		return nil, errors.New("unsupported cipher type")
	}
	if err != nil {
		return nil, err
	}

	switch mode {
	case ECB:
		blockSize := block.BlockSize()
		if len(data)%blockSize != 0 {
			return nil, errors.New("ciphertext is not a multiple of the block size")
		}

		decrypted := make([]byte, len(data))
		for bs, be := 0, blockSize; bs < len(data); bs, be = bs+blockSize, be+blockSize {
			block.Decrypt(decrypted[bs:be], data[bs:be])
		}
		return pkcs7UnPadding(decrypted)

	case CBC:
		blockSize := block.BlockSize()
		if len(data) < blockSize {
			return nil, errors.New("ciphertext too short")
		}

		iv := data[:blockSize]
		data = data[blockSize:]

		if len(data)%blockSize != 0 {
			return nil, errors.New("ciphertext is not a multiple of the block size")
		}

		cbcMode := cipher.NewCBCDecrypter(block, iv)
		decrypted := make([]byte, len(data))
		cbcMode.CryptBlocks(decrypted, data)
		return pkcs7UnPadding(decrypted)

	case CTR:
		blockSize := block.BlockSize()
		if len(data) < blockSize {
			return nil, errors.New("ciphertext too short")
		}

		iv := data[:blockSize]
		data = data[blockSize:]

		var stream cipher.Stream
		if cipherType == SM4 {
			stream = NewSM4CTR(block, iv)
		} else {
			stream = cipher.NewCTR(block, iv)
		}
		decrypted := make([]byte, len(data))
		stream.XORKeyStream(decrypted, data)
		return decrypted, nil

	case GCM:
		if cipherType == SM4 {
			blockSize := sm4.BlockSize
			if len(data) < blockSize {
				return nil, errors.New("ciphertext too short")
			}

			// 分离 nonce 和 ciphertext
			nonce := data[:blockSize]
			ciphertext := data[blockSize:]
			// 使用官方 Sm4GCM 接口解密
			plaintext, _, err := sm4.Sm4GCM(key, nonce, ciphertext, nil, false)
			if err != nil {
				return nil, err
			}
			return pkcs7UnPadding(plaintext)
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		nonceSize := gcm.NonceSize()
		if len(data) < nonceSize {
			return nil, errors.New("ciphertext too short")
		}

		nonce, ciphertext := data[:nonceSize], data[nonceSize:]
		return gcm.Open(nil, nonce, ciphertext, nil)

	case OFB:
		blockSize := block.BlockSize()
		if len(data) < blockSize {
			return nil, errors.New("ciphertext too short")
		}

		iv := data[:blockSize]
		data = data[blockSize:]

		stream := cipher.NewOFB(block, iv)
		decrypted := make([]byte, len(data))
		stream.XORKeyStream(decrypted, data)
		return decrypted, nil

	case CFB:
		blockSize := block.BlockSize()
		if len(data) < blockSize {
			return nil, errors.New("ciphertext too short")
		}

		iv := data[:blockSize]
		data = data[blockSize:]

		stream := cipher.NewCFBDecrypter(block, iv)
		decrypted := make([]byte, len(data))
		stream.XORKeyStream(decrypted, data)
		return decrypted, nil

	default:
		return nil, errors.New("unsupported encryption mode")
	}
}

// SM4CTR 实现 SM4 的 CTR 模式
type SM4CTR struct {
	block     cipher.Block
	ctr       []byte
	out       []byte
	outUsed   int
	blockSize int
}

func NewSM4CTR(block cipher.Block, iv []byte) *SM4CTR {
	return &SM4CTR{
		block:     block,
		ctr:       bytes.Clone(iv),
		out:       make([]byte, block.BlockSize()),
		outUsed:   block.BlockSize(),
		blockSize: block.BlockSize(),
	}
}

func (x *SM4CTR) XORKeyStream(dst, src []byte) {
	for i := 0; i < len(src); i++ {
		if x.outUsed == x.blockSize {
			x.block.Encrypt(x.out, x.ctr)

			// 递增计数器
			for j := len(x.ctr) - 1; j >= 0; j-- {
				x.ctr[j]++
				if x.ctr[j] != 0 {
					break
				}
			}

			x.outUsed = 0
		}

		dst[i] = src[i] ^ x.out[x.outUsed]
		x.outUsed++
	}
}

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding 去除填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data")
	}

	unPadding := int(data[length-1])
	if unPadding > length {
		return nil, errors.New("invalid padding")
	}

	return data[:(length - unPadding)], nil
}
