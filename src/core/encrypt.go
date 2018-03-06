package core

import (
	"crypto/aes"
	"crypto/md5"
	"crypto/rand"
	"io"
	"crypto/cipher"
	"net"
	"errors"
	"log"
)

type sconn struct {
	key []byte
	net.Conn
}

func newSconn(key string, conn net.Conn) sconn {
	k := hashKey(key)
	return sconn{
		k,
		conn,
	}
}

func hashKey(key string) []byte {
	h := md5.New()
	return h.Sum([]byte(key))[8:24]
}

func encrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, aes.BlockSize+len(src))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], src)
	return encrypted, nil
}

func decrypt(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(encrypted) < aes.BlockSize {
		return nil, errors.New("密文太短")
	}

	iv := encrypted[:aes.BlockSize]
	src := encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(src, src)
	return src, nil
}

// 加密写入
func (src *sconn) encryptWrite(painText []byte) (int, error) {
	encrypted, err := encrypt(painText, src.key)
	if err != nil {
		return 0, err
	}
	return src.Write(encrypted)
}

// 解密读入
func (src *sconn) decryptRead(painText []byte) (n int, err error) {
	encrypted := make([]byte, 32*1024)
	n, err = src.Read(encrypted)
	if err != nil {
		return
	} else {
		var b []byte
		b, err = decrypt(encrypted[:n], src.key)
		copy(painText, b)
		return len(b), err
	}
}

// 加密复制
func (src *sconn) encryptCopy(dst sconn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		if nr > 0 {
			// 写入
			nw, err := dst.encryptWrite(buf[:nr])
			if err != nil {
				log.Println(err)
				break
			}

			if nw > 0 {
				n += int64(nw)
			}
		}
	}

	return n, err
}

// 解密复制
func (src *sconn) decryptCopy(dst sconn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.decryptRead(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		if nr > 0 {
			// 写入
			nw, err := dst.Write(buf[0:nr])
			if err != nil {
				log.Println(err)
				break
			}

			if nw > 0 {
				n += int64(nw)
			}
		}
	}

	return n, err
}
