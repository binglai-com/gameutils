/*!
 * <结构体打包[]byte>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"log"
	"reflect"
	"unsafe"

	"github.com/modern-go/reflect2"
)

type ValUnPack interface {
	UnPack(ptr unsafe.Pointer, stream *Packet)
}

type ValPack interface {
	Pack(ptr unsafe.Pointer, stream *Packet)
	IsEmpty(ptr unsafe.Pointer) bool
}

func packerOfType(typ reflect2.Type) ValPack {
	switch typ.Kind() {
	case reflect.String:
		return &stringPacker{}
	case reflect.Int8:
		return &int8Packer{}
	case reflect.Int16:
		return &int16Packer{}
	case reflect.Int32:
		return &int32Packer{}
	case reflect.Int64:
		return &int64Packer{}
	case reflect.Uint8:
		return &uint8Packer{}
	case reflect.Uint16:
		return &uint16Packer{}
	case reflect.Uint32:
		return &uint32Packer{}
	case reflect.Uint64:
		return &uint64Packer{}
	case reflect.Float32:
		return &float32Packer{}
	case reflect.Float64:
		return &float64Packer{}
	case reflect.Slice:
		return packerOfSlice(typ)
	case reflect.Struct:
		return packerOfStruct(typ)
	case reflect.Ptr:
		return packerOfPtr(typ)
	default:
		log.Println("unsurport type :", typ.Kind())
	}

	return nil
}

func unpackOfType(typ reflect2.Type) ValUnPack {
	switch typ.Kind() {
	case reflect.String:
		return &stringPacker{}
	case reflect.Int8:
		return &int8Packer{}
	case reflect.Int16:
		return &int16Packer{}
	case reflect.Int32:
		return &int32Packer{}
	case reflect.Int64:
		return &int64Packer{}
	case reflect.Uint8:
		return &uint8Packer{}
	case reflect.Uint16:
		return &uint16Packer{}
	case reflect.Uint32:
		return &uint32Packer{}
	case reflect.Uint64:
		return &uint64Packer{}
	case reflect.Float32:
		return &float32Packer{}
	case reflect.Float64:
		return &float64Packer{}
	case reflect.Slice:
		return unpackerOfSlice(typ)
	case reflect.Struct:
		return unpackOfStruct(typ)
	case reflect.Ptr:
		return unpackerOfPtr(typ)
	default:
		log.Println("unsurport type :", typ.Kind())
	}

	return nil
}
