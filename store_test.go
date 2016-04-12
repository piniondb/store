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

package store_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/piniondb/store"
)

var (
	errTest  = errors.New("test error")
	timeTest = time.Date(1997, time.November, 28, 12, 0, 0, 0, time.UTC)
)

type sub struct {
	U64 uint64
	S8  int8
}

// Type all contains a variety of field types for testing purposes. The store
// package does not require fields to be exported; it is done here so this type
// can be used for both store and JSON benchmarking.
type all struct {
	U64 uint64
	S64 int64
	U32 uint32
	S32 int32
	U16 uint16
	S16 int16
	U8  uint8
	S8  int8
	S   string
	T   time.Time
	Sl  []sub
	B   []byte
	Mp  map[string]string
}

// recPopulate assigns all fields of the specified record with arbitrary
// values.
func recPopulate(r *all) {
	r.U64 = 3565123234760
	r.S64 = -50496192383
	r.U32 = 5470129
	r.S32 = -50129
	r.U16 = 45092
	r.S16 = -30901
	r.U8 = 212
	r.S8 = -34
	r.S = "example"
	r.T = timeTest
	r.Sl = make([]sub, 3)
	r.Sl[0].U64 = 123
	r.Sl[0].S8 = 2
	r.Sl[1].U64 = 345
	r.Sl[1].S8 = 5
	r.Sl[2].U64 = 567
	r.Sl[2].S8 = -8
	r.B = make([]byte, 3)
	r.B[0] = 42
	r.B[1] = 41
	r.B[2] = 40
	r.Mp = make(map[string]string)
	r.Mp["key1"] = "value1"
	r.Mp["key2"] = "value2"
	r.Mp["key3"] = "value3"
}

// String implements the fmt.Stringer interface. Map keys are sorted so that
// different structures with the same field values will have identical strings.
// This is useful for guaranteeing that a decoded structure is equivalent to
// its original encoded structure.
func (r all) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%d;%d;%d;%d;%d;%d;i%d;%d;%s;%s;",
		r.U64, r.S64, r.U32, r.S32, r.U16, r.S16, r.U8, r.S8, r.S, r.T.UTC())
	for _, sub := range r.Sl {
		fmt.Fprintf(&b, "%d;%d;", sub.U64, sub.S8)
	}
	fmt.Fprintf(&b, "%v;", r.B)
	var keyList []string
	for k := range r.Mp {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)
	for _, k := range keyList {
		fmt.Fprintf(&b, "%s:%s;", k, r.Mp[k])
	}
	return b.String()
}

// storeRecToBuf packs all record fields into a byte slice using a put buffer
// from the store package.
func storeRecToBuf(rec all) ([]byte, error) {
	var put store.PutBuffer
	// Pack structure into buffer
	put.Uint64(rec.U64)
	put.Int64(rec.S64)
	put.Uint32(rec.U32)
	put.Int32(rec.S32)
	put.Uint16(rec.U16)
	put.Int16(rec.S16)
	put.Uint8(rec.U8)
	put.Int8(rec.S8)
	put.Str(rec.S)
	put.Time(rec.T)
	put.Uint16(uint16(len(rec.Sl)))
	for _, sub := range rec.Sl {
		put.Uint64(sub.U64)
		put.Int8(sub.S8)
	}
	put.Bytes(rec.B)
	put.Uint16(uint16(len(rec.Mp)))
	for k, v := range rec.Mp {
		put.Str(k)
		put.Str(v)
	}
	return put.Data()
}

