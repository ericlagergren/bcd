package bcd

import (
	"fmt"
	"math/bits"
)

// A Word represents a single digit of a multi-precision unsigned integer.
type Word uint

const (
	_S = _W / 8 // word size in bytes

	_W = bits.UintSize // word size in bits
	_B = 1 << _W       // digit base
	_M = _B - 1        // digit mask

	_W2 = _W / 2   // half word size in bits
	_B2 = 1 << _W2 // half digit base
	_M2 = _B2 - 1  // half digit mask
)

const (
	maxWord   = 999999999999999
	sixmask   = 0x6666666666666666
	eightmask = 0x8888888888888888
)

// bin converts a BCD word to a binary word.
func bin(w Word) Word {
	z := (w & 0xF)
	z += ((w >> 4) & 0xF) * 10
	z += ((w >> 8) & 0xF) * 100
	z += ((w >> 12) & 0xF) * 1000
	z += ((w >> 16) & 0xF) * 10000
	z += ((w >> 20) & 0xF) * 100000
	z += ((w >> 24) & 0xF) * 1000000
	z += ((w >> 28) & 0xF) * 10000000
	z += ((w >> 32) & 0xF) * 100000000
	z += ((w >> 36) & 0xF) * 1000000000
	z += ((w >> 40) & 0xF) * 10000000000
	z += ((w >> 44) & 0xF) * 100000000000
	z += ((w >> 48) & 0xF) * 1000000000000
	z += ((w >> 52) & 0xF) * 10000000000000
	z += ((w >> 56) & 0xF) * 100000000000000
	z += ((w >> 60) & 0xF) * 1000000000000000
	return z
}

// bcd converts a binary word to a BCD word.
func bcd(x Word) (z Word) {
	if x > maxWord {
		panic(fmt.Sprintf("%d > maxWord", x))
	}
	for shift := Word(0); shift < _W; shift += 4 {
		z |= Word(x%10) << shift
		x /= 10
		if x == 0 {
			break
		}
	}
	return z
}

// ----------------------------------------------------------------------------
// Elementary operations on words
//
// These operations are used by the vector operations below.

func addWW_bcd_g(x, y, c Word) (z1, z0 Word) {
	y += sixmask
	z0 = x + y + c
	t := ((^x & (^y | z0)) | (^y & z0)) & eightmask
	z0 -= t + t>>2
	if z0 < x {
		z1 = 1
	}
	return
}

// z1<<_W + z0 = x+y+c, with c == 0 or 1
func addWW_g(x, y, c Word) (z1, z0 Word) {
	yc := y + c
	z0 = x + yc
	if z0 < x || yc < y {
		z1 = 1
	}
	return
}

func subWW_bcd_g(x, y, c Word) (z1, z0 Word) {
	z0 = x - y - c
	t := ((^x & (y | z0)) | (y & z0)) & eightmask
	z0 -= t + t>>2
	if z0 < y {
		z1 = 1
	}
	return
}

// z1<<_W + z0 = x-y-c, with c == 0 or 1
func subWW_g(x, y, c Word) (z1, z0 Word) {
	yc := y + c
	z0 = x - yc
	if z0 > x || yc < y {
		z1 = 1
	}
	return
}

func mulWW_bcd_g(x, y Word) (z1, z0 Word) {
	return conv128(mulWW_g(bin(x), bin(y)))
}

// z1<<_W + z0 = x*y
// Adapted from Warren, Hacker's Delight, p. 132.
func mulWW_g(x, y Word) (z1, z0 Word) {
	x0 := x & _M2
	x1 := x >> _W2
	y0 := y & _M2
	y1 := y >> _W2
	w0 := x0 * y0
	t := x1*y0 + w0>>_W2
	w1 := t & _M2
	w2 := t >> _W2
	w1 += x0 * y1
	z1 = x1*y1 + w2 + w1>>_W2
	z0 = x * y
	return
}

