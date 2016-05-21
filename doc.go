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

/*
Package store helps applications convert structured data quickly to and from
compact byte sequences.

Overview

In order to move data between an application and a storage mechanism such as a
database you generally need to convert native structures to and from byte
sequences. A number of widely used packages such as encoding/json and
encoding/gob perform this conversion but have costs and requirements that may
make them inappropriate for your particular application. For example, the
impressive encoding/gob package is not factored to operate on individual
records and the encoding/json package depends on runtime reflection for each
and every structure conversion. In contrast, the store package works a little
closer to the metal by having the application manually encode a structure and
manually decode a byte sequence. It achieves small encoded data size by
utilizing variable length integer encoding and dispensing with internal field
descriptors. It dramatically increases performance by not using runtime
reflection.

The store package has no dependencies other than the Go standard library. All
tests pass on Linux, Mac and Windows platforms.

Example

Use a store.PutBuffer to pack individual values (either standalone or members
of a structure) into a byte sequence. For example,

	var put store.PutBuffer
	put.Uint32(a)
	put.Int64(b)
	data, err := put.Data()

where a is of type uint32 and b is of type int64. To retrieve these values from
the byte slice, use a store.GetBuffer:

	get := store.NewGetBuffer(data)
	get.Uint32(&a)
	get.Int64(&b)
	err := get.Done()

The sequence of put and get method calls must mirror each other. An error will
be returned if too few or too many values are extracted from a Get buffer, but
otherwise making sure that the inbound and outbound value types match is up to
the programmer. In practice, this is easily done if functions are written to
handle the getting and putting of fields and are kept close in the code to the
structure definition. You can enhance the generality of your code by using
these conversion functions to implement the encoding.BinaryMarshaler and
encoding.BinaryUnmarshaler interfaces.

See the package documentation for more complete examples, including the
conversion of slice and map fields.

Installation

To install the package on your system, run

	go get github.com/piniondb/store

Errors

Converting data generally involves a lot of steps. This can make error checking
onerous. If an error occurs in a PutBuffer, GetBuffer, or KeyBuffer method, an
internal error field is set. After this occurs, subsequent method calls
typically return without performing any operations and the error state is
retained. This error management scheme facilitates data conversion since
individual method calls do not need to be examined for failure; it is generally
sufficient to wait until after put.Data(), get.Done() or key.Data() is
called. For the same reason, if an error occurs in the calling application
during conversion, it may be desirable for the application to transfer the
error to the buffer instance by calling its SetError() method.

Keys

A byte sequence that is used as a key must be sortable. The store package
provides a KeyBuffer to handle this case. Unlike the PutBuffer and GetBuffer
types, a KeyBuffer packs fields, including strings and byte slices, into fixed
length segments. Signed integers are handled by using excess representation in
which the lowest negative value has all bits clear and the highest positive
value has all bits set.

Benchmarks

The following metrics shows how much faster the piniondb/store package is than
the encoding/json package in converting a structure to a byte slice and back
again. The comparison unquestionably involves an apple and an orange since the
JSON encoded value is self-describing. However, if your application does not
require the flexibility of JSON, store encoding may be a fast and viable option
for it.

    BenchmarkJSONRoundtrip-2           50000             28859 ns/op
    BenchmarkStoreRoundtrip-2         500000              3886 ns/op

For the representative data structure used in the examples, the data sequence
produced by encoding/json is three times larger than the one generated by
store.

Contributing Changes

store is a global community effort and you are invited to make it even better.
If you have implemented a new feature or corrected a problem, please consider
contributing your change to the project. Your pull request should

• be compatible with the MIT License

• be properly documented

• include an example in store_test.go if appropriate

Use https://goreportcard.com/report/github.com/piniondb/store to assure that no
compliance issues have been introduced.

License

store is released under the MIT License.

*/
package store