// storeBufToRec unpacks all record fields from a byte slice using a get buffer
// from the store package.
func storeBufToRec(data []byte) (rec all, err error) {
	var slen uint16
	var keyStr, valStr string
	var get = store.NewGetBuffer(data)
	// Unpack buffer into new structure
	get.Uint64(&rec.U64)
	get.Int64(&rec.S64)
	get.Uint32(&rec.U32)
	get.Int32(&rec.S32)
	get.Uint16(&rec.U16)
	get.Int16(&rec.S16)
	get.Uint8(&rec.U8)
	get.Int8(&rec.S8)
	get.Str(&rec.S)
	get.Time(&rec.T)
	get.Uint16(&slen)
	// Retrieve length of slice
	rec.Sl = make([]sub, slen)
	for j := uint16(0); j < slen; j++ {
		get.Uint64(&rec.Sl[j].U64)
		get.Int8(&rec.Sl[j].S8)
	}
	get.Bytes(&rec.B)
	// Retrieve length of map
	get.Uint16(&slen)
	rec.Mp = make(map[string]string)
	for j := uint16(0); j < slen; j++ {
		get.Str(&keyStr)
		get.Str(&valStr)
		rec.Mp[keyStr] = valStr
	}
	err = get.Done()
	return
}

// jsonRecToBuf encodes all record fields into a byte slice using the JSON
// encoding package.
func jsonRecToBuf(rec all) ([]byte, error) {
	return json.Marshal(rec)
}

// jsonRecToBuf decodes all record fields from a byte slice using the JSON
// encoding package.
func jsonBufToRec(sl []byte) (rec all, err error) {
	err = json.Unmarshal(sl, &rec)
	return
}

// ExampleGetBuffer demonstrates basic packing and unpacking of a structure.
func ExampleGetBuffer() {
	var rec all
	var recBuf []byte
	var err error
	var put store.PutBuffer
	var originalStr, restoredStr string
	recPopulate(&rec)
	// Pack structure into buffer
	originalStr = rec.String()
	put.Uint64(rec.U64)
	put.Int64(rec.S64)
	put.Uint32(rec.U32)
	put.Int32(rec.S32)
	put.Uint16(rec.U16)
	put.Int16(rec.S16)
	put.Uint8(rec.U8)
	put.Int8(rec.S8)
	put.Str(rec.S)
	put.Time(rec.T)
	put.Uint16(uint16(len(rec.Sl)))
	for _, sub := range rec.Sl {
		put.Uint64(sub.U64)
		put.Int8(sub.S8)
	}
	put.Bytes(rec.B)
	put.Uint16(uint16(len(rec.Mp)))
	for k, v := range rec.Mp {
		put.Str(k)
		put.Str(v)
	}
	recBuf, err = put.Data()
	if err == nil {
		var newRec all
		var slen uint16
		var keyStr, valStr string
		var get = store.NewGetBuffer(recBuf)
		// Unpack buffer into new structure
		get.Uint64(&newRec.U64)
		get.Int64(&newRec.S64)
		get.Uint32(&newRec.U32)
		get.Int32(&newRec.S32)
		get.Uint16(&newRec.U16)
		get.Int16(&newRec.S16)
		get.Uint8(&newRec.U8)
		get.Int8(&newRec.S8)
		get.Str(&newRec.S)
		get.Time(&newRec.T)
		get.Uint16(&slen)
		// Retrieve length of slice
		newRec.Sl = make([]sub, slen)
		for j := uint16(0); j < slen; j++ {
			get.Uint64(&newRec.Sl[j].U64)
			get.Int8(&newRec.Sl[j].S8)
		}
		get.Bytes(&newRec.B)
		// Retrieve length of map
		get.Uint16(&slen)
		newRec.Mp = make(map[string]string)
		for j := uint16(0); j < slen; j++ {
			get.Str(&keyStr)
			get.Str(&valStr)
			newRec.Mp[keyStr] = valStr
		}
		err = get.Done()
		if err == nil {
			restoredStr = newRec.String()
			if restoredStr == originalStr {
				fmt.Printf("Original structure is the same as the restored structure\n")
			} else {
				err = fmt.Errorf("structure storage/extraction error; original: %s, restored: %s",
					originalStr, restoredStr)
			}
		}
	}
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// Original structure is the same as the restored structure
}

