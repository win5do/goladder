package ss

import (
	"io"
	"net"
)

type Sconn struct {
	net.Conn
	*scipher
}

func newSconn(conn net.Conn, key string, iv []byte) (*Sconn, error) {
	k := hashKey(key)

	scipher, err := NewScipher(k, iv)
	if err != nil {
		return nil, err
	}

	return &Sconn{
		conn,
		scipher,
	}, nil
}

// 加密写入
func (sconn *Sconn) EncryptWrite(b []byte) (int, error) {
	sconn.encrypt(b, b)
	return sconn.Write(b)
}

// 解密读入
func (sconn *Sconn) DecryptRead(b []byte) (n int, err error) {
	n, err = sconn.Read(b)
	sconn.decrypt(b[:n], b[:n])
	return
}

// 最少读 same as io.ReadAtLeast
func (sconn *Sconn) DecryptReadAtLeast(b []byte, min int) (n int, err error) {
	if len(b) < min {
		return 0, io.ErrShortBuffer
	}

	for n < min {
		add, er := sconn.DecryptRead(b)
		n += add
		if er != nil {
			err = er
			return
		}
	}
	return
}

// 满读 same as io.ReadFull
func (sconn *Sconn) DecryptReadFull(painBuf []byte) (int, error) {
	return sconn.DecryptReadAtLeast(painBuf, len(painBuf))
}

// 加密复制
func EncryptCopy(dst *Sconn, src net.Conn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.Read(buf)
		if err != nil {
			return n, err
		}

		if nr > 0 {
			// 写入
			nw, err := dst.EncryptWrite(buf[:nr])
			if err != nil {
				return n, err
			}

			if nw > 0 {
				n += int64(nw)
			}
		}
	}

	return n, err
}

// 解密复制
func DecryptCopy(dst net.Conn, src *Sconn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.DecryptRead(buf)
		if err != nil {
			return n, err
		}

		if nr > 0 {
			// 写入
			nw, err := dst.Write(buf[0:nr])
			if err != nil {
				return n, err
			}

			if nw > 0 {
				n += int64(nw)
			}
		}
	}

	return n, err
}

// 和服务器建立sconn
func InitSconn(conn net.Conn, key string) (sconn *Sconn, err error) {
	// 随机一个iv 创建加密器
	iv := randIv()
	sconn, err = newSconn(conn, key, iv)
	if err != nil {
		return
	}
	// 先把iv发给服务器
	_, err = sconn.Write(iv)
	if err != nil {
		return
	}
	return
}