// z1<<_W + z0 = x*y + c
func mulAddWWW_bcd_g(x, y, c Word) (z1, z0 Word) {
	z1, zz0 := conv128(mulWW_g(bin(x), bin(y)))
	if z0 = zz0 + c; z0 < zz0 {
		z1++
	}
	return
}

// z1<<_W + z0 = x*y + c
func mulAddWWW_g(x, y, c Word) (z1, z0 Word) {
	z1, zz0 := mulWW_g(x, y)
	if z0 = zz0 + c; z0 < zz0 {
		z1++
	}
	return
}

// nlz returns the number of leading zeros in x.
// Wraps bits.LeadingZeros call for convenience.
func nlz(x Word) uint {
	return uint(bits.LeadingZeros(uint(x)))
}

func divWW_bcd_g(u1, u0, v Word) (q, r Word) {
	return divWW(bin(u1), bin(u0), bin(v))
}

// q = (u1<<_W + u0 - r)/y
// Adapted from Warren, Hacker's Delight, p. 152.
func divWW_g(u1, u0, v Word) (q, r Word) {
	if u1 >= v {
		return 1<<_W - 1, 1<<_W - 1
	}

	s := nlz(v)
	v <<= s

	vn1 := v >> _W2
	vn0 := v & _M2
	un32 := u1<<s | u0>>(_W-s)
	un10 := u0 << s
	un1 := un10 >> _W2
	un0 := un10 & _M2
	q1 := un32 / vn1
	rhat := un32 - q1*vn1

	for q1 >= _B2 || q1*vn0 > _B2*rhat+un1 {
		q1--
		rhat += vn1
		if rhat >= _B2 {
			break
		}
	}

	un21 := un32*_B2 + un1 - q1*v
	q0 := un21 / vn1
	rhat = un21 - q0*vn1

	for q0 >= _B2 || q0*vn0 > _B2*rhat+un0 {
		q0--
		rhat += vn1
		if rhat >= _B2 {
			break
		}
	}

	return q1*_B2 + q0, (un21*_B2 + un0 - q0*v) >> s
}

// Keep for performance debugging.
// Using addWW_g is likely slower.
const use_addWW_g = false

func addVV_bcd_g(z, x, y []Word) (c Word) {
	for i, xi := range x[:len(z)] {
		yi := y[i] + sixmask
		zi := xi + yi + c
		t := ((^xi & (^yi | zi)) | (^yi & zi)) & eightmask
		z[i] = zi - t + t>>2
		if zi < xi {
			c = 1
		} else {
			c = 0
		}
	}
	return c
}

// The resulting carry c is either 0 or 1.
func addVV_g(z, x, y []Word) (c Word) {
	if use_addWW_g {
		for i := range z {
			c, z[i] = addWW_g(x[i], y[i], c)
		}
		return
	}

	for i, xi := range x[:len(z)] {
		yi := y[i]
		zi := xi + yi + c
		z[i] = zi
		// see "Hacker's Delight", section 2-12 (overflow detection)
		c = (xi&yi | (xi|yi)&^zi) >> (_W - 1)
	}
	return
}

func subVV_bcd_g(z, x, y []Word) (c Word) {
	for i, xi := range x[:len(z)] {
		yi := y[i]
		zi := xi - yi - c
		t := ((^xi & (yi | zi)) | (yi & zi)) & eightmask
		z[i] = zi - t + t>>2
		if xi < yi {
			c = 1
		} else {
			c = 0
		}
	}
	return c
}

// The resulting carry c is either 0 or 1.
func subVV_g(z, x, y []Word) (c Word) {
	if use_addWW_g {
		for i := range z {
			c, z[i] = subWW_g(x[i], y[i], c)
		}
		return
	}

	for i, xi := range x[:len(z)] {
		yi := y[i]
		zi := xi - yi - c
		z[i] = zi
		// see "Hacker's Delight", section 2-12 (overflow detection)
		c = (yi&^xi | (yi|^xi)&zi) >> (_W - 1)
	}
	return
}

