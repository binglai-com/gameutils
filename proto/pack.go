/*!
 * <结构体打包/解包>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"errors"

	"github.com/modern-go/reflect2"
)

//打包
func Pack(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte{}, errors.New("value is nil")
	}

	typ := reflect2.TypeOf(v)

	stream := new(Packet)
	packer := packerOfType(typ)
	packer.Pack(reflect2.PtrOf(v), stream)

	return stream.GetBuffer(), nil
}

//解包
func UnPack(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("value is nil")
	}

	stream := new(Packet)
	stream.Init(data)

	typ := reflect2.TypeOf(v)
	unpacker := unpackOfType(typ)
	unpacker.UnPack(reflect2.PtrOf(v), stream)

	return nil
}
