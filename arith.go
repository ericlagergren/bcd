package bcd

const (
	sixmask   = 0x6666666666666666
	eightmask = 0x8888888888888888
)

// Addition and subtraction algorithms are from Knuth's TAOCP Vol. 4A, part 1.

func median(x, y, z Word) Word { return (x & (y | z)) | (y & z) }

func addVV(z, x, y []Word) (c Word) {
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

func addVW(z, x []Word, y Word) (c Word) {
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

func subVV(z, x, y []Word) (c Word) {
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

func subVW(z, x []Word, y Word) (c Word) {
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
