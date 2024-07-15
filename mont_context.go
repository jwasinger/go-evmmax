package evmmax_arith

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
)

const maxModulusSize = 96 // 768 bits maximum modulus width

type FieldContext struct {
	Modulus []uint64
	R2      []uint64
	modInv  uint64

	scratchSpace []uint64
	AddSubCost   uint64
	MulCost      uint64

	addMod addOrSubFunc
	subMod addOrSubFunc
	mulMod mulFunc

	one                   []uint64
	modulusInt            *big.Int
	elemSize              uint
	scratchSpaceElemCount uint
}

func NewFieldContext(modBytes []byte, scratchSize int) (*FieldContext, error) {
	if len(modBytes) > maxModulusSize {
		return nil, errors.New("modulus cannot be greater than 768 bits")
	}
	if modBytes[len(modBytes)-1]%2 == 0 {
		return nil, errors.New("modulus cannot be even")
	}
	if modBytes[0] == 0 {
		return nil, errors.New("most significant byte of modulus must not be zero")
	}
	if scratchSize > 256 {
		return nil, errors.New("scratch space can allocate a maximum of 256 field elements")
	}

	mod := new(big.Int).SetBytes(modBytes)
	paddedSize := int(math.Ceil(float64(len(modBytes))/8.0)) * 8
	modInv := negModInverse(mod.Uint64())

	r2 := new(big.Int).Lsh(big.NewInt(1), uint(paddedSize)*8*2)
	r2.Mod(r2, mod)

	r2Bytes := r2.Bytes()
	if len(modBytes) < paddedSize {
		modBytes = append(make([]byte, paddedSize-len(modBytes)), modBytes...)
	}
	if len(r2Bytes) < paddedSize {
		r2Bytes = append(make([]byte, paddedSize-len(r2Bytes)), r2Bytes...)
	}

	one := make([]uint64, paddedSize/8)
	one[0] = 1

	m := FieldContext{
		Modulus:               bytesToLimbs(modBytes),
		modInv:                modInv,
		R2:                    bytesToLimbs(r2Bytes),
		mulMod:                mulmodPreset[paddedSize/8-1],
		addMod:                addmodPreset[paddedSize/8-1],
		subMod:                submodPreset[paddedSize/8-1],
		scratchSpace:          make([]uint64, (paddedSize/8)*scratchSize),
		scratchSpaceElemCount: uint(scratchSize),
		one:                   one,
		modulusInt:            mod,
		elemSize:              uint(paddedSize),
	}
	return &m, nil
}

func (f *FieldContext) NumElems() uint {
	return f.scratchSpaceElemCount
}

func (f *FieldContext) AllocedSize() uint {
	return uint(len(f.scratchSpace) * 8)
}

// elem size in bytes
func (f *FieldContext) ElemSize() uint {
	return f.elemSize
}

// compute -mod ** -1 % 1 << 64 .
// from (paper), used in go-stdlib
func negModInverse(mod uint64) uint64 {
	k0 := 2 - mod
	t := mod - 1
	for i := 1; i < bits.UintSize; i <<= 1 {
		t *= t
		k0 *= (t + 1)
	}
	k0 = -k0
	return k0
}

// Note: manually inlining the arith funcs here into the opcode handler seems to give overall ~6-7% performance increase on g2 mul
// benchmark
func (m *FieldContext) MulMod(out, x, y uint) error {
	elemSize := uint(len(m.Modulus))

	if greatest := max(out, x, y); greatest >= m.scratchSpaceElemCount {
		return errors.New("out of bounds field element access")
	}
	//fmt.Printf("mulmodx. x(%d)=%s, y(%d)=%s", x, limbsToInt(m.scratchSpace[x*elemSize:(x+1)*elemSize]), y, limbsToInt(m.scratchSpace[y*elemSize:(y+1)*elemSize]))
	m.mulMod(m.scratchSpace[out*elemSize:(out+1)*elemSize],
		m.scratchSpace[x*elemSize:(x+1)*elemSize],
		m.scratchSpace[y*elemSize:(y+1)*elemSize],
		m.Modulus,
		m.modInv)
	//fmt.Printf(" out(%d)=%s\n", out, limbsToInt(m.scratchSpace[out*elemSize:(out+1)*elemSize]))
	return nil
}

