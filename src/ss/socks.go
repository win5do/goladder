package ss

import (
	"net"
	"io"
)

type sconn struct {
	net.Conn
	*scipher
}

func newSconn(conn net.Conn, key string, iv []byte) (*sconn, error) {
	k := hashKey(key)

	scipher, err := NewScipher(k, iv)
	if err != nil {
		return nil, err
	}

	return &sconn{
		conn,
		scipher,
	}, nil
}

// 加密写入
func (sconn *sconn) encryptWrite(b []byte) (int, error) {
	sconn.encrypt(b, b)
	return sconn.Write(b)
}

// 解密读入
func (sconn *sconn) decryptRead(b []byte) (n int, err error) {
	n, err = sconn.Read(b)
	sconn.decrypt(b[:n], b[:n])
	return
}

// 最少读 same as io.ReadAtLeast
func (sconn *sconn) decryptReadAtLeast(b []byte, min int) (n int, err error) {
	if len(b) < min {
		return 0, io.ErrShortBuffer
	}

	for n < min {
		add, er := sconn.decryptRead(b)
		n += add
		if er != nil {
			err = er
			return
		}
	}
	return
}

// 满读 same as io.ReadFull
func (sconn *sconn) decryptReadFull(painBuf []byte) (int, error) {
	return sconn.decryptReadAtLeast(painBuf, len(painBuf))
}

// 加密复制
func encryptCopy(dst *sconn, src net.Conn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.Read(buf)
		if err != nil {
			return n, err
		}

		if nr > 0 {
			// 写入
			nw, err := dst.encryptWrite(buf[:nr])
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
func decryptCopy(dst net.Conn, src *sconn) (n int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		// 读取
		nr, err := src.decryptRead(buf)
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
