/*!
 * <指针>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"unsafe"

	"github.com/modern-go/reflect2"
)

func packerOfPtr(typ reflect2.Type) ValPack {
	ptrType := typ.(*reflect2.UnsafePtrType)
	elemType := ptrType.Elem()
	elemPacker := packerOfType(elemType)
	pak := &ptrPacker{elemPacker}
	return pak
}

func unpackerOfPtr(typ reflect2.Type) ValUnPack {
	ptrType := typ.(*reflect2.UnsafePtrType)
	elemType := ptrType.Elem()
	elemPacker := unpackOfType(elemType)
	unpak := &ptrUnPacker{elemPacker}
	return unpak
}

type ptrPacker struct {
	ValuePacker ValPack
}

func (this *ptrPacker) Pack(ptr unsafe.Pointer, stream *Packet) {
	newptr := unsafe.Pointer(&ptr)
	if *((*unsafe.Pointer)(newptr)) == nil {
		return
	}

	this.ValuePacker.Pack(*((*unsafe.Pointer)(newptr)), stream)
}

func (this *ptrPacker) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*unsafe.Pointer)(ptr)) == nil
}

type ptrUnPacker struct {
	unpack ValUnPack
}

func (this *ptrUnPacker) UnPack(ptr unsafe.Pointer, stream *Packet) {
	newptr := unsafe.Pointer(&ptr)
	if *((*unsafe.Pointer)(newptr)) == nil {
		return
	}

	this.unpack.UnPack(*((*unsafe.Pointer)(newptr)), stream)
}
