package varint

/*

From SPOE spec:

Variable-length integer (varint) are encoded using Peers encoding:


       0  <= X < 240        : 1 byte  (7.875 bits)  [ XXXX XXXX ]
      240 <= X < 2288       : 2 bytes (11 bits)     [ 1111 XXXX ] [ 0XXX XXXX ]
     2288 <= X < 264432     : 3 bytes (18 bits)     [ 1111 XXXX ] [ 1XXX XXXX ]   [ 0XXX XXXX ]
   264432 <= X < 33818864   : 4 bytes (25 bits)     [ 1111 XXXX ] [ 1XXX XXXX ]*2 [ 0XXX XXXX ]
 33818864 <= X < 4328786160 : 5 bytes (32 bits)     [ 1111 XXXX ] [ 1XXX XXXX ]*3 [ 0XXX XXXX ]
 ...

*/

func PutUvarint(buf []byte, n uint64) int {
	var p int

	if len(buf) == 0 {
		return -1
	}

	if n < 240 {
		buf[p] = byte(n)
		return 1
	}

	buf[p] = byte(n) | 0xF0

	p++

	n = (n - 240) >> 4

	for n >= 128 {
		if p >= len(buf) {
			return -1
		}

		buf[p] = byte(n) | 128
		n = (n - 128) >> 7

		p++
	}
	if p >= len(buf) {
		return -1
	}

	buf[p] = byte(n)

	return p + 1
}

func Uvarint(buf []byte) (uint64, int) {
	var p int

	if len(buf) == 0 {
		return 0, -1
	}

	n := uint64(buf[p])

	if n < 240 {
		return n, 1
	}

	r := uint(4)

	for {
		p++
		if p >= len(buf) {
			return 0, -1
		}
		n += uint64(buf[p]) << r
		r += 7
		if int64(buf[p]) < 128 {
			break
		}
	}

	return n, p + 1
}
