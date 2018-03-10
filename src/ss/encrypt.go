package ss

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

const (
	KEY_LEN = 16
)

type sconn struct {
	net.Conn
	key []byte
}

func newSconn(conn net.Conn, key string) sconn {
	k := hashKey(key)
	return sconn{
		conn,
		k,
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

	encBuf := make([]byte, aes.BlockSize+len(src))
	iv := encBuf[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encBuf[aes.BlockSize:], src)
	return encBuf, nil
}

func decrypt(encBuf, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(encBuf) < aes.BlockSize {
		return nil, errors.New("密文太短")
	}

	iv := encBuf[:aes.BlockSize]
	src := encBuf[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(src, src)
	return src, nil
}

// 加密写入
func (src *sconn) encryptWrite(painBuf []byte) (int, error) {
	encBuf, err := encrypt(painBuf, src.key)
	if err != nil {
		return 0, err
	}
	return src.Write(encBuf)
}

// 解密读入
func (src *sconn) decryptRead(painBuf []byte) (n int, err error) {
	encBuf := make([]byte, len(painBuf)+KEY_LEN)
	n, err = src.Read(encBuf)
	if err != nil {
		return
	} else {
		var b []byte
		b, err = decrypt(encBuf[:n], src.key)
		copy(painBuf, b)
		return len(b), err
	}
}

// 最少读 same as io.ReadAtLeast
func (src *sconn) decryptReadAtLeast(painBuf []byte, min int) (n int, err error) {
	if len(painBuf) < min {
		return 0, io.ErrShortBuffer
	}

	for n < min {
		add, er := src.decryptRead(painBuf)
		n += add
		if er != nil {
			err = er
			return
		}
	}
	return
}

// 满读 same as io.ReadFull
func (src *sconn) decryptReadFull(painBuf []byte) (int, error) {
	return src.decryptReadAtLeast(painBuf, len(painBuf))
}

// 加密复制
func (src *sconn) encryptCopy(dst sconn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.Read(buf)
		if err != nil {
			// 一般为连接断开错误
			//if err != io.EOF {
			//	log.Println(err)
			//}
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
			//if err != io.EOF {
			//	log.Println(err)
			//}
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
