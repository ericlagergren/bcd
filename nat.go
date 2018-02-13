package bcd

import (
	"math/bits"
)

const (
	_S = _W / 8 // word size in bytes

	_W = bits.UintSize // word size in bits
	_B = 1 << _W       // digit base
	_M = _B - 1        // digit mask

	_W2 = _W / 2   // half word size in bits
	_B2 = 1 << _W2 // half digit base
	_M2 = _B2 - 1  // half digit mask

	_W4 = _W2 / 2 // quarter word size in bits
)

type Word uint

type nat []Word

func (z nat) make(n int) nat {
	if n <= cap(z) {
		return z[:n]
	}
	const e = 4
	return make(nat, n, n+e)
}

func (z nat) set(x nat) nat {
	z = z.make(len(x))
	copy(z, x)
	return z
}

func (z nat) norm() nat {
	i := len(z)
	for i > 0 && z[i-1] == 0 {
		i--
	}
	return z[0:i]
}

func (z nat) setWord(x Word) nat {
	if x == 0 {
		return z[:0]
	}
	z = z.make(1)
	z[0] = x
	return z
}

func (z nat) setUint64(x uint64) nat {
	if x == 0 {
		return z[:0]
	}

	const maxWord = 999999999999999
	n := 1
	if x >= maxWord {
		n = 2
	}

	z = z.make(n)
	var shift Word
	for j := 0; j < n; j++ {
		for i := 0; i < 16; i++ {
			z[j] |= Word(x % 10 << shift)
			shift += 4
			x /= 10
			if x == 0 {
				break
			}
		}
		shift = 0
	}
	return z
}

func (z nat) setString(s string) nat {
	if s == "" {
		panic(`setString: s == ""`)
	}
	z = z.make((len(s) + 15) / 16)
	var (
		k     = len(s) - 1
		shift Word
	)
	for i := range z {
		for j := 0; j < 16; j++ {
			z[i] |= Word(s[k]-'0') << shift
			shift += 4
			k--
			if k < 0 {
				break
			}
		}
		shift = 0
	}
	return z
}

func (z nat) add(x, y nat) nat {
	m := len(x)
	n := len(y)

	switch {
	case m < n:
		return z.add(y, x)
	case m == 0:
		// n == 0 because m >= n; result is 0
		return z[:0]
	case n == 0:
		// result is x
		return z.set(x)
	}
	// m > 0

	z = z.make(m + 1)
	c := addVV(z[0:n], x, y)
	if m > n {
		c = addVW(z[n:m], x[n:], c)
	}
	z[m] = c

	return z.norm()
}

func (z nat) sub(x, y nat) nat {
	m := len(x)
	n := len(y)

	switch {
	case m < n:
		panic("underflow")
	case m == 0:
		// n == 0 because m >= n; result is 0
		return z[:0]
	case n == 0:
		// result is x
		return z.set(x)
	}
	// m > 0

	z = z.make(m)
	c := subVV(z[0:n], x, y)
	if m > n {
		c = subVW(z[n:], x[n:], c)
	}
	if c != 0 {
		panic("underflow")
	}

	return z.norm()
}

func (x nat) cmp(y nat) (r int) {
	m := len(x)
	n := len(y)
	if m != n || m == 0 {
		switch {
		case m < n:
			r = -1
		case m > n:
			r = 1
		}
		return
	}

	i := m - 1
	for i > 0 && x[i] == y[i] {
		i--
	}

	switch {
	case x[i] < y[i]:
		r = -1
	case x[i] > y[i]:
		r = 1
	}
	return
}

func (z nat) append(b []byte) []byte {
	if len(z) == 0 {
		return []byte{'0'}
	}

	n := len(z) * 16
	if cap(b) < n {
		b = make([]byte, n)
	} else {
		b = b[:n]
	}

	for _, w := range z[:len(z)-1] {
		for i := 0; i < 16; i++ {
			n--
			b[n] = byte(w%16) + '0'
			w >>= 4
		}
	}
	for w := z[len(z)-1]; w != 0; w >>= 4 {
		n--
		b[n] = byte(w%16) + '0'
	}
	return b[n:]
}

func (z nat) String() string {
	if len(z) == 0 {
		return "0"
	}
	return string(z.append(nil))
}
