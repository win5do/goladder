package core

import (
	"crypto/aes"
	"crypto/cipher"
	"bytes"
)

func AesEncrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	src = PKCS5Padding(src, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	enc := make([]byte, len(src))
	blockMode.CryptBlocks(enc, src)
	return enc, nil
}

func AesDecrypt(enc, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	src := make([]byte, len(enc))
	blockMode.CryptBlocks(src, enc)
	src = PKCS5UnPadding(src)
	return src, nil
}

func PKCS5Padding(enc []byte, blockSize int) []byte {
	// 只要少于256就能放到一个block中，默认的blockSize=16(即采用16*8=128, AES-128长的密钥)
	// 最少填充1个byte，如果原文刚好是blocksize的整数倍，则再填充一个blocksize
	padding := blockSize - len(enc)%blockSize        //需要padding的数目
	padtext := bytes.Repeat([]byte{byte(padding)}, padding) //生成填充的文本
	return append(enc, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}
