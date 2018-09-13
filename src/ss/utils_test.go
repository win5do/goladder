package ss

import (
	"bufio"
	"bytes"
	"testing"
)

func TestIsHttp(t *testing.T) {
	cs := []struct {
		in  string
		out bool
	}{
		{in: "GET /books/?sex=man&name=Professional HTTP/1.1\n", out: true},
		{in: "POST / HTTP/1.1\n", out: true},
	}

	for _, c := range cs {
		buf := bytes.NewBufferString(c.in)
		bufrd := bufio.NewReader(buf)
		ok := IsHttp(bufrd)
		t.Log(ok)
	}
}
