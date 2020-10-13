/*!
 * <字节流>
 *
 * Copyright (c) 2018 by <yfm/ BingLai Co.>
 */

package proto

import (
	"bytes"
	"encoding/binary"
	"log"
)

const (
	PACKET_LOG = "package_error"
)

var BYTE_ORDER = binary.BigEndian

//4字节数字转字节
func IntToByte(v interface{}) ([]byte, error) {
	header := bytes.NewBuffer([]byte{})
	err := binary.Write(header, BYTE_ORDER, v)
	if err != nil {
		return []byte{}, err
	}
	return header.Bytes(), nil
}

//uint16类型数字
func Uint16(d []byte) uint16 {
	var v uint16 = 0
	header := bytes.NewBuffer(d)
	binary.Read(header, BYTE_ORDER, &v)
	return v
}

//uint32类型数字
func Uint32(d []byte) uint32 {
	var v uint32 = 0
	header := bytes.NewBuffer(d)
	binary.Read(header, BYTE_ORDER, &v)
	return v
}

//int32类型数字
func Int32(d []byte) int32 {
	var v int32 = 0
	header := bytes.NewBuffer(d)
	binary.Read(header, BYTE_ORDER, &v)
	return v
}

type Packet struct {
	buffer  []byte
	readpos int
}

func (this *Packet) Init(data []byte) {
	this.Reset()
	this.buffer = append(this.buffer, data...)
}

func (this *Packet) Reset() {
	this.buffer = make([]byte, 0)
	this.readpos = 0
}

func (this *Packet) GetBuffer() []byte {
	return this.buffer
}

func (this *Packet) WriteInt(v interface{}) {
	by, err := IntToByte(v)
	if err != nil {
		log.Println("WriteInt error :", err.Error())
	}

	this.buffer = append(this.buffer, by...)
}

func (this *Packet) ReadUint8() uint8 {
	var value uint8 = 0
	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+1])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("ReadUint8 error :", err.Error())
	}
	this.readpos += 1

	return value
}

func (this *Packet) ReadInt8() int8 {
	var value int8 = 0
	if this.readpos+1 > len(this.buffer) {
		return value
	}

	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+1])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("Readint8 error :", err.Error())
	}
	this.readpos += 1

	return value
}

func (this *Packet) ReadUint16() uint16 {
	var value uint16 = 0
	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+2])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("ReadUint16 error :", err.Error())
	}
	this.readpos += 2

	return value
}

func (this *Packet) ReadInt16() int16 {
	var value int16 = 0
	if this.readpos+2 > len(this.buffer) {
		return value
	}

	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+2])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("Readint16 error :", err.Error())
	}
	this.readpos += 2

	return value
}

func (this *Packet) ReadUint32() uint32 {
	var value uint32 = 0
	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+4])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("ReadUint32 error :", err.Error())
	}
	this.readpos += 4

	return value
}

func (this *Packet) ReadInt32() int32 {
	var value int32 = 0
	if this.readpos+4 > len(this.buffer) {
		return value
	}

	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+4])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("ReadUint32 error :", err.Error())
	}
	this.readpos += 4

	return value
}

func (this *Packet) ReadUint64() uint64 {
	var value uint64 = 0
	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+8])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("ReadUint64 error :", err.Error())
	}
	this.readpos += 8

	return value
}

func (this *Packet) ReadInt64() int64 {
	var value int64 = 0
	if this.readpos+8 > len(this.buffer) {
		return value
	}

	header := bytes.NewBuffer(this.buffer[this.readpos : this.readpos+8])
	err := binary.Read(header, BYTE_ORDER, &value)
	if err != nil {
		log.Println("ReadInt64 error :", err.Error())
	}
	this.readpos += 8

	return value
}

func (this *Packet) WriteString(str string) {
	strby := []byte(str)
	strlen := len(strby)
	this.WriteInt(uint16(strlen))
	this.buffer = append(this.buffer, strby...)
}

func (this *Packet) ReadString() string {
	strlen := this.ReadUint16()
	strbyte := this.buffer[this.readpos : this.readpos+int(strlen)]
	this.readpos += len(strbyte)

	return string(strbyte)
}