// Write a hexadecimal representation of the byte slice to the specified writer.
func out(w io.Writer, sl []byte) {
	slen := len(sl)
	for pos := 0; pos < slen; pos += 16 {
		wlen := slen - pos
		if wlen > 16 {
			wlen = 16
		}
		fmt.Fprintf(w, "% x\n", sl[pos:pos+wlen])
	}
}

// ExampleKeyBuffer_build demonstrates building up a key composed with various
// fixed-length comparable values.
func ExampleKeyBuffer_build() {
	var kb store.KeyBuffer
	kb.Time(timeTest)
	kb.Uint64(3565123234760)
	kb.Int64(-50496192383)
	kb.Uint32(5470129)
	kb.Int32(-50129)
	kb.Uint16(45092)
	kb.Int16(-30901)
	kb.Uint8(212)
	kb.Int8(-34)
	kb.Str("example", 4)
	kb.Str("do", 4)
	kb.Bytes([]byte{1, 2, 3, 4, 5}, 4)
	kb.Bytes([]byte{1, 2}, 4)
	sl, err := kb.Data()
	if err == nil {
		out(os.Stdout, sl)
	} else {
		fmt.Println(err)
	}
	// Output:
	// 80 00 00 00 34 7e b2 40 00 00 03 3e 11 e7 6b c8
	// 7f ff ff f4 3e 31 40 81 00 53 77 b1 7f ff 3c 2f
	// b0 24 07 4b d4 5e 65 78 61 6d 64 6f 20 20 01 02
	// 03 04 01 02 00 00
}

// type simple includes a few elementary types
type simple struct {
	a int64
	b uint32
	c int8
	d string
}

// String implements the fmt.Stringer interface for type simple
func (r simple) String() string {
	return fmt.Sprintf("%15d | %12d | %4d | %-10s", r.a, r.b, r.c, r.d)
}