func addVW_bcd_g(z, x []Word, y Word) (c Word) {
	c = y
	for i, xi := range x[:len(z)] {
		zi := xi + c
		t := (xi &^ zi) & eightmask
		z[i] = zi - t + t>>2
		if zi < xi {
			c = 1
		} else {
			c = 0
		}
	}
	return c
}

// The resulting carry c is either 0 or 1.
func addVW_g(z, x []Word, y Word) (c Word) {
	if use_addWW_g {
		c = y
		for i := range z {
			c, z[i] = addWW_g(x[i], c, 0)
		}
		return
	}

	c = y
	for i, xi := range x[:len(z)] {
		zi := xi + c
		z[i] = zi
		c = xi &^ zi >> (_W - 1)
	}
	return
}

func subVW_bcd_g(z, x []Word, y Word) (c Word) {
	c = y
	for i, xi := range x[:len(z)] {
		zi := xi - c
		t := (xi &^ zi) & eightmask
		z[i] = zi - t + t>>2
		if xi < c {
			c = 1
		} else {
			c = 0
		}
	}
	return c
}

func subVW_g(z, x []Word, y Word) (c Word) {
	if use_addWW_g {
		c = y
		for i := range z {
			c, z[i] = subWW_g(x[i], c, 0)
		}
		return
	}

	c = y
	for i, xi := range x[:len(z)] {
		zi := xi - c
		z[i] = zi
		c = (zi &^ xi) >> (_W - 1)
	}
	return
}

func shlVU_g(z, x []Word, s uint) (c Word) {
	if n := len(z); n > 0 {
		ŝ := _W - s
		w1 := x[n-1]
		c = w1 >> ŝ
		for i := n - 1; i > 0; i-- {
			w := w1
			w1 = x[i-1]
			z[i] = w<<s | w1>>ŝ
		}
		z[0] = w1 << s
	}
	return
}

func shrVU_g(z, x []Word, s uint) (c Word) {
	if n := len(z); n > 0 {
		ŝ := _W - s
		w1 := x[0]
		c = w1 << ŝ
		for i := 0; i < n-1; i++ {
			w := w1
			w1 = x[i+1]
			z[i] = w>>s | w1<<ŝ
		}
		z[n-1] = w1 >> s
	}
	return
}

func mulAddVWW_bcd_g(z, x []Word, y, r Word) (c Word) {
	c = r
	for i := range z {
		c, z[i] = mulAddWWW_bcd_g(x[i], y, c)
	}
	return
}

func mulAddVWW_g(z, x []Word, y, r Word) (c Word) {
	c = r
	for i := range z {
		c, z[i] = mulAddWWW_g(x[i], y, c)
	}
	return
}

func addMulVVW_bcd_g(z, x []Word, y Word) (c Word) {
	for i := range z {
		z1, z0 := mulAddWWW_g(x[i], y, z[i])
		c, z[i] = addWW_bcd_g(z0, c, 0)
		c += z1
	}
	return
}

// TODO(gri) Remove use of addWW_g here and then we can remove addWW_g and subWW_g.
func addMulVVW_g(z, x []Word, y Word) (c Word) {
	for i := range z {
		z1, z0 := mulAddWWW_g(x[i], y, z[i])
		c, z[i] = addWW_g(z0, c, 0)
		c += z1
	}
	return
}

func divWVW_bcd_g(z []Word, xn Word, x []Word, y Word) (r Word) {
	// Don't call divBCDWW since y is static and we only need to convert the
	// initial r.
	r = bin(xn)
	y = bin(y)
	var zi Word
	for i := len(z) - 1; i >= 0; i-- {
		xi := bin(x[i])
		zi, r = divWW_g(r, xi, y)
		z[i] = bcd(zi)
	}
	return
}

func divWVW_g(z []Word, xn Word, x []Word, y Word) (r Word) {
	r = xn
	for i := len(z) - 1; i >= 0; i-- {
		z[i], r = divWW_g(r, x[i], y)
	}
	return
}
