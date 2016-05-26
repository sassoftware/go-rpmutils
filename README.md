Go RPM Utils
============
go-rpmutils is a library written in [go](http://golang.org) for parsing and extracting content from [RPMs](http://www.rpm.org).

#Overview
go-rpmutils provides a few interfaces for handling RPM packages. There is a highlevel `Rpm` struct that provides access to the RPM header and [CPIO](https://en.wikipedia.org/wiki/Cpio) payload. The CPIO payload can be extracted to a filesystem location via the `ExpandPayload` function or through a Reader interface, similar to the [tar implementation](https://golang.org/pkg/archive/tar/) in the go standard library.

#Example
```go
func Main() {
    f, err := os.Open("foo.rpm")
    if err != nil {
        panic(err)
    }

    // Parse the rpm
    rpm := rpmutils.ReadRpm(f)

    // Get the name, epoch, version, release, and arch
    nevra, err := rpm.GetNEVRA()
    if err != nil {
        panic(err)
    }

    fmt.Printf("%s\n", nevra)

    // Reading the provides header
    provides, err := rpm.Header.GetStrings(rpmutils.PROVIDENAME)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Provides:\n")
    for _, p := range provides {
        fmt.Printf("%s", p)
    }
}
```

## Contributing

1. Read contributor agreement
2. Fork it
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Push to the branch (`git push origin my-new-feature`)
6. Create new Pull Request


## License

go-rpmutils is released under the Apache 2.0 license. See [LICENSE](https://github.com/sassoftware/go-rpmutils/blob/master/LICENSE).
