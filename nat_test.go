package bcd

import "testing"

func TestAddSub(t *testing.T) {
	// A mélange of tests, using setString and setUint64 as extra sanity checks.
	for i, c := range [...]struct {
		x, y, r nat
	}{
		0:  { /* 0 + 0 = 0 */ },
		1:  {nat{0x1}, nat{0x2}, nat{0x3}},
		2:  {nat{}.setString("100"), nat{}.setString("42"), nat{}.setString("142")},
		3:  {nat{}.setUint64(100), nat{}.setUint64(42), nat{}.setUint64(142)},
		4:  {nat{0x9999999999999998}, nat{0x1}, nat{0x9999999999999999}},
		5:  {nat{0x9999999999999989}, nat{0x1}, nat{0x9999999999999990}},
		6:  {nat{0x9999999999999999}, nat{0x1}, nat{0x0, 0x1}},
		7:  {nat{0x010203}, nat{0x030201}, nat{0x040404}},
		8:  {nat{0x01}, nat{0x0200}, nat{0x0201}},
		9:  {nat{0x4294967295}, nat{0x4294967295}, nat{0x8589934590}},
		10: {nat{0x6744073709551615, 0x1844}, nat{0x6744073709551615, 0x1844}, nat{0x3488147419103230, 0x3689}},
		11: {nat{0x6744073709551616, 0x1844}, nat{0x6744073709551615, 0x1844}, nat{0x3488147419103231, 0x3689}},
		12: {nat{0x3488147419103230, 0x3689}, nat{0x01}, nat{0x3488147419103231, 0x3689}},
		13: {nat{0x87778366101931, 0x110}, nat{0x9979416004714189, 0x177}, nat{0x67194370816120, 0x288}},
		14: {nat{}.setString("423784981374892374987312482374987123"), nat{}.setString("4231432142314321421349823484884840124"), nat{}.setString("4655217123689213796337135967259827247")},
		15: {nat{}.setString("1234567812345678"), nat{}.setString("10000000000000000"), nat{}.setString("11234567812345678")},
		16: {nat{}.setString("14472334024676221"), nat{}.setString("8944394323791464"), nat{}.setString("23416728348467685")},
	} {
		za := nat(nil).add(c.x, c.y)
		zs := nat(nil).sub(za, c.x)
		if za.cmp(c.r) != 0 || zs.cmp(c.y) != 0 {
			t.Fatalf(`#%d: %s + %s, %s - %s
wanted: %s, %s
got   : %s, %s
`, i, c.x, c.y, c.r, c.x, c.r, c.y, za, zs)
		}
	}
}

func TestMul(t *testing.T) {
	for i, c := range [...]struct {
		x, y, r string
	}{
		{"1", "12", "12"},
		{"4", "4", "16"},
		{"100", "100", "10000"},
		{"123124", "12332", "1518365168"},
		{"9999999999999999", "9999999999999999", "99999999999999980000000000000001"},
	} {
		x := nat(nil).setString(c.x)
		y := nat(nil).setString(c.y)
		r := nat(nil).setString(c.r)

		z := nat(nil).mul(x, y)
		if z.cmp(r) != 0 {
			t.Fatalf(`#%d: %s * %s
wanted: %s
got   : %s
`, i, c.x, c.y, c.r, z)
		}
	}
}

func TestDiv(t *testing.T) {
	for i, c := range [...]struct {
		x, y, q, r string
	}{
		{"12", "1", "12", "0"},
		{"4", "5", "0", "4"},
		{"25", "5", "5", "0"},
		{"12312321434543624087245323432423412341234", "34580123616717148097544398509435", "356051978", "21326969640595703400318828928804"},
	} {
		x := nat(nil).setString(c.x)
		y := nat(nil).setString(c.y)
		rq := nat(nil).setString(c.q)
		rr := nat(nil).setString(c.r)

		q, r := nat(nil).div(nil, x, y)
		if q.cmp(rq) != 0 || r.cmp(rr) != 0 {
			t.Fatalf(`#%d: %s / %s
wanted: (%s, %s)
got   : (%s, %s)
`, i, c.x, c.y, c.q, c.r, q, r)
		}
	}
}

func fibo(n int) nat {
	switch n {
	case 0:
		return nil
	case 1:
		return nat{1}
	}
	f0 := fibo(0)
	f1 := fibo(1)
	var f2 nat
	for i := 1; i < n; i++ {
		f2 = f2.add(f0, f1)
		f0, f1, f2 = f1, f2, f0
	}
	return f1
}

var fiboNums = [...]string{
	"0",
	"55",
	"6765",
	"832040",
	"102334155",
	"12586269025",
	"1548008755920",
	"190392490709135",
	"23416728348467685",
	"2880067194370816120",
	"354224848179261915075",
}

func TestFibo(t *testing.T) {
	for i, want := range fiboNums {
		n := i * 10
		f := fibo(n)
		got := f.String()
		if got != want {
			t.Errorf("fibo(%d) failed: got %s want %s", n, got, want)
		}
	}
}

func BenchmarkFibo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fibo(1e0)
		fibo(1e1)
		fibo(1e2)
		fibo(1e3)
		fibo(1e4)
		fibo(1e5)
	}
}
