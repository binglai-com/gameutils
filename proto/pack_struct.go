/*!
 * <结构体>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"unsafe"

	"github.com/modern-go/reflect2"
)

type structPacker struct {
	typ    reflect2.Type
	fields []*structFieldPacker
}

type structFieldPacker struct {
	field       reflect2.StructField
	fieldPacker ValPack
}

type FieldInfo struct {
	Field    reflect2.StructField
	Packer   ValPack
	UnPacker ValUnPack
}

func (packer *structFieldPacker) Pack(ptr unsafe.Pointer, stream *Packet) {
	fieldPtr := packer.field.UnsafeGet(ptr)
	packer.fieldPacker.Pack(fieldPtr, stream)
}

func (packer *structFieldPacker) IsEmpty(ptr unsafe.Pointer) bool {
	fieldPtr := packer.field.UnsafeGet(ptr)
	return packer.fieldPacker.IsEmpty(fieldPtr)
}

func (packer *structPacker) Pack(ptr unsafe.Pointer, stream *Packet) {
	for _, field := range packer.fields {
		field.Pack(ptr, stream)
	}
}

func (packer *structPacker) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

func packerOfStruct(typ reflect2.Type) ValPack {
	allFields := describeStruct(typ)
	finalFields := []*structFieldPacker{}
	for _, field := range allFields {
		finalFields = append(finalFields, field.Packer.(*structFieldPacker))
	}

	return &structPacker{typ, finalFields}
}

func describeStruct(typ reflect2.Type) []*FieldInfo {
	structType := typ.(*reflect2.UnsafeStructType)
	allfields := []*FieldInfo{}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		unpacker := unpackOfType(field.Type())
		if unpacker == nil {
			return allfields
		}

		packer := packerOfType(field.Type())
		if packer == nil {
			return allfields
		}

		fields := &FieldInfo{
			Field:    field,
			Packer:   &structFieldPacker{field, packer},
			UnPacker: &structFieldUnPack{field, unpacker},
		}
		allfields = append(allfields, fields)
	}

	return allfields
}

////////////////////////////////////////////////////////////////////////////////
func unpackOfStruct(typ reflect2.Type) ValUnPack {
	allFields := describeStruct(typ)
	finalFields := []*structFieldUnPack{}
	for _, field := range allFields {
		finalFields = append(finalFields, field.UnPacker.(*structFieldUnPack))
	}

	return &StructUnPacker{typ, finalFields}
}

type StructUnPacker struct {
	typ    reflect2.Type
	fields []*structFieldUnPack
}

func (this *StructUnPacker) UnPack(ptr unsafe.Pointer, iter *Packet) {
	for _, v := range this.fields {
		v.UnPack(ptr, iter)
	}
}

type structFieldUnPack struct {
	field    reflect2.StructField
	unpacker ValUnPack
}

func (this *structFieldUnPack) UnPack(ptr unsafe.Pointer, iter *Packet) {
	fieldPtr := this.field.UnsafeGet(ptr)
	this.unpacker.UnPack(fieldPtr, iter)
}
