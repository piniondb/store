#store [![GoDoc](https://godoc.org/jung/store?status.png)](https://godoc.org/jung/store)

Package store helps applications convert structured data quickly to and from compact byte sequences.

Package store helps applications convert structured data quickly to and from
compact byte sequences.

##Overview
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

##Example
Use a store.PutBuffer to pack individual values (either standalone or members
of a structure) into a byte sequence. For example,

```
	put := store.NewPutBuffer()
	put.Uint32(a)
	put.Int64(b)
	data, err := put.Bytes()
```

where a is of type uint32 and b is of type int64. To retrieve these values from
the byte slice, use a store.GetBuffer:

```
	get := store.NewGetBuffer(data)
	get.Uint32(&a)
	get.Int64(&b)
	err := get.Done()
```

The sequence of put and get method calls must mirror each other. An error will
be returned if too few or too many values are extracted from a Get buffer, but
otherwise making sure that the inbound and outbound value types match is up to
the programmer. In practice, this is easily done if functions are written to
handle the getting and putting of fields and are kept close in the code to the
structure definition. You can enhance the generality of your code by using
these conversion functions to implement the encoding.MarshalBinary and
encoding.UnmarshalBinary interfaces.

See the Go documentation for more complete examples.

##Installation
To install the package on your system, run

```
go get github.com/piniondb/store
```

##Errors
Converting data generally involves a lot of steps. This can make error checking
onerous. If an error occurs in a PutBuffer, GetBuffer, or KeyBuffer method, an
internal error field is set. After this occurs, subsequent method calls
typically return without performing any operations and the error state is
retained. This error management scheme facilitates data conversion since
individual method calls do not need to be examined for failure; it is generally
sufficient to wait until after put.Bytes(), get.Done() or key.Bytes() is
called. For the same reason, if an error occurs in the calling application
during conversion, it may be desirable for the application to transfer the
error to the buffer instance by calling its SetError() method.

##Contributing Changes
store is a global community effort and you are invited to make it even better.
If you have implemented a new feature or corrected a problem, please consider

Here are guidelines for making submissions. Your change should

* be compatible with the MIT License
* be properly documented
* be formatted with `go fmt`
* include an example in store_test.go if appropriate
* conform to the standards of golint (https://github.com/golang/lint) and
go vet (https://godoc.org/golang.org/x/tools/cmd/vet), that is, `golint .` and
`go vet .` should not generate any warnings
* not diminish test coverage (https://blog.golang.org/cover)

[Pull requests](https://help.github.com/articles/using-pull-requests/) work
nicely as a means of contributing your changes.

##License
store is released under the MIT License.

