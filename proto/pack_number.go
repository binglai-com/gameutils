/*!
 * <数字>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"math"
	"unsafe"
)

type int8Packer struct {
}

func (this *int8Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*int8)(ptr)) = stream.ReadInt8()
}

func (this *int8Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*int8)(ptr)))
}

func (this *int8Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*int8)(ptr)) == 0
}

type int16Packer struct {
}

func (this *int16Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*int16)(ptr)) = stream.ReadInt16()
}

func (this *int16Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*int16)(ptr)))
}

func (this *int16Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*int16)(ptr)) == 0
}

type int32Packer struct {
}

func (this *int32Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*int32)(ptr)) = stream.ReadInt32()
}

func (this *int32Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*int32)(ptr)))
}

func (this *int32Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*int32)(ptr)) == 0
}

type int64Packer struct {
}

func (this *int64Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*int64)(ptr)) = stream.ReadInt64()
}

func (this *int64Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*int64)(ptr)))
}

func (this *int64Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*int64)(ptr)) == 0
}

type uint8Packer struct {
}

func (this *uint8Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*uint8)(ptr)) = stream.ReadUint8()
}

func (this *uint8Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*uint8)(ptr)))
}

func (this *uint8Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*uint8)(ptr)) == 0
}

type uint16Packer struct {
}

func (this *uint16Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*uint16)(ptr)) = stream.ReadUint16()
}

func (this *uint16Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*uint16)(ptr)))
}

func (this *uint16Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*uint16)(ptr)) == 0
}

type uint32Packer struct {
}

func (this *uint32Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*uint32)(ptr)) = stream.ReadUint32()
}

func (this *uint32Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*uint32)(ptr)))
}

func (this *uint32Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*uint32)(ptr)) == 0
}

type uint64Packer struct {
}

func (this *uint64Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*uint64)(ptr)) = stream.ReadUint64()
}

func (this *uint64Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(*((*uint64)(ptr)))
}

func (this *uint64Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*uint64)(ptr)) == 0
}

type float32Packer struct {
}

func (this *float32Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*float32)(ptr)) = math.Float32frombits(stream.ReadUint32())
}

func (this *float32Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(math.Float32bits(*((*float32)(ptr))))
}

func (this *float32Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*float32)(ptr)) == 0
}

type float64Packer struct {
}

func (this *float64Packer) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*float64)(ptr)) = math.Float64frombits(stream.ReadUint64())
}

func (this *float64Packer) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteInt(math.Float64bits(*((*float64)(ptr))))
}

func (this *float64Packer) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*float64)(ptr)) == 0
}
