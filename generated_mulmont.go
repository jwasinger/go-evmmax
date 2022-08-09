package mont_arith

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"unsafe"
)

// madd0 hi = a*b + c (discards lo bits)
func madd0(a, b, c uint64) uint64 {
	var carry, lo uint64
	hi, lo := bits.Mul64(a, b)
	_, carry = bits.Add64(lo, c, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi
}

// madd1 hi, lo = a*b + c
func madd1(a, b, c uint64) (uint64, uint64) {
	var carry uint64
	hi, lo := bits.Mul64(a, b)
	lo, carry = bits.Add64(lo, c, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi, lo
}

// madd2 hi, lo = a*b + c + d
func madd2(a, b, c, d uint64) (uint64, uint64) {
	var carry uint64
	hi, lo := bits.Mul64(a, b)
	c, carry = bits.Add64(c, d, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	lo, carry = bits.Add64(lo, c, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi, lo
}

func madd3(a, b, c, d, e uint64) (uint64, uint64) {
	var carry uint64
	var c_uint uint64
	hi, lo := bits.Mul64(a, b)
	c_uint, carry = bits.Add64(c, d, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	lo, carry = bits.Add64(lo, c_uint, 0)
	hi, _ = bits.Add64(hi, e, carry)
	return hi, lo
}

/*
 * begin mulmont implementations
 */

func mulMont64(f *Field, outBytes, xBytes, yBytes []byte) error {
	var product [2]uint64
	var c uint64
	mod := f.Modulus
	modinv := f.MontParamInterleaved

	x := (*[1]uint64)(unsafe.Pointer(&xBytes[0]))[:]
	y := (*[1]uint64)(unsafe.Pointer(&yBytes[0]))[:]
	out := (*[1]uint64)(unsafe.Pointer(&outBytes[0]))[:]

	if x[0] >= mod[0] || y[0] >= mod[0] {
		return errors.New(fmt.Sprintf("x/y gte modulus"))
	}

	product[1], product[0] = bits.Mul64(x[0], y[0])
	m := product[0] * modinv
	c, _ = madd1(m, mod[0], product[0])
	out[0] = c + product[1]

	if out[0] > mod[0] {
		out[0] = c - mod[0]
	}
	return nil
}

func mulMont128(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[2]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[2]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[2]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[2]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [3]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)

	t[2], D = bits.Add64(t[2], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	t[1], C = bits.Add64(t[2], C, 0)
	t[2], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)

	t[2], D = bits.Add64(t[2], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	t[1], C = bits.Add64(t[2], C, 0)
	t[2], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)

	if D != 0 && t[2] == 0 {
		// reduction was not necessary
		copy(z[:], t[:2])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont192(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[3]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[3]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[3]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[3]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [4]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)

	t[3], D = bits.Add64(t[3], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	t[2], C = bits.Add64(t[3], C, 0)
	t[3], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)

	t[3], D = bits.Add64(t[3], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	t[2], C = bits.Add64(t[3], C, 0)
	t[3], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)

	t[3], D = bits.Add64(t[3], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	t[2], C = bits.Add64(t[3], C, 0)
	t[3], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)

	if D != 0 && t[3] == 0 {
		// reduction was not necessary
		copy(z[:], t[:3])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont256(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[4]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[4]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[4]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[4]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [5]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)

	t[4], D = bits.Add64(t[4], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)

	if D != 0 && t[4] == 0 {
		// reduction was not necessary
		copy(z[:], t[:4])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont320(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[5]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[5]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[5]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[5]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [6]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)

	t[5], D = bits.Add64(t[5], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	t[4], C = bits.Add64(t[5], C, 0)
	t[5], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)

	t[5], D = bits.Add64(t[5], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	t[4], C = bits.Add64(t[5], C, 0)
	t[5], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)

	t[5], D = bits.Add64(t[5], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	t[4], C = bits.Add64(t[5], C, 0)
	t[5], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)

	t[5], D = bits.Add64(t[5], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	t[4], C = bits.Add64(t[5], C, 0)
	t[5], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)

	t[5], D = bits.Add64(t[5], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	t[4], C = bits.Add64(t[5], C, 0)
	t[5], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)

	if D != 0 && t[5] == 0 {
		// reduction was not necessary
		copy(z[:], t[:5])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont384(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[6]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[6]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[6]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[6]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [7]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)

	t[6], D = bits.Add64(t[6], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	t[5], C = bits.Add64(t[6], C, 0)
	t[6], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)

	t[6], D = bits.Add64(t[6], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	t[5], C = bits.Add64(t[6], C, 0)
	t[6], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)

	t[6], D = bits.Add64(t[6], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	t[5], C = bits.Add64(t[6], C, 0)
	t[6], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)

	t[6], D = bits.Add64(t[6], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	t[5], C = bits.Add64(t[6], C, 0)
	t[6], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)

	t[6], D = bits.Add64(t[6], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	t[5], C = bits.Add64(t[6], C, 0)
	t[6], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)

	t[6], D = bits.Add64(t[6], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	t[5], C = bits.Add64(t[6], C, 0)
	t[6], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)

	if D != 0 && t[6] == 0 {
		// reduction was not necessary
		copy(z[:], t[:6])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont448(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[7]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[7]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[7]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[7]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [8]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)

	t[7], D = bits.Add64(t[7], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	t[6], C = bits.Add64(t[7], C, 0)
	t[7], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)

	if D != 0 && t[7] == 0 {
		// reduction was not necessary
		copy(z[:], t[:7])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont512(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[8]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[8]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[8]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[8]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [9]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)

	t[8], D = bits.Add64(t[8], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	t[7], C = bits.Add64(t[8], C, 0)
	t[8], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)

	if D != 0 && t[8] == 0 {
		// reduction was not necessary
		copy(z[:], t[:8])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont576(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[9]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[9]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[9]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[9]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [10]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)

	t[9], D = bits.Add64(t[9], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	t[8], C = bits.Add64(t[9], C, 0)
	t[9], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)

	if D != 0 && t[9] == 0 {
		// reduction was not necessary
		copy(z[:], t[:9])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont640(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[10]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[10]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[10]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[10]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [11]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)

	t[10], D = bits.Add64(t[10], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	t[9], C = bits.Add64(t[10], C, 0)
	t[10], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)

	if D != 0 && t[10] == 0 {
		// reduction was not necessary
		copy(z[:], t[:10])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont704(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[11]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[11]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[11]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[11]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [12]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)

	t[11], D = bits.Add64(t[11], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	t[10], C = bits.Add64(t[11], C, 0)
	t[11], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)

	if D != 0 && t[11] == 0 {
		// reduction was not necessary
		copy(z[:], t[:11])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont768(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[12]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[12]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[12]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[12]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [13]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)

	t[12], D = bits.Add64(t[12], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	t[11], C = bits.Add64(t[12], C, 0)
	t[12], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)

	if D != 0 && t[12] == 0 {
		// reduction was not necessary
		copy(z[:], t[:12])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont832(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[13]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[13]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[13]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[13]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [14]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)

	t[13], D = bits.Add64(t[13], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	t[12], C = bits.Add64(t[13], C, 0)
	t[13], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)

	if D != 0 && t[13] == 0 {
		// reduction was not necessary
		copy(z[:], t[:13])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont896(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[14]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[14]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[14]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[14]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [15]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)

	t[14], D = bits.Add64(t[14], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	t[13], C = bits.Add64(t[14], C, 0)
	t[14], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)

	if D != 0 && t[14] == 0 {
		// reduction was not necessary
		copy(z[:], t[:14])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont960(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[15]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[15]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[15]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[15]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [16]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)

	t[15], D = bits.Add64(t[15], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	t[14], C = bits.Add64(t[15], C, 0)
	t[15], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)

	if D != 0 && t[15] == 0 {
		// reduction was not necessary
		copy(z[:], t[:15])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1024(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[16]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[16]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[16]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[16]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [17]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)

	t[16], D = bits.Add64(t[16], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	t[15], C = bits.Add64(t[16], C, 0)
	t[16], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)

	if D != 0 && t[16] == 0 {
		// reduction was not necessary
		copy(z[:], t[:16])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1088(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[17]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[17]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[17]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[17]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [18]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)

	t[17], D = bits.Add64(t[17], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	t[16], C = bits.Add64(t[17], C, 0)
	t[17], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)

	if D != 0 && t[17] == 0 {
		// reduction was not necessary
		copy(z[:], t[:17])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1152(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[18]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[18]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[18]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[18]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [19]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)

	t[18], D = bits.Add64(t[18], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	t[17], C = bits.Add64(t[18], C, 0)
	t[18], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)

	if D != 0 && t[18] == 0 {
		// reduction was not necessary
		copy(z[:], t[:18])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1216(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[19]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[19]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[19]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[19]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [20]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC1 = bits.Sub64(mod[18], x[18], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)
	_, gteC2 = bits.Sub64(mod[18], x[18], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)
	C, t[18] = madd1(x[0], y[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)
	C, t[18] = madd2(x[1], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)
	C, t[18] = madd2(x[2], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)
	C, t[18] = madd2(x[3], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)
	C, t[18] = madd2(x[4], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)
	C, t[18] = madd2(x[5], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)
	C, t[18] = madd2(x[6], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)
	C, t[18] = madd2(x[7], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)
	C, t[18] = madd2(x[8], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)
	C, t[18] = madd2(x[9], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)
	C, t[18] = madd2(x[10], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)
	C, t[18] = madd2(x[11], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)
	C, t[18] = madd2(x[12], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)
	C, t[18] = madd2(x[13], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)
	C, t[18] = madd2(x[14], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)
	C, t[18] = madd2(x[15], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)
	C, t[18] = madd2(x[16], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)
	C, t[18] = madd2(x[17], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[18], y[0], t[0])
	C, t[1] = madd2(x[18], y[1], t[1], C)
	C, t[2] = madd2(x[18], y[2], t[2], C)
	C, t[3] = madd2(x[18], y[3], t[3], C)
	C, t[4] = madd2(x[18], y[4], t[4], C)
	C, t[5] = madd2(x[18], y[5], t[5], C)
	C, t[6] = madd2(x[18], y[6], t[6], C)
	C, t[7] = madd2(x[18], y[7], t[7], C)
	C, t[8] = madd2(x[18], y[8], t[8], C)
	C, t[9] = madd2(x[18], y[9], t[9], C)
	C, t[10] = madd2(x[18], y[10], t[10], C)
	C, t[11] = madd2(x[18], y[11], t[11], C)
	C, t[12] = madd2(x[18], y[12], t[12], C)
	C, t[13] = madd2(x[18], y[13], t[13], C)
	C, t[14] = madd2(x[18], y[14], t[14], C)
	C, t[15] = madd2(x[18], y[15], t[15], C)
	C, t[16] = madd2(x[18], y[16], t[16], C)
	C, t[17] = madd2(x[18], y[17], t[17], C)
	C, t[18] = madd2(x[18], y[18], t[18], C)

	t[19], D = bits.Add64(t[19], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	t[18], C = bits.Add64(t[19], C, 0)
	t[19], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)
	z[18], D = bits.Sub64(t[18], mod[18], D)

	if D != 0 && t[19] == 0 {
		// reduction was not necessary
		copy(z[:], t[:19])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1280(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[20]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[20]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[20]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[20]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [21]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC1 = bits.Sub64(mod[18], x[18], gteC1)
	_, gteC1 = bits.Sub64(mod[19], x[19], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)
	_, gteC2 = bits.Sub64(mod[18], x[18], gteC2)
	_, gteC2 = bits.Sub64(mod[19], x[19], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)
	C, t[18] = madd1(x[0], y[18], C)
	C, t[19] = madd1(x[0], y[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)
	C, t[18] = madd2(x[1], y[18], t[18], C)
	C, t[19] = madd2(x[1], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)
	C, t[18] = madd2(x[2], y[18], t[18], C)
	C, t[19] = madd2(x[2], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)
	C, t[18] = madd2(x[3], y[18], t[18], C)
	C, t[19] = madd2(x[3], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)
	C, t[18] = madd2(x[4], y[18], t[18], C)
	C, t[19] = madd2(x[4], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)
	C, t[18] = madd2(x[5], y[18], t[18], C)
	C, t[19] = madd2(x[5], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)
	C, t[18] = madd2(x[6], y[18], t[18], C)
	C, t[19] = madd2(x[6], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)
	C, t[18] = madd2(x[7], y[18], t[18], C)
	C, t[19] = madd2(x[7], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)
	C, t[18] = madd2(x[8], y[18], t[18], C)
	C, t[19] = madd2(x[8], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)
	C, t[18] = madd2(x[9], y[18], t[18], C)
	C, t[19] = madd2(x[9], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)
	C, t[18] = madd2(x[10], y[18], t[18], C)
	C, t[19] = madd2(x[10], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)
	C, t[18] = madd2(x[11], y[18], t[18], C)
	C, t[19] = madd2(x[11], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)
	C, t[18] = madd2(x[12], y[18], t[18], C)
	C, t[19] = madd2(x[12], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)
	C, t[18] = madd2(x[13], y[18], t[18], C)
	C, t[19] = madd2(x[13], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)
	C, t[18] = madd2(x[14], y[18], t[18], C)
	C, t[19] = madd2(x[14], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)
	C, t[18] = madd2(x[15], y[18], t[18], C)
	C, t[19] = madd2(x[15], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)
	C, t[18] = madd2(x[16], y[18], t[18], C)
	C, t[19] = madd2(x[16], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)
	C, t[18] = madd2(x[17], y[18], t[18], C)
	C, t[19] = madd2(x[17], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[18], y[0], t[0])
	C, t[1] = madd2(x[18], y[1], t[1], C)
	C, t[2] = madd2(x[18], y[2], t[2], C)
	C, t[3] = madd2(x[18], y[3], t[3], C)
	C, t[4] = madd2(x[18], y[4], t[4], C)
	C, t[5] = madd2(x[18], y[5], t[5], C)
	C, t[6] = madd2(x[18], y[6], t[6], C)
	C, t[7] = madd2(x[18], y[7], t[7], C)
	C, t[8] = madd2(x[18], y[8], t[8], C)
	C, t[9] = madd2(x[18], y[9], t[9], C)
	C, t[10] = madd2(x[18], y[10], t[10], C)
	C, t[11] = madd2(x[18], y[11], t[11], C)
	C, t[12] = madd2(x[18], y[12], t[12], C)
	C, t[13] = madd2(x[18], y[13], t[13], C)
	C, t[14] = madd2(x[18], y[14], t[14], C)
	C, t[15] = madd2(x[18], y[15], t[15], C)
	C, t[16] = madd2(x[18], y[16], t[16], C)
	C, t[17] = madd2(x[18], y[17], t[17], C)
	C, t[18] = madd2(x[18], y[18], t[18], C)
	C, t[19] = madd2(x[18], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[19], y[0], t[0])
	C, t[1] = madd2(x[19], y[1], t[1], C)
	C, t[2] = madd2(x[19], y[2], t[2], C)
	C, t[3] = madd2(x[19], y[3], t[3], C)
	C, t[4] = madd2(x[19], y[4], t[4], C)
	C, t[5] = madd2(x[19], y[5], t[5], C)
	C, t[6] = madd2(x[19], y[6], t[6], C)
	C, t[7] = madd2(x[19], y[7], t[7], C)
	C, t[8] = madd2(x[19], y[8], t[8], C)
	C, t[9] = madd2(x[19], y[9], t[9], C)
	C, t[10] = madd2(x[19], y[10], t[10], C)
	C, t[11] = madd2(x[19], y[11], t[11], C)
	C, t[12] = madd2(x[19], y[12], t[12], C)
	C, t[13] = madd2(x[19], y[13], t[13], C)
	C, t[14] = madd2(x[19], y[14], t[14], C)
	C, t[15] = madd2(x[19], y[15], t[15], C)
	C, t[16] = madd2(x[19], y[16], t[16], C)
	C, t[17] = madd2(x[19], y[17], t[17], C)
	C, t[18] = madd2(x[19], y[18], t[18], C)
	C, t[19] = madd2(x[19], y[19], t[19], C)

	t[20], D = bits.Add64(t[20], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	t[19], C = bits.Add64(t[20], C, 0)
	t[20], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)
	z[18], D = bits.Sub64(t[18], mod[18], D)
	z[19], D = bits.Sub64(t[19], mod[19], D)

	if D != 0 && t[20] == 0 {
		// reduction was not necessary
		copy(z[:], t[:20])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1344(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[21]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[21]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[21]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[21]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [22]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC1 = bits.Sub64(mod[18], x[18], gteC1)
	_, gteC1 = bits.Sub64(mod[19], x[19], gteC1)
	_, gteC1 = bits.Sub64(mod[20], x[20], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)
	_, gteC2 = bits.Sub64(mod[18], x[18], gteC2)
	_, gteC2 = bits.Sub64(mod[19], x[19], gteC2)
	_, gteC2 = bits.Sub64(mod[20], x[20], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)
	C, t[18] = madd1(x[0], y[18], C)
	C, t[19] = madd1(x[0], y[19], C)
	C, t[20] = madd1(x[0], y[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)
	C, t[18] = madd2(x[1], y[18], t[18], C)
	C, t[19] = madd2(x[1], y[19], t[19], C)
	C, t[20] = madd2(x[1], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)
	C, t[18] = madd2(x[2], y[18], t[18], C)
	C, t[19] = madd2(x[2], y[19], t[19], C)
	C, t[20] = madd2(x[2], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)
	C, t[18] = madd2(x[3], y[18], t[18], C)
	C, t[19] = madd2(x[3], y[19], t[19], C)
	C, t[20] = madd2(x[3], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)
	C, t[18] = madd2(x[4], y[18], t[18], C)
	C, t[19] = madd2(x[4], y[19], t[19], C)
	C, t[20] = madd2(x[4], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)
	C, t[18] = madd2(x[5], y[18], t[18], C)
	C, t[19] = madd2(x[5], y[19], t[19], C)
	C, t[20] = madd2(x[5], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)
	C, t[18] = madd2(x[6], y[18], t[18], C)
	C, t[19] = madd2(x[6], y[19], t[19], C)
	C, t[20] = madd2(x[6], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)
	C, t[18] = madd2(x[7], y[18], t[18], C)
	C, t[19] = madd2(x[7], y[19], t[19], C)
	C, t[20] = madd2(x[7], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)
	C, t[18] = madd2(x[8], y[18], t[18], C)
	C, t[19] = madd2(x[8], y[19], t[19], C)
	C, t[20] = madd2(x[8], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)
	C, t[18] = madd2(x[9], y[18], t[18], C)
	C, t[19] = madd2(x[9], y[19], t[19], C)
	C, t[20] = madd2(x[9], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)
	C, t[18] = madd2(x[10], y[18], t[18], C)
	C, t[19] = madd2(x[10], y[19], t[19], C)
	C, t[20] = madd2(x[10], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)
	C, t[18] = madd2(x[11], y[18], t[18], C)
	C, t[19] = madd2(x[11], y[19], t[19], C)
	C, t[20] = madd2(x[11], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)
	C, t[18] = madd2(x[12], y[18], t[18], C)
	C, t[19] = madd2(x[12], y[19], t[19], C)
	C, t[20] = madd2(x[12], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)
	C, t[18] = madd2(x[13], y[18], t[18], C)
	C, t[19] = madd2(x[13], y[19], t[19], C)
	C, t[20] = madd2(x[13], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)
	C, t[18] = madd2(x[14], y[18], t[18], C)
	C, t[19] = madd2(x[14], y[19], t[19], C)
	C, t[20] = madd2(x[14], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)
	C, t[18] = madd2(x[15], y[18], t[18], C)
	C, t[19] = madd2(x[15], y[19], t[19], C)
	C, t[20] = madd2(x[15], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)
	C, t[18] = madd2(x[16], y[18], t[18], C)
	C, t[19] = madd2(x[16], y[19], t[19], C)
	C, t[20] = madd2(x[16], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)
	C, t[18] = madd2(x[17], y[18], t[18], C)
	C, t[19] = madd2(x[17], y[19], t[19], C)
	C, t[20] = madd2(x[17], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[18], y[0], t[0])
	C, t[1] = madd2(x[18], y[1], t[1], C)
	C, t[2] = madd2(x[18], y[2], t[2], C)
	C, t[3] = madd2(x[18], y[3], t[3], C)
	C, t[4] = madd2(x[18], y[4], t[4], C)
	C, t[5] = madd2(x[18], y[5], t[5], C)
	C, t[6] = madd2(x[18], y[6], t[6], C)
	C, t[7] = madd2(x[18], y[7], t[7], C)
	C, t[8] = madd2(x[18], y[8], t[8], C)
	C, t[9] = madd2(x[18], y[9], t[9], C)
	C, t[10] = madd2(x[18], y[10], t[10], C)
	C, t[11] = madd2(x[18], y[11], t[11], C)
	C, t[12] = madd2(x[18], y[12], t[12], C)
	C, t[13] = madd2(x[18], y[13], t[13], C)
	C, t[14] = madd2(x[18], y[14], t[14], C)
	C, t[15] = madd2(x[18], y[15], t[15], C)
	C, t[16] = madd2(x[18], y[16], t[16], C)
	C, t[17] = madd2(x[18], y[17], t[17], C)
	C, t[18] = madd2(x[18], y[18], t[18], C)
	C, t[19] = madd2(x[18], y[19], t[19], C)
	C, t[20] = madd2(x[18], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[19], y[0], t[0])
	C, t[1] = madd2(x[19], y[1], t[1], C)
	C, t[2] = madd2(x[19], y[2], t[2], C)
	C, t[3] = madd2(x[19], y[3], t[3], C)
	C, t[4] = madd2(x[19], y[4], t[4], C)
	C, t[5] = madd2(x[19], y[5], t[5], C)
	C, t[6] = madd2(x[19], y[6], t[6], C)
	C, t[7] = madd2(x[19], y[7], t[7], C)
	C, t[8] = madd2(x[19], y[8], t[8], C)
	C, t[9] = madd2(x[19], y[9], t[9], C)
	C, t[10] = madd2(x[19], y[10], t[10], C)
	C, t[11] = madd2(x[19], y[11], t[11], C)
	C, t[12] = madd2(x[19], y[12], t[12], C)
	C, t[13] = madd2(x[19], y[13], t[13], C)
	C, t[14] = madd2(x[19], y[14], t[14], C)
	C, t[15] = madd2(x[19], y[15], t[15], C)
	C, t[16] = madd2(x[19], y[16], t[16], C)
	C, t[17] = madd2(x[19], y[17], t[17], C)
	C, t[18] = madd2(x[19], y[18], t[18], C)
	C, t[19] = madd2(x[19], y[19], t[19], C)
	C, t[20] = madd2(x[19], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[20], y[0], t[0])
	C, t[1] = madd2(x[20], y[1], t[1], C)
	C, t[2] = madd2(x[20], y[2], t[2], C)
	C, t[3] = madd2(x[20], y[3], t[3], C)
	C, t[4] = madd2(x[20], y[4], t[4], C)
	C, t[5] = madd2(x[20], y[5], t[5], C)
	C, t[6] = madd2(x[20], y[6], t[6], C)
	C, t[7] = madd2(x[20], y[7], t[7], C)
	C, t[8] = madd2(x[20], y[8], t[8], C)
	C, t[9] = madd2(x[20], y[9], t[9], C)
	C, t[10] = madd2(x[20], y[10], t[10], C)
	C, t[11] = madd2(x[20], y[11], t[11], C)
	C, t[12] = madd2(x[20], y[12], t[12], C)
	C, t[13] = madd2(x[20], y[13], t[13], C)
	C, t[14] = madd2(x[20], y[14], t[14], C)
	C, t[15] = madd2(x[20], y[15], t[15], C)
	C, t[16] = madd2(x[20], y[16], t[16], C)
	C, t[17] = madd2(x[20], y[17], t[17], C)
	C, t[18] = madd2(x[20], y[18], t[18], C)
	C, t[19] = madd2(x[20], y[19], t[19], C)
	C, t[20] = madd2(x[20], y[20], t[20], C)

	t[21], D = bits.Add64(t[21], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	t[20], C = bits.Add64(t[21], C, 0)
	t[21], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)
	z[18], D = bits.Sub64(t[18], mod[18], D)
	z[19], D = bits.Sub64(t[19], mod[19], D)
	z[20], D = bits.Sub64(t[20], mod[20], D)

	if D != 0 && t[21] == 0 {
		// reduction was not necessary
		copy(z[:], t[:21])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1408(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[22]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[22]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[22]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[22]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [23]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC1 = bits.Sub64(mod[18], x[18], gteC1)
	_, gteC1 = bits.Sub64(mod[19], x[19], gteC1)
	_, gteC1 = bits.Sub64(mod[20], x[20], gteC1)
	_, gteC1 = bits.Sub64(mod[21], x[21], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)
	_, gteC2 = bits.Sub64(mod[18], x[18], gteC2)
	_, gteC2 = bits.Sub64(mod[19], x[19], gteC2)
	_, gteC2 = bits.Sub64(mod[20], x[20], gteC2)
	_, gteC2 = bits.Sub64(mod[21], x[21], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)
	C, t[18] = madd1(x[0], y[18], C)
	C, t[19] = madd1(x[0], y[19], C)
	C, t[20] = madd1(x[0], y[20], C)
	C, t[21] = madd1(x[0], y[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)
	C, t[18] = madd2(x[1], y[18], t[18], C)
	C, t[19] = madd2(x[1], y[19], t[19], C)
	C, t[20] = madd2(x[1], y[20], t[20], C)
	C, t[21] = madd2(x[1], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)
	C, t[18] = madd2(x[2], y[18], t[18], C)
	C, t[19] = madd2(x[2], y[19], t[19], C)
	C, t[20] = madd2(x[2], y[20], t[20], C)
	C, t[21] = madd2(x[2], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)
	C, t[18] = madd2(x[3], y[18], t[18], C)
	C, t[19] = madd2(x[3], y[19], t[19], C)
	C, t[20] = madd2(x[3], y[20], t[20], C)
	C, t[21] = madd2(x[3], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)
	C, t[18] = madd2(x[4], y[18], t[18], C)
	C, t[19] = madd2(x[4], y[19], t[19], C)
	C, t[20] = madd2(x[4], y[20], t[20], C)
	C, t[21] = madd2(x[4], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)
	C, t[18] = madd2(x[5], y[18], t[18], C)
	C, t[19] = madd2(x[5], y[19], t[19], C)
	C, t[20] = madd2(x[5], y[20], t[20], C)
	C, t[21] = madd2(x[5], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)
	C, t[18] = madd2(x[6], y[18], t[18], C)
	C, t[19] = madd2(x[6], y[19], t[19], C)
	C, t[20] = madd2(x[6], y[20], t[20], C)
	C, t[21] = madd2(x[6], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)
	C, t[18] = madd2(x[7], y[18], t[18], C)
	C, t[19] = madd2(x[7], y[19], t[19], C)
	C, t[20] = madd2(x[7], y[20], t[20], C)
	C, t[21] = madd2(x[7], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)
	C, t[18] = madd2(x[8], y[18], t[18], C)
	C, t[19] = madd2(x[8], y[19], t[19], C)
	C, t[20] = madd2(x[8], y[20], t[20], C)
	C, t[21] = madd2(x[8], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)
	C, t[18] = madd2(x[9], y[18], t[18], C)
	C, t[19] = madd2(x[9], y[19], t[19], C)
	C, t[20] = madd2(x[9], y[20], t[20], C)
	C, t[21] = madd2(x[9], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)
	C, t[18] = madd2(x[10], y[18], t[18], C)
	C, t[19] = madd2(x[10], y[19], t[19], C)
	C, t[20] = madd2(x[10], y[20], t[20], C)
	C, t[21] = madd2(x[10], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)
	C, t[18] = madd2(x[11], y[18], t[18], C)
	C, t[19] = madd2(x[11], y[19], t[19], C)
	C, t[20] = madd2(x[11], y[20], t[20], C)
	C, t[21] = madd2(x[11], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)
	C, t[18] = madd2(x[12], y[18], t[18], C)
	C, t[19] = madd2(x[12], y[19], t[19], C)
	C, t[20] = madd2(x[12], y[20], t[20], C)
	C, t[21] = madd2(x[12], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)
	C, t[18] = madd2(x[13], y[18], t[18], C)
	C, t[19] = madd2(x[13], y[19], t[19], C)
	C, t[20] = madd2(x[13], y[20], t[20], C)
	C, t[21] = madd2(x[13], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)
	C, t[18] = madd2(x[14], y[18], t[18], C)
	C, t[19] = madd2(x[14], y[19], t[19], C)
	C, t[20] = madd2(x[14], y[20], t[20], C)
	C, t[21] = madd2(x[14], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)
	C, t[18] = madd2(x[15], y[18], t[18], C)
	C, t[19] = madd2(x[15], y[19], t[19], C)
	C, t[20] = madd2(x[15], y[20], t[20], C)
	C, t[21] = madd2(x[15], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)
	C, t[18] = madd2(x[16], y[18], t[18], C)
	C, t[19] = madd2(x[16], y[19], t[19], C)
	C, t[20] = madd2(x[16], y[20], t[20], C)
	C, t[21] = madd2(x[16], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)
	C, t[18] = madd2(x[17], y[18], t[18], C)
	C, t[19] = madd2(x[17], y[19], t[19], C)
	C, t[20] = madd2(x[17], y[20], t[20], C)
	C, t[21] = madd2(x[17], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[18], y[0], t[0])
	C, t[1] = madd2(x[18], y[1], t[1], C)
	C, t[2] = madd2(x[18], y[2], t[2], C)
	C, t[3] = madd2(x[18], y[3], t[3], C)
	C, t[4] = madd2(x[18], y[4], t[4], C)
	C, t[5] = madd2(x[18], y[5], t[5], C)
	C, t[6] = madd2(x[18], y[6], t[6], C)
	C, t[7] = madd2(x[18], y[7], t[7], C)
	C, t[8] = madd2(x[18], y[8], t[8], C)
	C, t[9] = madd2(x[18], y[9], t[9], C)
	C, t[10] = madd2(x[18], y[10], t[10], C)
	C, t[11] = madd2(x[18], y[11], t[11], C)
	C, t[12] = madd2(x[18], y[12], t[12], C)
	C, t[13] = madd2(x[18], y[13], t[13], C)
	C, t[14] = madd2(x[18], y[14], t[14], C)
	C, t[15] = madd2(x[18], y[15], t[15], C)
	C, t[16] = madd2(x[18], y[16], t[16], C)
	C, t[17] = madd2(x[18], y[17], t[17], C)
	C, t[18] = madd2(x[18], y[18], t[18], C)
	C, t[19] = madd2(x[18], y[19], t[19], C)
	C, t[20] = madd2(x[18], y[20], t[20], C)
	C, t[21] = madd2(x[18], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[19], y[0], t[0])
	C, t[1] = madd2(x[19], y[1], t[1], C)
	C, t[2] = madd2(x[19], y[2], t[2], C)
	C, t[3] = madd2(x[19], y[3], t[3], C)
	C, t[4] = madd2(x[19], y[4], t[4], C)
	C, t[5] = madd2(x[19], y[5], t[5], C)
	C, t[6] = madd2(x[19], y[6], t[6], C)
	C, t[7] = madd2(x[19], y[7], t[7], C)
	C, t[8] = madd2(x[19], y[8], t[8], C)
	C, t[9] = madd2(x[19], y[9], t[9], C)
	C, t[10] = madd2(x[19], y[10], t[10], C)
	C, t[11] = madd2(x[19], y[11], t[11], C)
	C, t[12] = madd2(x[19], y[12], t[12], C)
	C, t[13] = madd2(x[19], y[13], t[13], C)
	C, t[14] = madd2(x[19], y[14], t[14], C)
	C, t[15] = madd2(x[19], y[15], t[15], C)
	C, t[16] = madd2(x[19], y[16], t[16], C)
	C, t[17] = madd2(x[19], y[17], t[17], C)
	C, t[18] = madd2(x[19], y[18], t[18], C)
	C, t[19] = madd2(x[19], y[19], t[19], C)
	C, t[20] = madd2(x[19], y[20], t[20], C)
	C, t[21] = madd2(x[19], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[20], y[0], t[0])
	C, t[1] = madd2(x[20], y[1], t[1], C)
	C, t[2] = madd2(x[20], y[2], t[2], C)
	C, t[3] = madd2(x[20], y[3], t[3], C)
	C, t[4] = madd2(x[20], y[4], t[4], C)
	C, t[5] = madd2(x[20], y[5], t[5], C)
	C, t[6] = madd2(x[20], y[6], t[6], C)
	C, t[7] = madd2(x[20], y[7], t[7], C)
	C, t[8] = madd2(x[20], y[8], t[8], C)
	C, t[9] = madd2(x[20], y[9], t[9], C)
	C, t[10] = madd2(x[20], y[10], t[10], C)
	C, t[11] = madd2(x[20], y[11], t[11], C)
	C, t[12] = madd2(x[20], y[12], t[12], C)
	C, t[13] = madd2(x[20], y[13], t[13], C)
	C, t[14] = madd2(x[20], y[14], t[14], C)
	C, t[15] = madd2(x[20], y[15], t[15], C)
	C, t[16] = madd2(x[20], y[16], t[16], C)
	C, t[17] = madd2(x[20], y[17], t[17], C)
	C, t[18] = madd2(x[20], y[18], t[18], C)
	C, t[19] = madd2(x[20], y[19], t[19], C)
	C, t[20] = madd2(x[20], y[20], t[20], C)
	C, t[21] = madd2(x[20], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[21], y[0], t[0])
	C, t[1] = madd2(x[21], y[1], t[1], C)
	C, t[2] = madd2(x[21], y[2], t[2], C)
	C, t[3] = madd2(x[21], y[3], t[3], C)
	C, t[4] = madd2(x[21], y[4], t[4], C)
	C, t[5] = madd2(x[21], y[5], t[5], C)
	C, t[6] = madd2(x[21], y[6], t[6], C)
	C, t[7] = madd2(x[21], y[7], t[7], C)
	C, t[8] = madd2(x[21], y[8], t[8], C)
	C, t[9] = madd2(x[21], y[9], t[9], C)
	C, t[10] = madd2(x[21], y[10], t[10], C)
	C, t[11] = madd2(x[21], y[11], t[11], C)
	C, t[12] = madd2(x[21], y[12], t[12], C)
	C, t[13] = madd2(x[21], y[13], t[13], C)
	C, t[14] = madd2(x[21], y[14], t[14], C)
	C, t[15] = madd2(x[21], y[15], t[15], C)
	C, t[16] = madd2(x[21], y[16], t[16], C)
	C, t[17] = madd2(x[21], y[17], t[17], C)
	C, t[18] = madd2(x[21], y[18], t[18], C)
	C, t[19] = madd2(x[21], y[19], t[19], C)
	C, t[20] = madd2(x[21], y[20], t[20], C)
	C, t[21] = madd2(x[21], y[21], t[21], C)

	t[22], D = bits.Add64(t[22], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	t[21], C = bits.Add64(t[22], C, 0)
	t[22], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)
	z[18], D = bits.Sub64(t[18], mod[18], D)
	z[19], D = bits.Sub64(t[19], mod[19], D)
	z[20], D = bits.Sub64(t[20], mod[20], D)
	z[21], D = bits.Sub64(t[21], mod[21], D)

	if D != 0 && t[22] == 0 {
		// reduction was not necessary
		copy(z[:], t[:22])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1472(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[23]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[23]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[23]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[23]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [24]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC1 = bits.Sub64(mod[18], x[18], gteC1)
	_, gteC1 = bits.Sub64(mod[19], x[19], gteC1)
	_, gteC1 = bits.Sub64(mod[20], x[20], gteC1)
	_, gteC1 = bits.Sub64(mod[21], x[21], gteC1)
	_, gteC1 = bits.Sub64(mod[22], x[22], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)
	_, gteC2 = bits.Sub64(mod[18], x[18], gteC2)
	_, gteC2 = bits.Sub64(mod[19], x[19], gteC2)
	_, gteC2 = bits.Sub64(mod[20], x[20], gteC2)
	_, gteC2 = bits.Sub64(mod[21], x[21], gteC2)
	_, gteC2 = bits.Sub64(mod[22], x[22], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)
	C, t[18] = madd1(x[0], y[18], C)
	C, t[19] = madd1(x[0], y[19], C)
	C, t[20] = madd1(x[0], y[20], C)
	C, t[21] = madd1(x[0], y[21], C)
	C, t[22] = madd1(x[0], y[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)
	C, t[18] = madd2(x[1], y[18], t[18], C)
	C, t[19] = madd2(x[1], y[19], t[19], C)
	C, t[20] = madd2(x[1], y[20], t[20], C)
	C, t[21] = madd2(x[1], y[21], t[21], C)
	C, t[22] = madd2(x[1], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)
	C, t[18] = madd2(x[2], y[18], t[18], C)
	C, t[19] = madd2(x[2], y[19], t[19], C)
	C, t[20] = madd2(x[2], y[20], t[20], C)
	C, t[21] = madd2(x[2], y[21], t[21], C)
	C, t[22] = madd2(x[2], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)
	C, t[18] = madd2(x[3], y[18], t[18], C)
	C, t[19] = madd2(x[3], y[19], t[19], C)
	C, t[20] = madd2(x[3], y[20], t[20], C)
	C, t[21] = madd2(x[3], y[21], t[21], C)
	C, t[22] = madd2(x[3], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)
	C, t[18] = madd2(x[4], y[18], t[18], C)
	C, t[19] = madd2(x[4], y[19], t[19], C)
	C, t[20] = madd2(x[4], y[20], t[20], C)
	C, t[21] = madd2(x[4], y[21], t[21], C)
	C, t[22] = madd2(x[4], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)
	C, t[18] = madd2(x[5], y[18], t[18], C)
	C, t[19] = madd2(x[5], y[19], t[19], C)
	C, t[20] = madd2(x[5], y[20], t[20], C)
	C, t[21] = madd2(x[5], y[21], t[21], C)
	C, t[22] = madd2(x[5], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)
	C, t[18] = madd2(x[6], y[18], t[18], C)
	C, t[19] = madd2(x[6], y[19], t[19], C)
	C, t[20] = madd2(x[6], y[20], t[20], C)
	C, t[21] = madd2(x[6], y[21], t[21], C)
	C, t[22] = madd2(x[6], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)
	C, t[18] = madd2(x[7], y[18], t[18], C)
	C, t[19] = madd2(x[7], y[19], t[19], C)
	C, t[20] = madd2(x[7], y[20], t[20], C)
	C, t[21] = madd2(x[7], y[21], t[21], C)
	C, t[22] = madd2(x[7], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)
	C, t[18] = madd2(x[8], y[18], t[18], C)
	C, t[19] = madd2(x[8], y[19], t[19], C)
	C, t[20] = madd2(x[8], y[20], t[20], C)
	C, t[21] = madd2(x[8], y[21], t[21], C)
	C, t[22] = madd2(x[8], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)
	C, t[18] = madd2(x[9], y[18], t[18], C)
	C, t[19] = madd2(x[9], y[19], t[19], C)
	C, t[20] = madd2(x[9], y[20], t[20], C)
	C, t[21] = madd2(x[9], y[21], t[21], C)
	C, t[22] = madd2(x[9], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)
	C, t[18] = madd2(x[10], y[18], t[18], C)
	C, t[19] = madd2(x[10], y[19], t[19], C)
	C, t[20] = madd2(x[10], y[20], t[20], C)
	C, t[21] = madd2(x[10], y[21], t[21], C)
	C, t[22] = madd2(x[10], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)
	C, t[18] = madd2(x[11], y[18], t[18], C)
	C, t[19] = madd2(x[11], y[19], t[19], C)
	C, t[20] = madd2(x[11], y[20], t[20], C)
	C, t[21] = madd2(x[11], y[21], t[21], C)
	C, t[22] = madd2(x[11], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)
	C, t[18] = madd2(x[12], y[18], t[18], C)
	C, t[19] = madd2(x[12], y[19], t[19], C)
	C, t[20] = madd2(x[12], y[20], t[20], C)
	C, t[21] = madd2(x[12], y[21], t[21], C)
	C, t[22] = madd2(x[12], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)
	C, t[18] = madd2(x[13], y[18], t[18], C)
	C, t[19] = madd2(x[13], y[19], t[19], C)
	C, t[20] = madd2(x[13], y[20], t[20], C)
	C, t[21] = madd2(x[13], y[21], t[21], C)
	C, t[22] = madd2(x[13], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)
	C, t[18] = madd2(x[14], y[18], t[18], C)
	C, t[19] = madd2(x[14], y[19], t[19], C)
	C, t[20] = madd2(x[14], y[20], t[20], C)
	C, t[21] = madd2(x[14], y[21], t[21], C)
	C, t[22] = madd2(x[14], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)
	C, t[18] = madd2(x[15], y[18], t[18], C)
	C, t[19] = madd2(x[15], y[19], t[19], C)
	C, t[20] = madd2(x[15], y[20], t[20], C)
	C, t[21] = madd2(x[15], y[21], t[21], C)
	C, t[22] = madd2(x[15], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)
	C, t[18] = madd2(x[16], y[18], t[18], C)
	C, t[19] = madd2(x[16], y[19], t[19], C)
	C, t[20] = madd2(x[16], y[20], t[20], C)
	C, t[21] = madd2(x[16], y[21], t[21], C)
	C, t[22] = madd2(x[16], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)
	C, t[18] = madd2(x[17], y[18], t[18], C)
	C, t[19] = madd2(x[17], y[19], t[19], C)
	C, t[20] = madd2(x[17], y[20], t[20], C)
	C, t[21] = madd2(x[17], y[21], t[21], C)
	C, t[22] = madd2(x[17], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[18], y[0], t[0])
	C, t[1] = madd2(x[18], y[1], t[1], C)
	C, t[2] = madd2(x[18], y[2], t[2], C)
	C, t[3] = madd2(x[18], y[3], t[3], C)
	C, t[4] = madd2(x[18], y[4], t[4], C)
	C, t[5] = madd2(x[18], y[5], t[5], C)
	C, t[6] = madd2(x[18], y[6], t[6], C)
	C, t[7] = madd2(x[18], y[7], t[7], C)
	C, t[8] = madd2(x[18], y[8], t[8], C)
	C, t[9] = madd2(x[18], y[9], t[9], C)
	C, t[10] = madd2(x[18], y[10], t[10], C)
	C, t[11] = madd2(x[18], y[11], t[11], C)
	C, t[12] = madd2(x[18], y[12], t[12], C)
	C, t[13] = madd2(x[18], y[13], t[13], C)
	C, t[14] = madd2(x[18], y[14], t[14], C)
	C, t[15] = madd2(x[18], y[15], t[15], C)
	C, t[16] = madd2(x[18], y[16], t[16], C)
	C, t[17] = madd2(x[18], y[17], t[17], C)
	C, t[18] = madd2(x[18], y[18], t[18], C)
	C, t[19] = madd2(x[18], y[19], t[19], C)
	C, t[20] = madd2(x[18], y[20], t[20], C)
	C, t[21] = madd2(x[18], y[21], t[21], C)
	C, t[22] = madd2(x[18], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[19], y[0], t[0])
	C, t[1] = madd2(x[19], y[1], t[1], C)
	C, t[2] = madd2(x[19], y[2], t[2], C)
	C, t[3] = madd2(x[19], y[3], t[3], C)
	C, t[4] = madd2(x[19], y[4], t[4], C)
	C, t[5] = madd2(x[19], y[5], t[5], C)
	C, t[6] = madd2(x[19], y[6], t[6], C)
	C, t[7] = madd2(x[19], y[7], t[7], C)
	C, t[8] = madd2(x[19], y[8], t[8], C)
	C, t[9] = madd2(x[19], y[9], t[9], C)
	C, t[10] = madd2(x[19], y[10], t[10], C)
	C, t[11] = madd2(x[19], y[11], t[11], C)
	C, t[12] = madd2(x[19], y[12], t[12], C)
	C, t[13] = madd2(x[19], y[13], t[13], C)
	C, t[14] = madd2(x[19], y[14], t[14], C)
	C, t[15] = madd2(x[19], y[15], t[15], C)
	C, t[16] = madd2(x[19], y[16], t[16], C)
	C, t[17] = madd2(x[19], y[17], t[17], C)
	C, t[18] = madd2(x[19], y[18], t[18], C)
	C, t[19] = madd2(x[19], y[19], t[19], C)
	C, t[20] = madd2(x[19], y[20], t[20], C)
	C, t[21] = madd2(x[19], y[21], t[21], C)
	C, t[22] = madd2(x[19], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[20], y[0], t[0])
	C, t[1] = madd2(x[20], y[1], t[1], C)
	C, t[2] = madd2(x[20], y[2], t[2], C)
	C, t[3] = madd2(x[20], y[3], t[3], C)
	C, t[4] = madd2(x[20], y[4], t[4], C)
	C, t[5] = madd2(x[20], y[5], t[5], C)
	C, t[6] = madd2(x[20], y[6], t[6], C)
	C, t[7] = madd2(x[20], y[7], t[7], C)
	C, t[8] = madd2(x[20], y[8], t[8], C)
	C, t[9] = madd2(x[20], y[9], t[9], C)
	C, t[10] = madd2(x[20], y[10], t[10], C)
	C, t[11] = madd2(x[20], y[11], t[11], C)
	C, t[12] = madd2(x[20], y[12], t[12], C)
	C, t[13] = madd2(x[20], y[13], t[13], C)
	C, t[14] = madd2(x[20], y[14], t[14], C)
	C, t[15] = madd2(x[20], y[15], t[15], C)
	C, t[16] = madd2(x[20], y[16], t[16], C)
	C, t[17] = madd2(x[20], y[17], t[17], C)
	C, t[18] = madd2(x[20], y[18], t[18], C)
	C, t[19] = madd2(x[20], y[19], t[19], C)
	C, t[20] = madd2(x[20], y[20], t[20], C)
	C, t[21] = madd2(x[20], y[21], t[21], C)
	C, t[22] = madd2(x[20], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[21], y[0], t[0])
	C, t[1] = madd2(x[21], y[1], t[1], C)
	C, t[2] = madd2(x[21], y[2], t[2], C)
	C, t[3] = madd2(x[21], y[3], t[3], C)
	C, t[4] = madd2(x[21], y[4], t[4], C)
	C, t[5] = madd2(x[21], y[5], t[5], C)
	C, t[6] = madd2(x[21], y[6], t[6], C)
	C, t[7] = madd2(x[21], y[7], t[7], C)
	C, t[8] = madd2(x[21], y[8], t[8], C)
	C, t[9] = madd2(x[21], y[9], t[9], C)
	C, t[10] = madd2(x[21], y[10], t[10], C)
	C, t[11] = madd2(x[21], y[11], t[11], C)
	C, t[12] = madd2(x[21], y[12], t[12], C)
	C, t[13] = madd2(x[21], y[13], t[13], C)
	C, t[14] = madd2(x[21], y[14], t[14], C)
	C, t[15] = madd2(x[21], y[15], t[15], C)
	C, t[16] = madd2(x[21], y[16], t[16], C)
	C, t[17] = madd2(x[21], y[17], t[17], C)
	C, t[18] = madd2(x[21], y[18], t[18], C)
	C, t[19] = madd2(x[21], y[19], t[19], C)
	C, t[20] = madd2(x[21], y[20], t[20], C)
	C, t[21] = madd2(x[21], y[21], t[21], C)
	C, t[22] = madd2(x[21], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[22], y[0], t[0])
	C, t[1] = madd2(x[22], y[1], t[1], C)
	C, t[2] = madd2(x[22], y[2], t[2], C)
	C, t[3] = madd2(x[22], y[3], t[3], C)
	C, t[4] = madd2(x[22], y[4], t[4], C)
	C, t[5] = madd2(x[22], y[5], t[5], C)
	C, t[6] = madd2(x[22], y[6], t[6], C)
	C, t[7] = madd2(x[22], y[7], t[7], C)
	C, t[8] = madd2(x[22], y[8], t[8], C)
	C, t[9] = madd2(x[22], y[9], t[9], C)
	C, t[10] = madd2(x[22], y[10], t[10], C)
	C, t[11] = madd2(x[22], y[11], t[11], C)
	C, t[12] = madd2(x[22], y[12], t[12], C)
	C, t[13] = madd2(x[22], y[13], t[13], C)
	C, t[14] = madd2(x[22], y[14], t[14], C)
	C, t[15] = madd2(x[22], y[15], t[15], C)
	C, t[16] = madd2(x[22], y[16], t[16], C)
	C, t[17] = madd2(x[22], y[17], t[17], C)
	C, t[18] = madd2(x[22], y[18], t[18], C)
	C, t[19] = madd2(x[22], y[19], t[19], C)
	C, t[20] = madd2(x[22], y[20], t[20], C)
	C, t[21] = madd2(x[22], y[21], t[21], C)
	C, t[22] = madd2(x[22], y[22], t[22], C)

	t[23], D = bits.Add64(t[23], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	t[22], C = bits.Add64(t[23], C, 0)
	t[23], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)
	z[18], D = bits.Sub64(t[18], mod[18], D)
	z[19], D = bits.Sub64(t[19], mod[19], D)
	z[20], D = bits.Sub64(t[20], mod[20], D)
	z[21], D = bits.Sub64(t[21], mod[21], D)
	z[22], D = bits.Sub64(t[22], mod[22], D)

	if D != 0 && t[23] == 0 {
		// reduction was not necessary
		copy(z[:], t[:23])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

func mulMont1536(ctx *Field, out_bytes, x_bytes, y_bytes []byte) error {
	x := (*[24]uint64)(unsafe.Pointer(&x_bytes[0]))[:]
	y := (*[24]uint64)(unsafe.Pointer(&y_bytes[0]))[:]
	z := (*[24]uint64)(unsafe.Pointer(&out_bytes[0]))[:]
	mod := (*[24]uint64)(unsafe.Pointer(&ctx.Modulus[0]))[:]
	var t [25]uint64
	var D uint64
	var m, C uint64

	var gteC1, gteC2 uint64
	_, gteC1 = bits.Sub64(mod[0], x[0], gteC1)
	_, gteC1 = bits.Sub64(mod[1], x[1], gteC1)
	_, gteC1 = bits.Sub64(mod[2], x[2], gteC1)
	_, gteC1 = bits.Sub64(mod[3], x[3], gteC1)
	_, gteC1 = bits.Sub64(mod[4], x[4], gteC1)
	_, gteC1 = bits.Sub64(mod[5], x[5], gteC1)
	_, gteC1 = bits.Sub64(mod[6], x[6], gteC1)
	_, gteC1 = bits.Sub64(mod[7], x[7], gteC1)
	_, gteC1 = bits.Sub64(mod[8], x[8], gteC1)
	_, gteC1 = bits.Sub64(mod[9], x[9], gteC1)
	_, gteC1 = bits.Sub64(mod[10], x[10], gteC1)
	_, gteC1 = bits.Sub64(mod[11], x[11], gteC1)
	_, gteC1 = bits.Sub64(mod[12], x[12], gteC1)
	_, gteC1 = bits.Sub64(mod[13], x[13], gteC1)
	_, gteC1 = bits.Sub64(mod[14], x[14], gteC1)
	_, gteC1 = bits.Sub64(mod[15], x[15], gteC1)
	_, gteC1 = bits.Sub64(mod[16], x[16], gteC1)
	_, gteC1 = bits.Sub64(mod[17], x[17], gteC1)
	_, gteC1 = bits.Sub64(mod[18], x[18], gteC1)
	_, gteC1 = bits.Sub64(mod[19], x[19], gteC1)
	_, gteC1 = bits.Sub64(mod[20], x[20], gteC1)
	_, gteC1 = bits.Sub64(mod[21], x[21], gteC1)
	_, gteC1 = bits.Sub64(mod[22], x[22], gteC1)
	_, gteC1 = bits.Sub64(mod[23], x[23], gteC1)
	_, gteC2 = bits.Sub64(mod[0], x[0], gteC2)
	_, gteC2 = bits.Sub64(mod[1], x[1], gteC2)
	_, gteC2 = bits.Sub64(mod[2], x[2], gteC2)
	_, gteC2 = bits.Sub64(mod[3], x[3], gteC2)
	_, gteC2 = bits.Sub64(mod[4], x[4], gteC2)
	_, gteC2 = bits.Sub64(mod[5], x[5], gteC2)
	_, gteC2 = bits.Sub64(mod[6], x[6], gteC2)
	_, gteC2 = bits.Sub64(mod[7], x[7], gteC2)
	_, gteC2 = bits.Sub64(mod[8], x[8], gteC2)
	_, gteC2 = bits.Sub64(mod[9], x[9], gteC2)
	_, gteC2 = bits.Sub64(mod[10], x[10], gteC2)
	_, gteC2 = bits.Sub64(mod[11], x[11], gteC2)
	_, gteC2 = bits.Sub64(mod[12], x[12], gteC2)
	_, gteC2 = bits.Sub64(mod[13], x[13], gteC2)
	_, gteC2 = bits.Sub64(mod[14], x[14], gteC2)
	_, gteC2 = bits.Sub64(mod[15], x[15], gteC2)
	_, gteC2 = bits.Sub64(mod[16], x[16], gteC2)
	_, gteC2 = bits.Sub64(mod[17], x[17], gteC2)
	_, gteC2 = bits.Sub64(mod[18], x[18], gteC2)
	_, gteC2 = bits.Sub64(mod[19], x[19], gteC2)
	_, gteC2 = bits.Sub64(mod[20], x[20], gteC2)
	_, gteC2 = bits.Sub64(mod[21], x[21], gteC2)
	_, gteC2 = bits.Sub64(mod[22], x[22], gteC2)
	_, gteC2 = bits.Sub64(mod[23], x[23], gteC2)

	if gteC1 != 0 || gteC2 != 0 {
		return errors.New(fmt.Sprintf("input gte modulus"))
	}

	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(x[0], y[0])
	C, t[1] = madd1(x[0], y[1], C)
	C, t[2] = madd1(x[0], y[2], C)
	C, t[3] = madd1(x[0], y[3], C)
	C, t[4] = madd1(x[0], y[4], C)
	C, t[5] = madd1(x[0], y[5], C)
	C, t[6] = madd1(x[0], y[6], C)
	C, t[7] = madd1(x[0], y[7], C)
	C, t[8] = madd1(x[0], y[8], C)
	C, t[9] = madd1(x[0], y[9], C)
	C, t[10] = madd1(x[0], y[10], C)
	C, t[11] = madd1(x[0], y[11], C)
	C, t[12] = madd1(x[0], y[12], C)
	C, t[13] = madd1(x[0], y[13], C)
	C, t[14] = madd1(x[0], y[14], C)
	C, t[15] = madd1(x[0], y[15], C)
	C, t[16] = madd1(x[0], y[16], C)
	C, t[17] = madd1(x[0], y[17], C)
	C, t[18] = madd1(x[0], y[18], C)
	C, t[19] = madd1(x[0], y[19], C)
	C, t[20] = madd1(x[0], y[20], C)
	C, t[21] = madd1(x[0], y[21], C)
	C, t[22] = madd1(x[0], y[22], C)
	C, t[23] = madd1(x[0], y[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[1], y[0], t[0])
	C, t[1] = madd2(x[1], y[1], t[1], C)
	C, t[2] = madd2(x[1], y[2], t[2], C)
	C, t[3] = madd2(x[1], y[3], t[3], C)
	C, t[4] = madd2(x[1], y[4], t[4], C)
	C, t[5] = madd2(x[1], y[5], t[5], C)
	C, t[6] = madd2(x[1], y[6], t[6], C)
	C, t[7] = madd2(x[1], y[7], t[7], C)
	C, t[8] = madd2(x[1], y[8], t[8], C)
	C, t[9] = madd2(x[1], y[9], t[9], C)
	C, t[10] = madd2(x[1], y[10], t[10], C)
	C, t[11] = madd2(x[1], y[11], t[11], C)
	C, t[12] = madd2(x[1], y[12], t[12], C)
	C, t[13] = madd2(x[1], y[13], t[13], C)
	C, t[14] = madd2(x[1], y[14], t[14], C)
	C, t[15] = madd2(x[1], y[15], t[15], C)
	C, t[16] = madd2(x[1], y[16], t[16], C)
	C, t[17] = madd2(x[1], y[17], t[17], C)
	C, t[18] = madd2(x[1], y[18], t[18], C)
	C, t[19] = madd2(x[1], y[19], t[19], C)
	C, t[20] = madd2(x[1], y[20], t[20], C)
	C, t[21] = madd2(x[1], y[21], t[21], C)
	C, t[22] = madd2(x[1], y[22], t[22], C)
	C, t[23] = madd2(x[1], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[2], y[0], t[0])
	C, t[1] = madd2(x[2], y[1], t[1], C)
	C, t[2] = madd2(x[2], y[2], t[2], C)
	C, t[3] = madd2(x[2], y[3], t[3], C)
	C, t[4] = madd2(x[2], y[4], t[4], C)
	C, t[5] = madd2(x[2], y[5], t[5], C)
	C, t[6] = madd2(x[2], y[6], t[6], C)
	C, t[7] = madd2(x[2], y[7], t[7], C)
	C, t[8] = madd2(x[2], y[8], t[8], C)
	C, t[9] = madd2(x[2], y[9], t[9], C)
	C, t[10] = madd2(x[2], y[10], t[10], C)
	C, t[11] = madd2(x[2], y[11], t[11], C)
	C, t[12] = madd2(x[2], y[12], t[12], C)
	C, t[13] = madd2(x[2], y[13], t[13], C)
	C, t[14] = madd2(x[2], y[14], t[14], C)
	C, t[15] = madd2(x[2], y[15], t[15], C)
	C, t[16] = madd2(x[2], y[16], t[16], C)
	C, t[17] = madd2(x[2], y[17], t[17], C)
	C, t[18] = madd2(x[2], y[18], t[18], C)
	C, t[19] = madd2(x[2], y[19], t[19], C)
	C, t[20] = madd2(x[2], y[20], t[20], C)
	C, t[21] = madd2(x[2], y[21], t[21], C)
	C, t[22] = madd2(x[2], y[22], t[22], C)
	C, t[23] = madd2(x[2], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[3], y[0], t[0])
	C, t[1] = madd2(x[3], y[1], t[1], C)
	C, t[2] = madd2(x[3], y[2], t[2], C)
	C, t[3] = madd2(x[3], y[3], t[3], C)
	C, t[4] = madd2(x[3], y[4], t[4], C)
	C, t[5] = madd2(x[3], y[5], t[5], C)
	C, t[6] = madd2(x[3], y[6], t[6], C)
	C, t[7] = madd2(x[3], y[7], t[7], C)
	C, t[8] = madd2(x[3], y[8], t[8], C)
	C, t[9] = madd2(x[3], y[9], t[9], C)
	C, t[10] = madd2(x[3], y[10], t[10], C)
	C, t[11] = madd2(x[3], y[11], t[11], C)
	C, t[12] = madd2(x[3], y[12], t[12], C)
	C, t[13] = madd2(x[3], y[13], t[13], C)
	C, t[14] = madd2(x[3], y[14], t[14], C)
	C, t[15] = madd2(x[3], y[15], t[15], C)
	C, t[16] = madd2(x[3], y[16], t[16], C)
	C, t[17] = madd2(x[3], y[17], t[17], C)
	C, t[18] = madd2(x[3], y[18], t[18], C)
	C, t[19] = madd2(x[3], y[19], t[19], C)
	C, t[20] = madd2(x[3], y[20], t[20], C)
	C, t[21] = madd2(x[3], y[21], t[21], C)
	C, t[22] = madd2(x[3], y[22], t[22], C)
	C, t[23] = madd2(x[3], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[4], y[0], t[0])
	C, t[1] = madd2(x[4], y[1], t[1], C)
	C, t[2] = madd2(x[4], y[2], t[2], C)
	C, t[3] = madd2(x[4], y[3], t[3], C)
	C, t[4] = madd2(x[4], y[4], t[4], C)
	C, t[5] = madd2(x[4], y[5], t[5], C)
	C, t[6] = madd2(x[4], y[6], t[6], C)
	C, t[7] = madd2(x[4], y[7], t[7], C)
	C, t[8] = madd2(x[4], y[8], t[8], C)
	C, t[9] = madd2(x[4], y[9], t[9], C)
	C, t[10] = madd2(x[4], y[10], t[10], C)
	C, t[11] = madd2(x[4], y[11], t[11], C)
	C, t[12] = madd2(x[4], y[12], t[12], C)
	C, t[13] = madd2(x[4], y[13], t[13], C)
	C, t[14] = madd2(x[4], y[14], t[14], C)
	C, t[15] = madd2(x[4], y[15], t[15], C)
	C, t[16] = madd2(x[4], y[16], t[16], C)
	C, t[17] = madd2(x[4], y[17], t[17], C)
	C, t[18] = madd2(x[4], y[18], t[18], C)
	C, t[19] = madd2(x[4], y[19], t[19], C)
	C, t[20] = madd2(x[4], y[20], t[20], C)
	C, t[21] = madd2(x[4], y[21], t[21], C)
	C, t[22] = madd2(x[4], y[22], t[22], C)
	C, t[23] = madd2(x[4], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[5], y[0], t[0])
	C, t[1] = madd2(x[5], y[1], t[1], C)
	C, t[2] = madd2(x[5], y[2], t[2], C)
	C, t[3] = madd2(x[5], y[3], t[3], C)
	C, t[4] = madd2(x[5], y[4], t[4], C)
	C, t[5] = madd2(x[5], y[5], t[5], C)
	C, t[6] = madd2(x[5], y[6], t[6], C)
	C, t[7] = madd2(x[5], y[7], t[7], C)
	C, t[8] = madd2(x[5], y[8], t[8], C)
	C, t[9] = madd2(x[5], y[9], t[9], C)
	C, t[10] = madd2(x[5], y[10], t[10], C)
	C, t[11] = madd2(x[5], y[11], t[11], C)
	C, t[12] = madd2(x[5], y[12], t[12], C)
	C, t[13] = madd2(x[5], y[13], t[13], C)
	C, t[14] = madd2(x[5], y[14], t[14], C)
	C, t[15] = madd2(x[5], y[15], t[15], C)
	C, t[16] = madd2(x[5], y[16], t[16], C)
	C, t[17] = madd2(x[5], y[17], t[17], C)
	C, t[18] = madd2(x[5], y[18], t[18], C)
	C, t[19] = madd2(x[5], y[19], t[19], C)
	C, t[20] = madd2(x[5], y[20], t[20], C)
	C, t[21] = madd2(x[5], y[21], t[21], C)
	C, t[22] = madd2(x[5], y[22], t[22], C)
	C, t[23] = madd2(x[5], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[6], y[0], t[0])
	C, t[1] = madd2(x[6], y[1], t[1], C)
	C, t[2] = madd2(x[6], y[2], t[2], C)
	C, t[3] = madd2(x[6], y[3], t[3], C)
	C, t[4] = madd2(x[6], y[4], t[4], C)
	C, t[5] = madd2(x[6], y[5], t[5], C)
	C, t[6] = madd2(x[6], y[6], t[6], C)
	C, t[7] = madd2(x[6], y[7], t[7], C)
	C, t[8] = madd2(x[6], y[8], t[8], C)
	C, t[9] = madd2(x[6], y[9], t[9], C)
	C, t[10] = madd2(x[6], y[10], t[10], C)
	C, t[11] = madd2(x[6], y[11], t[11], C)
	C, t[12] = madd2(x[6], y[12], t[12], C)
	C, t[13] = madd2(x[6], y[13], t[13], C)
	C, t[14] = madd2(x[6], y[14], t[14], C)
	C, t[15] = madd2(x[6], y[15], t[15], C)
	C, t[16] = madd2(x[6], y[16], t[16], C)
	C, t[17] = madd2(x[6], y[17], t[17], C)
	C, t[18] = madd2(x[6], y[18], t[18], C)
	C, t[19] = madd2(x[6], y[19], t[19], C)
	C, t[20] = madd2(x[6], y[20], t[20], C)
	C, t[21] = madd2(x[6], y[21], t[21], C)
	C, t[22] = madd2(x[6], y[22], t[22], C)
	C, t[23] = madd2(x[6], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[7], y[0], t[0])
	C, t[1] = madd2(x[7], y[1], t[1], C)
	C, t[2] = madd2(x[7], y[2], t[2], C)
	C, t[3] = madd2(x[7], y[3], t[3], C)
	C, t[4] = madd2(x[7], y[4], t[4], C)
	C, t[5] = madd2(x[7], y[5], t[5], C)
	C, t[6] = madd2(x[7], y[6], t[6], C)
	C, t[7] = madd2(x[7], y[7], t[7], C)
	C, t[8] = madd2(x[7], y[8], t[8], C)
	C, t[9] = madd2(x[7], y[9], t[9], C)
	C, t[10] = madd2(x[7], y[10], t[10], C)
	C, t[11] = madd2(x[7], y[11], t[11], C)
	C, t[12] = madd2(x[7], y[12], t[12], C)
	C, t[13] = madd2(x[7], y[13], t[13], C)
	C, t[14] = madd2(x[7], y[14], t[14], C)
	C, t[15] = madd2(x[7], y[15], t[15], C)
	C, t[16] = madd2(x[7], y[16], t[16], C)
	C, t[17] = madd2(x[7], y[17], t[17], C)
	C, t[18] = madd2(x[7], y[18], t[18], C)
	C, t[19] = madd2(x[7], y[19], t[19], C)
	C, t[20] = madd2(x[7], y[20], t[20], C)
	C, t[21] = madd2(x[7], y[21], t[21], C)
	C, t[22] = madd2(x[7], y[22], t[22], C)
	C, t[23] = madd2(x[7], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[8], y[0], t[0])
	C, t[1] = madd2(x[8], y[1], t[1], C)
	C, t[2] = madd2(x[8], y[2], t[2], C)
	C, t[3] = madd2(x[8], y[3], t[3], C)
	C, t[4] = madd2(x[8], y[4], t[4], C)
	C, t[5] = madd2(x[8], y[5], t[5], C)
	C, t[6] = madd2(x[8], y[6], t[6], C)
	C, t[7] = madd2(x[8], y[7], t[7], C)
	C, t[8] = madd2(x[8], y[8], t[8], C)
	C, t[9] = madd2(x[8], y[9], t[9], C)
	C, t[10] = madd2(x[8], y[10], t[10], C)
	C, t[11] = madd2(x[8], y[11], t[11], C)
	C, t[12] = madd2(x[8], y[12], t[12], C)
	C, t[13] = madd2(x[8], y[13], t[13], C)
	C, t[14] = madd2(x[8], y[14], t[14], C)
	C, t[15] = madd2(x[8], y[15], t[15], C)
	C, t[16] = madd2(x[8], y[16], t[16], C)
	C, t[17] = madd2(x[8], y[17], t[17], C)
	C, t[18] = madd2(x[8], y[18], t[18], C)
	C, t[19] = madd2(x[8], y[19], t[19], C)
	C, t[20] = madd2(x[8], y[20], t[20], C)
	C, t[21] = madd2(x[8], y[21], t[21], C)
	C, t[22] = madd2(x[8], y[22], t[22], C)
	C, t[23] = madd2(x[8], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[9], y[0], t[0])
	C, t[1] = madd2(x[9], y[1], t[1], C)
	C, t[2] = madd2(x[9], y[2], t[2], C)
	C, t[3] = madd2(x[9], y[3], t[3], C)
	C, t[4] = madd2(x[9], y[4], t[4], C)
	C, t[5] = madd2(x[9], y[5], t[5], C)
	C, t[6] = madd2(x[9], y[6], t[6], C)
	C, t[7] = madd2(x[9], y[7], t[7], C)
	C, t[8] = madd2(x[9], y[8], t[8], C)
	C, t[9] = madd2(x[9], y[9], t[9], C)
	C, t[10] = madd2(x[9], y[10], t[10], C)
	C, t[11] = madd2(x[9], y[11], t[11], C)
	C, t[12] = madd2(x[9], y[12], t[12], C)
	C, t[13] = madd2(x[9], y[13], t[13], C)
	C, t[14] = madd2(x[9], y[14], t[14], C)
	C, t[15] = madd2(x[9], y[15], t[15], C)
	C, t[16] = madd2(x[9], y[16], t[16], C)
	C, t[17] = madd2(x[9], y[17], t[17], C)
	C, t[18] = madd2(x[9], y[18], t[18], C)
	C, t[19] = madd2(x[9], y[19], t[19], C)
	C, t[20] = madd2(x[9], y[20], t[20], C)
	C, t[21] = madd2(x[9], y[21], t[21], C)
	C, t[22] = madd2(x[9], y[22], t[22], C)
	C, t[23] = madd2(x[9], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[10], y[0], t[0])
	C, t[1] = madd2(x[10], y[1], t[1], C)
	C, t[2] = madd2(x[10], y[2], t[2], C)
	C, t[3] = madd2(x[10], y[3], t[3], C)
	C, t[4] = madd2(x[10], y[4], t[4], C)
	C, t[5] = madd2(x[10], y[5], t[5], C)
	C, t[6] = madd2(x[10], y[6], t[6], C)
	C, t[7] = madd2(x[10], y[7], t[7], C)
	C, t[8] = madd2(x[10], y[8], t[8], C)
	C, t[9] = madd2(x[10], y[9], t[9], C)
	C, t[10] = madd2(x[10], y[10], t[10], C)
	C, t[11] = madd2(x[10], y[11], t[11], C)
	C, t[12] = madd2(x[10], y[12], t[12], C)
	C, t[13] = madd2(x[10], y[13], t[13], C)
	C, t[14] = madd2(x[10], y[14], t[14], C)
	C, t[15] = madd2(x[10], y[15], t[15], C)
	C, t[16] = madd2(x[10], y[16], t[16], C)
	C, t[17] = madd2(x[10], y[17], t[17], C)
	C, t[18] = madd2(x[10], y[18], t[18], C)
	C, t[19] = madd2(x[10], y[19], t[19], C)
	C, t[20] = madd2(x[10], y[20], t[20], C)
	C, t[21] = madd2(x[10], y[21], t[21], C)
	C, t[22] = madd2(x[10], y[22], t[22], C)
	C, t[23] = madd2(x[10], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[11], y[0], t[0])
	C, t[1] = madd2(x[11], y[1], t[1], C)
	C, t[2] = madd2(x[11], y[2], t[2], C)
	C, t[3] = madd2(x[11], y[3], t[3], C)
	C, t[4] = madd2(x[11], y[4], t[4], C)
	C, t[5] = madd2(x[11], y[5], t[5], C)
	C, t[6] = madd2(x[11], y[6], t[6], C)
	C, t[7] = madd2(x[11], y[7], t[7], C)
	C, t[8] = madd2(x[11], y[8], t[8], C)
	C, t[9] = madd2(x[11], y[9], t[9], C)
	C, t[10] = madd2(x[11], y[10], t[10], C)
	C, t[11] = madd2(x[11], y[11], t[11], C)
	C, t[12] = madd2(x[11], y[12], t[12], C)
	C, t[13] = madd2(x[11], y[13], t[13], C)
	C, t[14] = madd2(x[11], y[14], t[14], C)
	C, t[15] = madd2(x[11], y[15], t[15], C)
	C, t[16] = madd2(x[11], y[16], t[16], C)
	C, t[17] = madd2(x[11], y[17], t[17], C)
	C, t[18] = madd2(x[11], y[18], t[18], C)
	C, t[19] = madd2(x[11], y[19], t[19], C)
	C, t[20] = madd2(x[11], y[20], t[20], C)
	C, t[21] = madd2(x[11], y[21], t[21], C)
	C, t[22] = madd2(x[11], y[22], t[22], C)
	C, t[23] = madd2(x[11], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[12], y[0], t[0])
	C, t[1] = madd2(x[12], y[1], t[1], C)
	C, t[2] = madd2(x[12], y[2], t[2], C)
	C, t[3] = madd2(x[12], y[3], t[3], C)
	C, t[4] = madd2(x[12], y[4], t[4], C)
	C, t[5] = madd2(x[12], y[5], t[5], C)
	C, t[6] = madd2(x[12], y[6], t[6], C)
	C, t[7] = madd2(x[12], y[7], t[7], C)
	C, t[8] = madd2(x[12], y[8], t[8], C)
	C, t[9] = madd2(x[12], y[9], t[9], C)
	C, t[10] = madd2(x[12], y[10], t[10], C)
	C, t[11] = madd2(x[12], y[11], t[11], C)
	C, t[12] = madd2(x[12], y[12], t[12], C)
	C, t[13] = madd2(x[12], y[13], t[13], C)
	C, t[14] = madd2(x[12], y[14], t[14], C)
	C, t[15] = madd2(x[12], y[15], t[15], C)
	C, t[16] = madd2(x[12], y[16], t[16], C)
	C, t[17] = madd2(x[12], y[17], t[17], C)
	C, t[18] = madd2(x[12], y[18], t[18], C)
	C, t[19] = madd2(x[12], y[19], t[19], C)
	C, t[20] = madd2(x[12], y[20], t[20], C)
	C, t[21] = madd2(x[12], y[21], t[21], C)
	C, t[22] = madd2(x[12], y[22], t[22], C)
	C, t[23] = madd2(x[12], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[13], y[0], t[0])
	C, t[1] = madd2(x[13], y[1], t[1], C)
	C, t[2] = madd2(x[13], y[2], t[2], C)
	C, t[3] = madd2(x[13], y[3], t[3], C)
	C, t[4] = madd2(x[13], y[4], t[4], C)
	C, t[5] = madd2(x[13], y[5], t[5], C)
	C, t[6] = madd2(x[13], y[6], t[6], C)
	C, t[7] = madd2(x[13], y[7], t[7], C)
	C, t[8] = madd2(x[13], y[8], t[8], C)
	C, t[9] = madd2(x[13], y[9], t[9], C)
	C, t[10] = madd2(x[13], y[10], t[10], C)
	C, t[11] = madd2(x[13], y[11], t[11], C)
	C, t[12] = madd2(x[13], y[12], t[12], C)
	C, t[13] = madd2(x[13], y[13], t[13], C)
	C, t[14] = madd2(x[13], y[14], t[14], C)
	C, t[15] = madd2(x[13], y[15], t[15], C)
	C, t[16] = madd2(x[13], y[16], t[16], C)
	C, t[17] = madd2(x[13], y[17], t[17], C)
	C, t[18] = madd2(x[13], y[18], t[18], C)
	C, t[19] = madd2(x[13], y[19], t[19], C)
	C, t[20] = madd2(x[13], y[20], t[20], C)
	C, t[21] = madd2(x[13], y[21], t[21], C)
	C, t[22] = madd2(x[13], y[22], t[22], C)
	C, t[23] = madd2(x[13], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[14], y[0], t[0])
	C, t[1] = madd2(x[14], y[1], t[1], C)
	C, t[2] = madd2(x[14], y[2], t[2], C)
	C, t[3] = madd2(x[14], y[3], t[3], C)
	C, t[4] = madd2(x[14], y[4], t[4], C)
	C, t[5] = madd2(x[14], y[5], t[5], C)
	C, t[6] = madd2(x[14], y[6], t[6], C)
	C, t[7] = madd2(x[14], y[7], t[7], C)
	C, t[8] = madd2(x[14], y[8], t[8], C)
	C, t[9] = madd2(x[14], y[9], t[9], C)
	C, t[10] = madd2(x[14], y[10], t[10], C)
	C, t[11] = madd2(x[14], y[11], t[11], C)
	C, t[12] = madd2(x[14], y[12], t[12], C)
	C, t[13] = madd2(x[14], y[13], t[13], C)
	C, t[14] = madd2(x[14], y[14], t[14], C)
	C, t[15] = madd2(x[14], y[15], t[15], C)
	C, t[16] = madd2(x[14], y[16], t[16], C)
	C, t[17] = madd2(x[14], y[17], t[17], C)
	C, t[18] = madd2(x[14], y[18], t[18], C)
	C, t[19] = madd2(x[14], y[19], t[19], C)
	C, t[20] = madd2(x[14], y[20], t[20], C)
	C, t[21] = madd2(x[14], y[21], t[21], C)
	C, t[22] = madd2(x[14], y[22], t[22], C)
	C, t[23] = madd2(x[14], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[15], y[0], t[0])
	C, t[1] = madd2(x[15], y[1], t[1], C)
	C, t[2] = madd2(x[15], y[2], t[2], C)
	C, t[3] = madd2(x[15], y[3], t[3], C)
	C, t[4] = madd2(x[15], y[4], t[4], C)
	C, t[5] = madd2(x[15], y[5], t[5], C)
	C, t[6] = madd2(x[15], y[6], t[6], C)
	C, t[7] = madd2(x[15], y[7], t[7], C)
	C, t[8] = madd2(x[15], y[8], t[8], C)
	C, t[9] = madd2(x[15], y[9], t[9], C)
	C, t[10] = madd2(x[15], y[10], t[10], C)
	C, t[11] = madd2(x[15], y[11], t[11], C)
	C, t[12] = madd2(x[15], y[12], t[12], C)
	C, t[13] = madd2(x[15], y[13], t[13], C)
	C, t[14] = madd2(x[15], y[14], t[14], C)
	C, t[15] = madd2(x[15], y[15], t[15], C)
	C, t[16] = madd2(x[15], y[16], t[16], C)
	C, t[17] = madd2(x[15], y[17], t[17], C)
	C, t[18] = madd2(x[15], y[18], t[18], C)
	C, t[19] = madd2(x[15], y[19], t[19], C)
	C, t[20] = madd2(x[15], y[20], t[20], C)
	C, t[21] = madd2(x[15], y[21], t[21], C)
	C, t[22] = madd2(x[15], y[22], t[22], C)
	C, t[23] = madd2(x[15], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[16], y[0], t[0])
	C, t[1] = madd2(x[16], y[1], t[1], C)
	C, t[2] = madd2(x[16], y[2], t[2], C)
	C, t[3] = madd2(x[16], y[3], t[3], C)
	C, t[4] = madd2(x[16], y[4], t[4], C)
	C, t[5] = madd2(x[16], y[5], t[5], C)
	C, t[6] = madd2(x[16], y[6], t[6], C)
	C, t[7] = madd2(x[16], y[7], t[7], C)
	C, t[8] = madd2(x[16], y[8], t[8], C)
	C, t[9] = madd2(x[16], y[9], t[9], C)
	C, t[10] = madd2(x[16], y[10], t[10], C)
	C, t[11] = madd2(x[16], y[11], t[11], C)
	C, t[12] = madd2(x[16], y[12], t[12], C)
	C, t[13] = madd2(x[16], y[13], t[13], C)
	C, t[14] = madd2(x[16], y[14], t[14], C)
	C, t[15] = madd2(x[16], y[15], t[15], C)
	C, t[16] = madd2(x[16], y[16], t[16], C)
	C, t[17] = madd2(x[16], y[17], t[17], C)
	C, t[18] = madd2(x[16], y[18], t[18], C)
	C, t[19] = madd2(x[16], y[19], t[19], C)
	C, t[20] = madd2(x[16], y[20], t[20], C)
	C, t[21] = madd2(x[16], y[21], t[21], C)
	C, t[22] = madd2(x[16], y[22], t[22], C)
	C, t[23] = madd2(x[16], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[17], y[0], t[0])
	C, t[1] = madd2(x[17], y[1], t[1], C)
	C, t[2] = madd2(x[17], y[2], t[2], C)
	C, t[3] = madd2(x[17], y[3], t[3], C)
	C, t[4] = madd2(x[17], y[4], t[4], C)
	C, t[5] = madd2(x[17], y[5], t[5], C)
	C, t[6] = madd2(x[17], y[6], t[6], C)
	C, t[7] = madd2(x[17], y[7], t[7], C)
	C, t[8] = madd2(x[17], y[8], t[8], C)
	C, t[9] = madd2(x[17], y[9], t[9], C)
	C, t[10] = madd2(x[17], y[10], t[10], C)
	C, t[11] = madd2(x[17], y[11], t[11], C)
	C, t[12] = madd2(x[17], y[12], t[12], C)
	C, t[13] = madd2(x[17], y[13], t[13], C)
	C, t[14] = madd2(x[17], y[14], t[14], C)
	C, t[15] = madd2(x[17], y[15], t[15], C)
	C, t[16] = madd2(x[17], y[16], t[16], C)
	C, t[17] = madd2(x[17], y[17], t[17], C)
	C, t[18] = madd2(x[17], y[18], t[18], C)
	C, t[19] = madd2(x[17], y[19], t[19], C)
	C, t[20] = madd2(x[17], y[20], t[20], C)
	C, t[21] = madd2(x[17], y[21], t[21], C)
	C, t[22] = madd2(x[17], y[22], t[22], C)
	C, t[23] = madd2(x[17], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[18], y[0], t[0])
	C, t[1] = madd2(x[18], y[1], t[1], C)
	C, t[2] = madd2(x[18], y[2], t[2], C)
	C, t[3] = madd2(x[18], y[3], t[3], C)
	C, t[4] = madd2(x[18], y[4], t[4], C)
	C, t[5] = madd2(x[18], y[5], t[5], C)
	C, t[6] = madd2(x[18], y[6], t[6], C)
	C, t[7] = madd2(x[18], y[7], t[7], C)
	C, t[8] = madd2(x[18], y[8], t[8], C)
	C, t[9] = madd2(x[18], y[9], t[9], C)
	C, t[10] = madd2(x[18], y[10], t[10], C)
	C, t[11] = madd2(x[18], y[11], t[11], C)
	C, t[12] = madd2(x[18], y[12], t[12], C)
	C, t[13] = madd2(x[18], y[13], t[13], C)
	C, t[14] = madd2(x[18], y[14], t[14], C)
	C, t[15] = madd2(x[18], y[15], t[15], C)
	C, t[16] = madd2(x[18], y[16], t[16], C)
	C, t[17] = madd2(x[18], y[17], t[17], C)
	C, t[18] = madd2(x[18], y[18], t[18], C)
	C, t[19] = madd2(x[18], y[19], t[19], C)
	C, t[20] = madd2(x[18], y[20], t[20], C)
	C, t[21] = madd2(x[18], y[21], t[21], C)
	C, t[22] = madd2(x[18], y[22], t[22], C)
	C, t[23] = madd2(x[18], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[19], y[0], t[0])
	C, t[1] = madd2(x[19], y[1], t[1], C)
	C, t[2] = madd2(x[19], y[2], t[2], C)
	C, t[3] = madd2(x[19], y[3], t[3], C)
	C, t[4] = madd2(x[19], y[4], t[4], C)
	C, t[5] = madd2(x[19], y[5], t[5], C)
	C, t[6] = madd2(x[19], y[6], t[6], C)
	C, t[7] = madd2(x[19], y[7], t[7], C)
	C, t[8] = madd2(x[19], y[8], t[8], C)
	C, t[9] = madd2(x[19], y[9], t[9], C)
	C, t[10] = madd2(x[19], y[10], t[10], C)
	C, t[11] = madd2(x[19], y[11], t[11], C)
	C, t[12] = madd2(x[19], y[12], t[12], C)
	C, t[13] = madd2(x[19], y[13], t[13], C)
	C, t[14] = madd2(x[19], y[14], t[14], C)
	C, t[15] = madd2(x[19], y[15], t[15], C)
	C, t[16] = madd2(x[19], y[16], t[16], C)
	C, t[17] = madd2(x[19], y[17], t[17], C)
	C, t[18] = madd2(x[19], y[18], t[18], C)
	C, t[19] = madd2(x[19], y[19], t[19], C)
	C, t[20] = madd2(x[19], y[20], t[20], C)
	C, t[21] = madd2(x[19], y[21], t[21], C)
	C, t[22] = madd2(x[19], y[22], t[22], C)
	C, t[23] = madd2(x[19], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[20], y[0], t[0])
	C, t[1] = madd2(x[20], y[1], t[1], C)
	C, t[2] = madd2(x[20], y[2], t[2], C)
	C, t[3] = madd2(x[20], y[3], t[3], C)
	C, t[4] = madd2(x[20], y[4], t[4], C)
	C, t[5] = madd2(x[20], y[5], t[5], C)
	C, t[6] = madd2(x[20], y[6], t[6], C)
	C, t[7] = madd2(x[20], y[7], t[7], C)
	C, t[8] = madd2(x[20], y[8], t[8], C)
	C, t[9] = madd2(x[20], y[9], t[9], C)
	C, t[10] = madd2(x[20], y[10], t[10], C)
	C, t[11] = madd2(x[20], y[11], t[11], C)
	C, t[12] = madd2(x[20], y[12], t[12], C)
	C, t[13] = madd2(x[20], y[13], t[13], C)
	C, t[14] = madd2(x[20], y[14], t[14], C)
	C, t[15] = madd2(x[20], y[15], t[15], C)
	C, t[16] = madd2(x[20], y[16], t[16], C)
	C, t[17] = madd2(x[20], y[17], t[17], C)
	C, t[18] = madd2(x[20], y[18], t[18], C)
	C, t[19] = madd2(x[20], y[19], t[19], C)
	C, t[20] = madd2(x[20], y[20], t[20], C)
	C, t[21] = madd2(x[20], y[21], t[21], C)
	C, t[22] = madd2(x[20], y[22], t[22], C)
	C, t[23] = madd2(x[20], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[21], y[0], t[0])
	C, t[1] = madd2(x[21], y[1], t[1], C)
	C, t[2] = madd2(x[21], y[2], t[2], C)
	C, t[3] = madd2(x[21], y[3], t[3], C)
	C, t[4] = madd2(x[21], y[4], t[4], C)
	C, t[5] = madd2(x[21], y[5], t[5], C)
	C, t[6] = madd2(x[21], y[6], t[6], C)
	C, t[7] = madd2(x[21], y[7], t[7], C)
	C, t[8] = madd2(x[21], y[8], t[8], C)
	C, t[9] = madd2(x[21], y[9], t[9], C)
	C, t[10] = madd2(x[21], y[10], t[10], C)
	C, t[11] = madd2(x[21], y[11], t[11], C)
	C, t[12] = madd2(x[21], y[12], t[12], C)
	C, t[13] = madd2(x[21], y[13], t[13], C)
	C, t[14] = madd2(x[21], y[14], t[14], C)
	C, t[15] = madd2(x[21], y[15], t[15], C)
	C, t[16] = madd2(x[21], y[16], t[16], C)
	C, t[17] = madd2(x[21], y[17], t[17], C)
	C, t[18] = madd2(x[21], y[18], t[18], C)
	C, t[19] = madd2(x[21], y[19], t[19], C)
	C, t[20] = madd2(x[21], y[20], t[20], C)
	C, t[21] = madd2(x[21], y[21], t[21], C)
	C, t[22] = madd2(x[21], y[22], t[22], C)
	C, t[23] = madd2(x[21], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[22], y[0], t[0])
	C, t[1] = madd2(x[22], y[1], t[1], C)
	C, t[2] = madd2(x[22], y[2], t[2], C)
	C, t[3] = madd2(x[22], y[3], t[3], C)
	C, t[4] = madd2(x[22], y[4], t[4], C)
	C, t[5] = madd2(x[22], y[5], t[5], C)
	C, t[6] = madd2(x[22], y[6], t[6], C)
	C, t[7] = madd2(x[22], y[7], t[7], C)
	C, t[8] = madd2(x[22], y[8], t[8], C)
	C, t[9] = madd2(x[22], y[9], t[9], C)
	C, t[10] = madd2(x[22], y[10], t[10], C)
	C, t[11] = madd2(x[22], y[11], t[11], C)
	C, t[12] = madd2(x[22], y[12], t[12], C)
	C, t[13] = madd2(x[22], y[13], t[13], C)
	C, t[14] = madd2(x[22], y[14], t[14], C)
	C, t[15] = madd2(x[22], y[15], t[15], C)
	C, t[16] = madd2(x[22], y[16], t[16], C)
	C, t[17] = madd2(x[22], y[17], t[17], C)
	C, t[18] = madd2(x[22], y[18], t[18], C)
	C, t[19] = madd2(x[22], y[19], t[19], C)
	C, t[20] = madd2(x[22], y[20], t[20], C)
	C, t[21] = madd2(x[22], y[21], t[21], C)
	C, t[22] = madd2(x[22], y[22], t[22], C)
	C, t[23] = madd2(x[22], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(x[23], y[0], t[0])
	C, t[1] = madd2(x[23], y[1], t[1], C)
	C, t[2] = madd2(x[23], y[2], t[2], C)
	C, t[3] = madd2(x[23], y[3], t[3], C)
	C, t[4] = madd2(x[23], y[4], t[4], C)
	C, t[5] = madd2(x[23], y[5], t[5], C)
	C, t[6] = madd2(x[23], y[6], t[6], C)
	C, t[7] = madd2(x[23], y[7], t[7], C)
	C, t[8] = madd2(x[23], y[8], t[8], C)
	C, t[9] = madd2(x[23], y[9], t[9], C)
	C, t[10] = madd2(x[23], y[10], t[10], C)
	C, t[11] = madd2(x[23], y[11], t[11], C)
	C, t[12] = madd2(x[23], y[12], t[12], C)
	C, t[13] = madd2(x[23], y[13], t[13], C)
	C, t[14] = madd2(x[23], y[14], t[14], C)
	C, t[15] = madd2(x[23], y[15], t[15], C)
	C, t[16] = madd2(x[23], y[16], t[16], C)
	C, t[17] = madd2(x[23], y[17], t[17], C)
	C, t[18] = madd2(x[23], y[18], t[18], C)
	C, t[19] = madd2(x[23], y[19], t[19], C)
	C, t[20] = madd2(x[23], y[20], t[20], C)
	C, t[21] = madd2(x[23], y[21], t[21], C)
	C, t[22] = madd2(x[23], y[22], t[22], C)
	C, t[23] = madd2(x[23], y[23], t[23], C)

	t[24], D = bits.Add64(t[24], C, 0)
	// m = t[0]n'[0] mod W
	m = t[0] * ctx.MontParamInterleaved
	// -----------------------------------
	// Second loop
	C = madd0(m, mod[0], t[0])
	C, t[0] = madd2(m, mod[1], t[1], C)
	C, t[1] = madd2(m, mod[2], t[2], C)
	C, t[2] = madd2(m, mod[3], t[3], C)
	C, t[3] = madd2(m, mod[4], t[4], C)
	C, t[4] = madd2(m, mod[5], t[5], C)
	C, t[5] = madd2(m, mod[6], t[6], C)
	C, t[6] = madd2(m, mod[7], t[7], C)
	C, t[7] = madd2(m, mod[8], t[8], C)
	C, t[8] = madd2(m, mod[9], t[9], C)
	C, t[9] = madd2(m, mod[10], t[10], C)
	C, t[10] = madd2(m, mod[11], t[11], C)
	C, t[11] = madd2(m, mod[12], t[12], C)
	C, t[12] = madd2(m, mod[13], t[13], C)
	C, t[13] = madd2(m, mod[14], t[14], C)
	C, t[14] = madd2(m, mod[15], t[15], C)
	C, t[15] = madd2(m, mod[16], t[16], C)
	C, t[16] = madd2(m, mod[17], t[17], C)
	C, t[17] = madd2(m, mod[18], t[18], C)
	C, t[18] = madd2(m, mod[19], t[19], C)
	C, t[19] = madd2(m, mod[20], t[20], C)
	C, t[20] = madd2(m, mod[21], t[21], C)
	C, t[21] = madd2(m, mod[22], t[22], C)
	C, t[22] = madd2(m, mod[23], t[23], C)
	t[23], C = bits.Add64(t[24], C, 0)
	t[24], _ = bits.Add64(0, D, C)
	z[0], D = bits.Sub64(t[0], mod[0], 0)
	z[1], D = bits.Sub64(t[1], mod[1], D)
	z[2], D = bits.Sub64(t[2], mod[2], D)
	z[3], D = bits.Sub64(t[3], mod[3], D)
	z[4], D = bits.Sub64(t[4], mod[4], D)
	z[5], D = bits.Sub64(t[5], mod[5], D)
	z[6], D = bits.Sub64(t[6], mod[6], D)
	z[7], D = bits.Sub64(t[7], mod[7], D)
	z[8], D = bits.Sub64(t[8], mod[8], D)
	z[9], D = bits.Sub64(t[9], mod[9], D)
	z[10], D = bits.Sub64(t[10], mod[10], D)
	z[11], D = bits.Sub64(t[11], mod[11], D)
	z[12], D = bits.Sub64(t[12], mod[12], D)
	z[13], D = bits.Sub64(t[13], mod[13], D)
	z[14], D = bits.Sub64(t[14], mod[14], D)
	z[15], D = bits.Sub64(t[15], mod[15], D)
	z[16], D = bits.Sub64(t[16], mod[16], D)
	z[17], D = bits.Sub64(t[17], mod[17], D)
	z[18], D = bits.Sub64(t[18], mod[18], D)
	z[19], D = bits.Sub64(t[19], mod[19], D)
	z[20], D = bits.Sub64(t[20], mod[20], D)
	z[21], D = bits.Sub64(t[21], mod[21], D)
	z[22], D = bits.Sub64(t[22], mod[22], D)
	z[23], D = bits.Sub64(t[23], mod[23], D)

	if D != 0 && t[24] == 0 {
		// reduction was not necessary
		copy(z[:], t[:24])
	} /* else {
	    panic("not worst case performance")
	}*/

	return nil
}

// NOTE: this assumes that x and y are in Montgomery form and can produce unexpected results when they are not
func MulMontNonInterleaved(f *Field, out_bytes, x_bytes, y_bytes []byte) error {
	// length x == y assumed

	product := new(big.Int)
	x := LEBytesToInt(x_bytes)
	y := LEBytesToInt(y_bytes)

	if x.Cmp(f.ModulusNonInterleaved) > 0 || y.Cmp(f.ModulusNonInterleaved) > 0 {
		return errors.New("x/y >= modulus")
	}

	// m <- ((x*y mod R)N`) mod R
	product.Mul(x, y)
	x.And(product, f.mask)
	x.Mul(x, f.MontParamNonInterleaved)
	x.And(x, f.mask)

	// t <- (T + mN) / R
	x.Mul(x, f.ModulusNonInterleaved)
	x.Add(x, product)
	x.Rsh(x, f.NumLimbs*64)

	if x.Cmp(f.ModulusNonInterleaved) >= 0 {
		x.Sub(x, f.ModulusNonInterleaved)
	}

	copy(out_bytes, LimbsToLEBytes(IntToLimbs(x, f.NumLimbs)))

	return nil
}