// ExampleKeyBuffer_sort demonstrates the comparability of keys built up with
// KeyBuffer methods.
func ExampleKeyBuffer_sort() {
	var keyStr string
	var keyList []string
	var sl []byte
	var err error
	mp := make(map[string]simple)
	var r simple
	for _, a := range []int64{3023434, -543870, 10023494551, -3} {
		r.a = a
		for _, b := range []uint32{454567, 100232, 3450123420} {
			r.b = b
			for _, c := range []int8{-98, 0, 32} {
				r.c = c
				for _, d := range []string{"abc", "rstuvwxyz"} {
					var kb store.KeyBuffer
					r.d = d
					kb.Int64(r.a)
					kb.Uint32(r.b)
					kb.Int8(r.c)
					kb.Str(r.d, 8)
					sl, err = kb.Data()
					if err == nil {
						keyStr = string(sl)
						keyList = append(keyList, keyStr)
						mp[keyStr] = r
					}
				}
			}
		}
	}
	sort.Strings(keyList)
	for _, key := range keyList {
		fmt.Printf("[%s]\n", mp[key])
	}
	// Output:
	// [        -543870 |       100232 |  -98 | abc       ]
	// [        -543870 |       100232 |  -98 | rstuvwxyz ]
	// [        -543870 |       100232 |    0 | abc       ]
	// [        -543870 |       100232 |    0 | rstuvwxyz ]
	// [        -543870 |       100232 |   32 | abc       ]
	// [        -543870 |       100232 |   32 | rstuvwxyz ]
	// [        -543870 |       454567 |  -98 | abc       ]
	// [        -543870 |       454567 |  -98 | rstuvwxyz ]
	// [        -543870 |       454567 |    0 | abc       ]
	// [        -543870 |       454567 |    0 | rstuvwxyz ]
	// [        -543870 |       454567 |   32 | abc       ]
	// [        -543870 |       454567 |   32 | rstuvwxyz ]
	// [        -543870 |   3450123420 |  -98 | abc       ]
	// [        -543870 |   3450123420 |  -98 | rstuvwxyz ]
	// [        -543870 |   3450123420 |    0 | abc       ]
	// [        -543870 |   3450123420 |    0 | rstuvwxyz ]
	// [        -543870 |   3450123420 |   32 | abc       ]
	// [        -543870 |   3450123420 |   32 | rstuvwxyz ]
	// [             -3 |       100232 |  -98 | abc       ]
	// [             -3 |       100232 |  -98 | rstuvwxyz ]
	// [             -3 |       100232 |    0 | abc       ]
	// [             -3 |       100232 |    0 | rstuvwxyz ]
	// [             -3 |       100232 |   32 | abc       ]
	// [             -3 |       100232 |   32 | rstuvwxyz ]
	// [             -3 |       454567 |  -98 | abc       ]
	// [             -3 |       454567 |  -98 | rstuvwxyz ]
	// [             -3 |       454567 |    0 | abc       ]
	// [             -3 |       454567 |    0 | rstuvwxyz ]
	// [             -3 |       454567 |   32 | abc       ]
	// [             -3 |       454567 |   32 | rstuvwxyz ]
	// [             -3 |   3450123420 |  -98 | abc       ]
	// [             -3 |   3450123420 |  -98 | rstuvwxyz ]
	// [             -3 |   3450123420 |    0 | abc       ]
	// [             -3 |   3450123420 |    0 | rstuvwxyz ]
	// [             -3 |   3450123420 |   32 | abc       ]
	// [             -3 |   3450123420 |   32 | rstuvwxyz ]
	// [        3023434 |       100232 |  -98 | abc       ]
	// [        3023434 |       100232 |  -98 | rstuvwxyz ]
	// [        3023434 |       100232 |    0 | abc       ]
	// [        3023434 |       100232 |    0 | rstuvwxyz ]
	// [        3023434 |       100232 |   32 | abc       ]
	// [        3023434 |       100232 |   32 | rstuvwxyz ]
	// [        3023434 |       454567 |  -98 | abc       ]
	// [        3023434 |       454567 |  -98 | rstuvwxyz ]
	// [        3023434 |       454567 |    0 | abc       ]
	// [        3023434 |       454567 |    0 | rstuvwxyz ]
	// [        3023434 |       454567 |   32 | abc       ]
	// [        3023434 |       454567 |   32 | rstuvwxyz ]
	// [        3023434 |   3450123420 |  -98 | abc       ]
	// [        3023434 |   3450123420 |  -98 | rstuvwxyz ]
	// [        3023434 |   3450123420 |    0 | abc       ]
	// [        3023434 |   3450123420 |    0 | rstuvwxyz ]
	// [        3023434 |   3450123420 |   32 | abc       ]
	// [        3023434 |   3450123420 |   32 | rstuvwxyz ]
	// [    10023494551 |       100232 |  -98 | abc       ]
	// [    10023494551 |       100232 |  -98 | rstuvwxyz ]
	// [    10023494551 |       100232 |    0 | abc       ]
	// [    10023494551 |       100232 |    0 | rstuvwxyz ]
	// [    10023494551 |       100232 |   32 | abc       ]
	// [    10023494551 |       100232 |   32 | rstuvwxyz ]
	// [    10023494551 |       454567 |  -98 | abc       ]
	// [    10023494551 |       454567 |  -98 | rstuvwxyz ]
	// [    10023494551 |       454567 |    0 | abc       ]
	// [    10023494551 |       454567 |    0 | rstuvwxyz ]
	// [    10023494551 |       454567 |   32 | abc       ]
	// [    10023494551 |       454567 |   32 | rstuvwxyz ]
	// [    10023494551 |   3450123420 |  -98 | abc       ]
	// [    10023494551 |   3450123420 |  -98 | rstuvwxyz ]
	// [    10023494551 |   3450123420 |    0 | abc       ]
	// [    10023494551 |   3450123420 |    0 | rstuvwxyz ]
	// [    10023494551 |   3450123420 |   32 | abc       ]
	// [    10023494551 |   3450123420 |   32 | rstuvwxyz ]
}

