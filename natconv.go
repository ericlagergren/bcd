package bcd

func (z nat) setString(s string) nat {
	if s == "" {
		panic(`setString: s == ""`)
	}
	if s == "0" {
		return z[:0]
	}

	z = z.make((len(s) + 15) / 16)
	k := len(s) - 1
	for i := range z {
		for shift := Word(0); shift < _W; shift += 4 {
			z[i] |= Word(s[k]-'0') << shift
			k--
			if k < 0 {
				break
			}
		}
	}
	return z
}

func (z nat) append(b []byte) []byte {
	// http://www.hackersdelight.org/corres.txt has some routines for converting
	// to/from ASCII.
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
