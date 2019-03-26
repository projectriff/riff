// Protocol Buffers for Go with Gadgets
//
// Copyright (c) 2013, The GoGo Authors. All rights reserved.
// https://github.com/gogo/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

/*
	Package custom contains custom types for test and example purposes.
	These types are used by the test structures generated by gogoprotobuf.
*/
package custom

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Uint128 [2]uint64

func (u Uint128) Marshal() ([]byte, error) {
	buffer := make([]byte, 16)
	_, err := u.MarshalTo(buffer)
	return buffer, err
}

func (u Uint128) MarshalTo(data []byte) (n int, err error) {
	PutLittleEndianUint128(data, 0, u)
	return 16, nil
}

func GetLittleEndianUint64(b []byte, offset int) uint64 {
	v := uint64(b[offset+7]) << 56
	v += uint64(b[offset+6]) << 48
	v += uint64(b[offset+5]) << 40
	v += uint64(b[offset+4]) << 32
	v += uint64(b[offset+3]) << 24
	v += uint64(b[offset+2]) << 16
	v += uint64(b[offset+1]) << 8
	v += uint64(b[offset])
	return v
}

func PutLittleEndianUint64(b []byte, offset int, v uint64) {
	b[offset] = byte(v)
	b[offset+1] = byte(v >> 8)
	b[offset+2] = byte(v >> 16)
	b[offset+3] = byte(v >> 24)
	b[offset+4] = byte(v >> 32)
	b[offset+5] = byte(v >> 40)
	b[offset+6] = byte(v >> 48)
	b[offset+7] = byte(v >> 56)
}

func PutLittleEndianUint128(buffer []byte, offset int, v [2]uint64) {
	PutLittleEndianUint64(buffer, offset, v[0])
	PutLittleEndianUint64(buffer, offset+8, v[1])
}

func GetLittleEndianUint128(buffer []byte, offset int) (value [2]uint64) {
	value[0] = GetLittleEndianUint64(buffer, offset)
	value[1] = GetLittleEndianUint64(buffer, offset+8)
	return
}

func (u *Uint128) Unmarshal(data []byte) error {
	if data == nil {
		u = nil
		return nil
	}
	if len(data) == 0 {
		pu := Uint128{}
		*u = pu
		return nil
	}
	if len(data) != 16 {
		return errors.New("Uint128: invalid length")
	}
	pu := Uint128(GetLittleEndianUint128(data, 0))
	*u = pu
	return nil
}

func (u Uint128) MarshalJSON() ([]byte, error) {
	data, err := u.Marshal()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (u Uint128) Size() int {
	return 16
}

func (u *Uint128) UnmarshalJSON(data []byte) error {
	v := new([]byte)
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return u.Unmarshal(*v)
}

func (this Uint128) Equal(that Uint128) bool {
	return this == that
}

func (this Uint128) Compare(that Uint128) int {
	thisdata, err := this.Marshal()
	if err != nil {
		panic(err)
	}
	thatdata, err := that.Marshal()
	if err != nil {
		panic(err)
	}
	return bytes.Compare(thisdata, thatdata)
}

type randy interface {
	Intn(n int) int
}

func NewPopulatedUint128(r randy) *Uint128 {
	data := make([]byte, 16)
	for i := 0; i < 16; i++ {
		data[i] = byte(r.Intn(255))
	}
	u := Uint128(GetLittleEndianUint128(data, 0))
	return &u
}
