/*!
 * <字符串>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"unsafe"
)

type stringPacker struct {
}

func (this *stringPacker) UnPack(ptr unsafe.Pointer, stream *Packet) {
	*((*string)(ptr)) = stream.ReadString()
}

func (this *stringPacker) Pack(ptr unsafe.Pointer, stream *Packet) {
	stream.WriteString(*((*string)(ptr)))
}

func (this *stringPacker) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*string)(ptr)) == ""
}
