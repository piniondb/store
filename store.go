/*
 * Copyright (c) 2016 Kurt Jung (Gmail: piniondb)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"time"
)

var errNonempty = errors.New("the get buffer has not been completely emptied")

// KeyUint64 returns a comparable eight byte slice representation of val
// suitable for use in keys.
func KeyUint64(val uint64) (sl []byte) {
	sl = make([]byte, 8)
	binary.BigEndian.PutUint64(sl, val)
	return
}

// KeyInt64 returns a comparable eight byte slice representation of val
// suitable for use in keys.
func KeyInt64(val int64) (sl []byte) {
	return KeyUint64(uint64(val) + 1<<63)
}

// KeyUint32 returns a comparable four byte slice representation of val
// suitable for use in keys.
func KeyUint32(val uint32) (sl []byte) {
	sl = make([]byte, 4)
	binary.BigEndian.PutUint32(sl, val)
	return
}

// KeyInt32 returns a comparable four byte slice representation of val suitable
// for use in keys.
func KeyInt32(val int32) (sl []byte) {
	return KeyUint32(uint32(val) + 1<<31)
}

// KeyUint16 returns a comparable two byte slice representation of val suitable
// for use in keys.
func KeyUint16(val uint16) (sl []byte) {
	sl = make([]byte, 2)
	binary.BigEndian.PutUint16(sl, val)
	return
}

// KeyInt16 returns a comparable two byte slice representation of val suitable
// for use in keys.
func KeyInt16(val int16) (sl []byte) {
	return KeyUint16(uint16(val) + 1<<15)
}

// KeyInt8 returns a comparable 1 byte representation of val suitable
// for use in keys.
func KeyInt8(val int8) byte {
	return uint8(val) + 1<<7
}

// KeyBuffer facilitates the storage of one or more fields to be used in
// comparable, fixed-length index keys.
type KeyBuffer struct {
	buf bytes.Buffer
	err error
}

// Reset prepares the instance to be used as if new. Specifically, it clears
// the internal error code and removes any buffered content.
func (kb *KeyBuffer) Reset() {
	kb.buf.Reset()
	kb.err = nil
}

func (kb *KeyBuffer) write(sl []byte) {
	if kb.err == nil {
		_, kb.err = kb.buf.Write(sl)
	}
}

// Time stores the specified time.Time value into the receiving key
// buffer.
func (kb *KeyBuffer) Time(tm time.Time) {
	kb.write(KeyInt64(tm.Unix()))
}

// Uint64 stores the specified uint64 value into the receiving key
// buffer.
func (kb *KeyBuffer) Uint64(val uint64) {
	kb.write(KeyUint64(val))
}

// Int64 stores the specified int64 value into the receiving key buffer.
func (kb *KeyBuffer) Int64(val int64) {
	kb.write(KeyInt64(val))
}

// Uint32 stores the specified uint32 value into the receiving key
// buffer.
func (kb *KeyBuffer) Uint32(val uint32) {
	kb.write(KeyUint32(val))
}

// Int32 stores the specified int32 value into the receiving key
// buffer.
func (kb *KeyBuffer) Int32(val int32) {
	kb.write(KeyInt32(val))
}

// Uint16 stores the specifed uint16 value into the receiving key
// buffer.
func (kb *KeyBuffer) Uint16(val uint16) {
	kb.write(KeyUint16(val))
}

// Int16 stores the specifed int16 value into the receiving key
// buffer.
func (kb *KeyBuffer) Int16(val int16) {
	kb.write(KeyInt16(val))
}

// Uint8 stores the specified uint8 value into the receiving key buffer.
func (kb *KeyBuffer) Uint8(val uint8) {
	if kb.err == nil {
		kb.err = kb.buf.WriteByte(val)
	}
}

// Int8 stores the specified int8 value into the receiving key buffer.
func (kb *KeyBuffer) Int8(val int8) {
	kb.Uint8(KeyInt8(val))
}

// Str stores the specifed string value into the receiving key buffer.
// It will be either truncated or space-filled to the length specified by
// width.
func (kb *KeyBuffer) Str(str string, width uint) {
	// Consider case insensitivity
	if kb.err == nil {
		wd := int(width)
		ln := len(str)
		if ln >= int(wd) {
			_, kb.err = kb.buf.WriteString(str[:wd])
		} else {
			_, kb.err = kb.buf.WriteString(str)
			if kb.err == nil {
				// Following could be optimized by maintaining space buffer
				_, kb.err = kb.buf.WriteString(strings.Repeat(" ", wd-ln))
			}
		}
	}
}

// SetError permits the caller to assign an error value to the key buffer. In
// some cases, this may simplify the construction of a key by deferring the
// handling of an error to the point at which Bytes() is called. This method
// unconditionally overwrites the current internal error value.
func (kb *KeyBuffer) SetError(err error) {
	kb.err = err
}

// Bytes returns the slice of bytes corresponding to the stored values in the
// receiving key buffer. This is followed by the internal error code which will
// be nil if each key field has been properly loaded.
func (kb *KeyBuffer) Bytes() ([]byte, error) {
	if kb.err == nil {
		return kb.buf.Bytes(), nil
	}
	return nil, kb.err
}

func (put *PutBuffer) vluEncode(val uint64) {
	if put.err == nil {
		var hold [binary.MaxVarintLen64]byte // Holds enough septets to contain a uint64
		len := binary.PutUvarint(hold[:], val)
		_, put.err = put.buf.Write(hold[0:len])
	}
}

func vluDecode(buf *bytes.Buffer) (val uint64, err error) {
	val, err = binary.ReadUvarint(buf)
	return
}

func (put *PutBuffer) vlsEncode(val int64) {
	if put.err == nil {
		var hold [binary.MaxVarintLen64]byte // Holds enough septets to contain an int64
		len := binary.PutVarint(hold[:], val)
		_, put.err = put.buf.Write(hold[0:len])
	}
}

func vlsDecode(buf *bytes.Buffer) (val int64, err error) {
	val, err = binary.ReadVarint(buf)
	return
}

// func vluTest() {
// 	var val uint64
// 	var j int
// 	var err error
// 	var buf bytes.Buffer
// 	for j = 0; j < 12 && err == nil; j++ {
// 		val = val*31 + uint64(j)
// 		fmt.Printf("In:  %d\n", val)
// 		err = vluEncode(&buf, val)
// 		// fmt.Printf("%v\n", buf.Bytes())
// 	}
// 	for err == nil {
// 		val, err = vluDecode(&buf)
// 		if err == nil {
// 			fmt.Printf("Out: %d\n", val)
// 		}
// 	}
// 	if err != nil && err != io.EOF {
// 		fmt.Println(err)
// 	}
// }

// PutBuffer facilitates the packing of structures so that they can implement
// the encoding.BinaryMarshaler interface.
type PutBuffer struct {
	buf bytes.Buffer
	err error
}

// GetBuffer facilitates the unpacking of structures so that they can implement
// the encoding.BinaryUnmarshaler interface.
type GetBuffer struct {
	buf bytes.Buffer
	err error
}

// NewPutBuffer returns an initialized buffer that can be used to construct a
// byte slice representation of a variety of values.
func NewPutBuffer() (put *PutBuffer) {
	put = new(PutBuffer)
	return
}

// NewGetBuffer returns an initialized buffer that can be used to extract
// values from data. data specifies a byte slice that was generated using a
// PutBuffer.
func NewGetBuffer(data []byte) (get *GetBuffer) {
	get = new(GetBuffer)
	_, get.err = get.buf.Write(data)
	return
}

// Time packs the specified time.Time value into the receiving storage
// buffer.
func (put *PutBuffer) Time(tm time.Time) {
	put.vlsEncode(tm.Unix())
}

// Time unpacks a time.Time value from the receiving storage buffer.
func (get *GetBuffer) Time(tm *time.Time) {
	var val int64
	if get.err == nil {
		val, get.err = vlsDecode(&get.buf)
		if get.err == nil {
			*tm = time.Unix(val, 0)
		}
	}
}

// // Int packs the specified int value into the receiving storage buffer.
// func (put *PutBuffer) Int(val int) {
// 	put.vlsEncode(int64(val))
// }

// // Int unpacks an int value from the receiving storage buffer.
// func (b *GetBuffer) Int(val *int) {
// 	if b.err == nil {
// 		var u int64
// 		u, b.err = vlsDecode(&b.buf)
// 		if b.err == nil {
// 			*val = int(u)
// 		}
// 	}
// }

// // Uint packs the specified uint value into the receiving storage buffer.
// func (put *PutBuffer) Uint(val uint) {
// 	put.vluEncode(uint64(val))
// }

// // Uint unpacks a uint value from the receiving storage buffer.
// func (b *GetBuffer) Uint(val *uint) {
// 	if b.err == nil {
// 		var u uint64
// 		u, b.err = vluDecode(&b.buf)
// 		if b.err == nil {
// 			*val = uint(u)
// 		}
// 	}
// }

// Uint64 packs the specified uint64 value into the receiving storage
// buffer.
func (put *PutBuffer) Uint64(val uint64) {
	put.vluEncode(val)
}

// Uint64 unpacks a uint64 value from the receiving storage buffer.
func (get *GetBuffer) Uint64(val *uint64) {
	if get.err == nil {
		*val, get.err = vluDecode(&get.buf)
	}
}

// Int64 packs the specified int64 value into the receiving storage buffer.
func (put *PutBuffer) Int64(val int64) {
	put.vlsEncode(val)
}

// Int64 unpacks an int64 value from the receiving storage buffer.
func (get *GetBuffer) Int64(val *int64) {
	if get.err == nil {
		*val, get.err = vlsDecode(&get.buf)
	}
}

// Uint32 packs the specified uint32 value into the receiving storage
// buffer.
func (put *PutBuffer) Uint32(val uint32) {
	put.vluEncode(uint64(val))
}

// Uint32 unpacks a uint32 value from the receiving storage buffer.
func (get *GetBuffer) Uint32(val *uint32) {
	if get.err == nil {
		var u uint64
		u, get.err = vluDecode(&get.buf)
		if get.err == nil {
			*val = uint32(u)
		}
	}
}

// Int32 packs the specified int32 value into the receiving storage
// buffer.
func (put *PutBuffer) Int32(val int32) {
	put.vlsEncode(int64(val))
}

// Int32 unpacks an int32 value from the receiving storage buffer.
func (get *GetBuffer) Int32(val *int32) {
	if get.err == nil {
		var s int64
		s, get.err = vlsDecode(&get.buf)
		if get.err == nil {
			*val = int32(s)
		}
	}
}

// Uint16 packs the specifed uint16 value into the receiving storage
// buffer.
func (put *PutBuffer) Uint16(val uint16) {
	put.vluEncode(uint64(val))
}

// Uint16 unpacks a uint16 value from the receiving storage buffer.
func (get *GetBuffer) Uint16(val *uint16) {
	if get.err == nil {
		var u uint64
		u, get.err = vluDecode(&get.buf)
		if get.err == nil {
			*val = uint16(u)
		}
	}
}

// Int16 packs the specifed int16 value into the receiving storage
// buffer.
func (put *PutBuffer) Int16(val int16) {
	put.vlsEncode(int64(val))
}

// Int16 unpacks an int16 value from the receiving storage buffer.
func (get *GetBuffer) Int16(val *int16) {
	if get.err == nil {
		var s int64
		s, get.err = vlsDecode(&get.buf)
		if get.err == nil {
			*val = int16(s)
		}
	}
}

// Uint8 packs the specified uint8 value into the receiving storage buffer.
func (put *PutBuffer) Uint8(val uint8) {
	if put.err == nil {
		put.err = put.buf.WriteByte(val)
	}
}

// Uint8 unpacks a uint8 value from the receiving storage buffer.
func (get *GetBuffer) Uint8(val *uint8) {
	if get.err == nil {
		*val, get.err = get.buf.ReadByte()
	}
}

// Int8 packs the specified int8 value into the receiving storage buffer.
func (put *PutBuffer) Int8(val int8) {
	if put.err == nil {
		put.err = put.buf.WriteByte(uint8(val))
	}
}

// Int8 unpacks an int8 value from the receiving storage buffer.
func (get *GetBuffer) Int8(val *int8) {
	if get.err == nil {
		var b uint8
		b, get.err = get.buf.ReadByte()
		if get.err == nil {
			*val = int8(b)
		}
	}
}

// Str packs the specifed string value into the receiving storage
// buffer.
func (put *PutBuffer) Str(str string) {
	put.vluEncode(uint64(len(str)))
	if put.err == nil {
		_, put.err = put.buf.Write([]byte(str[:]))
	}
}

// Str unpacks a string value from the receiving storage buffer.
func (get *GetBuffer) Str(str *string) {
	if get.err == nil {
		var u uint64
		u, get.err = vluDecode(&get.buf)
		if get.err == nil {
			sl := make([]byte, u)
			_, get.err = get.buf.Read(sl)
			if get.err == nil {
				*str = string(sl)
			}
		}
	}
}

// SetError permits the caller to assign an error value to the put buffer. In
// some cases, this may simplify record packing by deferring the handling of an
// error to the point at which Bytes() is called. This method unconditionally
// overwrites the current internal error value.
func (put *PutBuffer) SetError(err error) {
	put.err = err
}

// Error returns the current value for the packing or unpacking operation. This
// value may be nil, in which case no error has occurred.
func (put PutBuffer) Error() error {
	return put.err
}

// SetError permits the caller to assign an error value to the get buffer. In
// some cases, this may simplify record unpacking by deferring the handling of
// an error to the point at which Done() is called. This method
// unconditionally overwrites the current internal error value.
func (get *GetBuffer) SetError(err error) {
	get.err = err
}

// Done is called to indicate that all get operations have been performed. If no error has occurred and no content remains buffered, nil is returned, otherwise an appropriate error value.
func (get GetBuffer) Done() error {
	if get.err == nil {
		if get.buf.Len() > 0 {
			get.err = errNonempty
		}
	}
	return get.err
}

// Error returns the current value for the packing or unpacking operation. This
// value may be nil, in which case no error has occurred.
func (get GetBuffer) Error() error {
	return get.err
}

// Bytes returns the currently packed fields in the form of a byte slice. The
// second return value is an error code that will be nil if all fields have
// been successfully packed.
func (put *PutBuffer) Bytes() ([]byte, error) {
	if put.err == nil {
		return put.buf.Bytes(), nil
	}
	return nil, put.err
}
