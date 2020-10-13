/*!
 * <切片>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"unsafe"

	"github.com/modern-go/reflect2"
)

func packerOfSlice(typ reflect2.Type) ValPack {
	sliceType := typ.(*reflect2.UnsafeSliceType)
	packer := packerOfType(sliceType.Elem())
	return &slicePacker{sliceType, packer}
}

func unpackerOfSlice(typ reflect2.Type) ValUnPack {
	sliceType := typ.(*reflect2.UnsafeSliceType)
	unpacker := unpackOfType(sliceType.Elem())
	return &sliceUnPacker{sliceType, unpacker}
}

type slicePacker struct {
	sliceType  *reflect2.UnsafeSliceType
	elemPacker ValPack
}

func (this *slicePacker) Pack(ptr unsafe.Pointer, stream *Packet) {
	if this.sliceType.UnsafeIsNil(ptr) {
		stream.WriteInt(uint16(0))
		return
	}

	length := uint16(this.sliceType.UnsafeLengthOf(ptr))
	stream.WriteInt(length)
	if length == 0 {
		return
	}

	this.elemPacker.Pack(this.sliceType.UnsafeGetIndex(ptr, 0), stream)
	for i := uint16(1); i < length; i++ {
		elemPtr := this.sliceType.UnsafeGetIndex(ptr, int(i))
		this.elemPacker.Pack(elemPtr, stream)
	}
}

func (this *slicePacker) IsEmpty(ptr unsafe.Pointer) bool {
	return this.sliceType.UnsafeLengthOf(ptr) == 0
}

type sliceUnPacker struct {
	sliceType *reflect2.UnsafeSliceType
	unpacker  ValUnPack
}

func (this *sliceUnPacker) UnPack(ptr unsafe.Pointer, stream *Packet) {
	sliceType := this.sliceType
	arrlen := stream.ReadUint16()
	if arrlen == 0 {
		return
	}

	length := 1
	sliceType.UnsafeGrow(ptr, 1)
	elemPtr := sliceType.UnsafeGetIndex(ptr, 0)
	this.unpacker.UnPack(elemPtr, stream)

	for i := uint16(1); i < arrlen; i++ {
		idx := length
		length += 1
		sliceType.UnsafeGrow(ptr, length)
		elemPtr = sliceType.UnsafeGetIndex(ptr, idx)
		this.unpacker.UnPack(elemPtr, stream)
	}
}