// Compare JSON and store encodings
func TestPutBuffer_Compare(t *testing.T) {
	var rec all
	var err error
	var data []byte
	var buf bytes.Buffer
	recPopulate(&rec)
	str := rec.String()
	fmt.Fprintf(&buf, "Original record [%s]\n", str)
	data, err = storeRecToBuf(rec)
	if err == nil {
		storeLen := len(data)
		fmt.Fprintf(&buf, "store-encoded, len %d\n", storeLen)
		out(&buf, data)
		rec, err = storeBufToRec(data)
		if err == nil {
			if str == rec.String() {
				data, err = jsonRecToBuf(rec)
				if err == nil {
					jsonLen := len(data)
					fmt.Fprintf(&buf, "json-encoded, len %d\n", jsonLen)
					out(&buf, data)
					rec, err = jsonBufToRec(data)
					if err == nil {
						fmt.Fprintf(&buf, "JSON to Store size ratio: %.1f\n", float64(jsonLen)/float64(storeLen))
						if str == rec.String() {
							err = ioutil.WriteFile("report", buf.Bytes(), 0660)
						} else {
							err = fmt.Errorf("JSON-encoding problem")
						}
					}
				}
			} else {
				err = fmt.Errorf("store-encoding problem")
			}
		}
	}
	if err != nil {
		t.Error(err)
	}
}

// Ensure that error in put buffer loading is reported
func TestPutBuffer_Error(t *testing.T) {
	var put store.PutBuffer
	put.Int8(-2)
	put.SetError(errTest)
	sl, err := put.Data()
	if sl != nil || err == nil {
		t.Fatal("PutBuffer error not reported")
	}
	if put.Error() == nil {
		t.Fatal("PutBuffer error not reported in call to Error()")
	}
}

// Ensure that error in get buffer loading is reported
func TestGetBuffer_Error(t *testing.T) {
	get := store.NewGetBuffer([]byte{0, 0, 0})
	get.SetError(errTest)
	if get.Error() == nil {
		t.Fatal("GetBuffer error not reported in call to Error()")
	}
}

// Ensure that error leftover content in buffer is reported
func TestGetBuffer_Leftover(t *testing.T) {
	var put store.PutBuffer
	put.Uint32(5)
	put.Uint32(8)
	data, err := put.Data()
	if err == nil {
		var v uint32
		get := store.NewGetBuffer(data)
		get.Uint32(&v)
		err = get.Done()
		if err == nil {
			t.Fatal("Remaining buffered content not reported")
		}
	} else {
		t.Fatal(err)
	}
}

// Ensure that error in key buffer loading is reported
func TestKeyBuffer_Error(t *testing.T) {
	var kb store.KeyBuffer
	kb.Uint32(3)
	kb.SetError(errTest)
	sl, err := kb.Data()
	if sl != nil || err == nil {
		t.Fatal("KeyBuffer error not reported")
	}
}

// BenchmarkJSONRoundtrip times the JSON encoding and decoding of a
// representative type.
func BenchmarkJSONRoundtrip(b *testing.B) {
	var rec all
	var err error
	var data []byte
	recPopulate(&rec)
	b.ResetTimer()
	for j := 0; err == nil && j < b.N; j++ {
		data, err = jsonRecToBuf(rec)
		if err == nil {
			rec, err = jsonBufToRec(data)
		}
	}
	b.StopTimer()
	if err != nil {
		b.Error(err)
	}
}

// BenchmarkStoreRoundtrip times the store encoding and decoding of a
// representative type.
func BenchmarkStoreRoundtrip(b *testing.B) {
	var rec all
	var err error
	var data []byte
	recPopulate(&rec)
	b.ResetTimer()
	for j := 0; err == nil && j < b.N; j++ {
		data, err = storeRecToBuf(rec)
		if err == nil {
			rec, err = storeBufToRec(data)
		}
	}
	b.StopTimer()
	if err != nil {
		b.Error(err)
	}
}