func (m *FieldContext) SubMod(out, x, y uint) error {
	elemSize := uint(len(m.Modulus))
	if greatest := max(out, x, y); greatest >= m.scratchSpaceElemCount {
		return errors.New("out of bounds field element access")
	}
	//fmt.Printf("submodx. x(%d)=%s, y(%d)=%s", x, limbsToInt(m.scratchSpace[x*elemSize:(x+1)*elemSize]), y, limbsToInt(m.scratchSpace[y*elemSize:(y+1)*elemSize]))
	m.subMod(m.scratchSpace[out*elemSize:(out+1)*elemSize],
		m.scratchSpace[x*elemSize:(x+1)*elemSize],
		m.scratchSpace[y*elemSize:(y+1)*elemSize],
		m.Modulus)
	//fmt.Printf(" out=%s\n", limbsToInt(m.scratchSpace[out*elemSize:(out+1)*elemSize]))
	return nil
}

func (m *FieldContext) AddMod(out, x, y uint) error {
	elemSize := uint(len(m.Modulus))
	if greatest := max(out, x, y); greatest >= m.scratchSpaceElemCount {
		return errors.New("out of bounds field element access")
	}
	//fmt.Printf("addmodx. x(%d)=%s, y(%d)=%s", x, limbsToInt(m.scratchSpace[x*elemSize:(x+1)*elemSize]), y, limbsToInt(m.scratchSpace[y*elemSize:(y+1)*elemSize]))
	m.addMod(m.scratchSpace[out*elemSize:(out+1)*elemSize],
		m.scratchSpace[x*elemSize:(x+1)*elemSize],
		m.scratchSpace[y*elemSize:(y+1)*elemSize],
		m.Modulus)
	//fmt.Printf(" out=%s\n", limbsToInt(m.scratchSpace[out*elemSize:(out+1)*elemSize]))
	return nil
}

func (m *FieldContext) Store(dst, count uint, from []byte) error {
	elemSize := uint(len(m.Modulus))
	dstIdx := dst * elemSize

	if dstIdx+count > m.scratchSpaceElemCount {
		return errors.New("out of bounds field element store")
	}
	for i := uint(0); i < count; i++ {
		srcIdx := i * elemSize * 8
		dstIdx := dst*elemSize + i*elemSize

		// convert the big-endian bytes to little-endian limbs, descending-significance ordered
		val := bytesToLimbs(from[srcIdx : srcIdx+elemSize*8])
		if !lt(val, m.Modulus) {
			return fmt.Errorf("value (%+v) must be less than modulus (%+v)", val, m.Modulus)
		}

		//fmt.Printf("store (%d).  norm %s\n", dst+i, limbsToInt(val).String())
		// convert to Montgomery form
		m.mulMod(m.scratchSpace[dstIdx:dstIdx+elemSize],
			val,
			m.R2,
			m.Modulus,
			m.modInv)
		//fmt.Printf("mont %s\n\n", limbsToInt(m.scratchSpace[dstIdx:dstIdx+elemSize]).String())
		dstIdx++
	}
	return nil
}

func (m *FieldContext) Load(dst []byte, from, count int) {
	elemSize := len(m.Modulus)
	var dstIdx int
	// TODO: source bounds already checked in gas table?
	for srcIdx := from; srcIdx < from+count; srcIdx++ {
		res := make([]uint64, elemSize)
		//fmt.Printf("load (%d).  mont %s\n", srcIdx, limbsToInt(m.scratchSpace[srcIdx:srcIdx+elemSize]).String())
		// convert from Montgomery to canonical form
		m.mulMod(res, m.scratchSpace[srcIdx*elemSize:(srcIdx+1)*elemSize], m.one, m.Modulus, m.modInv)
		//fmt.Printf("load (%d).  norm %s\n", srcIdx, limbsToInt(res).String())
		// swap each limb to big endian (the result in dst is a big-endian number)
		for i := 0; i < elemSize; i++ {
			binary.BigEndian.PutUint64(dst[dstIdx+i*8:dstIdx+(i+1)*8], res[len(res)-(i+1)])
		}
		dstIdx += elemSize * 8
	}
}
