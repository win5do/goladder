package ss

import (
	"crypto/aes"
	"crypto/md5"
	"crypto/cipher"
	"io"
	"crypto/rand"
)

const (
	KEY_LEN = 16
)

type scipher struct {
	key       []byte
	iv        []byte
	encStream cipher.Stream
	decStream cipher.Stream
}

func hashKey(key string) []byte {
	h := md5.New()
	return h.Sum([]byte(key))[8:24]
}

func NewScipher(key, iv []byte) (*scipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encStream := cipher.NewCFBEncrypter(block, iv)
	decStream := cipher.NewCFBDecrypter(block, iv)

	return &scipher{
		key,
		iv,
		encStream,
		decStream,
	}, nil
}

// 随机初始向量
func randIv() ([]byte) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	return iv
}

func (scipher *scipher) encrypt(dst, src []byte) {
	scipher.encStream.XORKeyStream(dst, src)
}

func (scipher *scipher) decrypt(dst, src []byte) {
	scipher.decStream.XORKeyStream(dst, src)
}

// delete
func (scipher *scipher) initEncrypt() (err error) {
	block, err := aes.NewCipher(scipher.key)
	if err != nil {
		return err
	}

	scipher.encStream = cipher.NewCFBEncrypter(block, scipher.iv)
	return
}

func (scipher *scipher) initDecrypt() (err error) {
	block, err := aes.NewCipher(scipher.key)
	if err != nil {
		return err
	}

	scipher.decStream = cipher.NewCFBEncrypter(block, scipher.iv)
	return
}
