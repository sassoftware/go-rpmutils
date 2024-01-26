# Go RPM Utils

[![Go Reference](https://pkg.go.dev/badge/github.com/sassoftware/go-rpmutils.svg)](https://pkg.go.dev/github.com/sassoftware/go-rpmutils)

go-rpmutils is a library written in [go](http://golang.org) for parsing and extracting content from [RPMs](http://www.rpm.org).

## Overview

go-rpmutils provides a few interfaces for handling RPM packages. There is a highlevel `Rpm` struct that provides access to the RPM header and [CPIO](https://en.wikipedia.org/wiki/Cpio) payload. The CPIO payload can be extracted to a filesystem location via the `ExpandPayload` function or through a Reader interface, similar to the [tar implementation](https://golang.org/pkg/archive/tar/) in the go standard library.

## Example

```go
// Opening a RPM file
f, err := os.Open("foo.rpm")
if err != nil {
    panic(err)
}
rpm, err := rpmutils.ReadRpm(f)
if err != nil {
    panic(err)
}
// Getting metadata
nevra, err := rpm.Header.GetNEVRA()
if err != nil {
    panic(err)
}
fmt.Println(nevra)
provides, err := rpm.Header.GetStrings(rpmutils.PROVIDENAME)
if err != nil {
    panic(err)
}
fmt.Println("Provides:")
for _, p := range provides {
    fmt.Println(p)
}
// Extracting payload
if err := rpm.ExpandPayload("destdir"); err != nil {
    panic(err)
}
```

## Validating Signatures

RPMs may contain one or more PGP signatures.
Unfortunately, Linux distributions have traditionally lagged far behind
on what signature configurations they use by default.
The PGP V3 signature format was replaced in 1998, yet even some new
versions of popular Linux distributions continue to ship with this format.

Meanwhile, cryptographic libraries continue to prune obsolete configurations
in an effort to make software more secure by default.
This creates an awkward situation for tools which need to interact
with RPM signatures, as the intersection between supported cryptography
and supported Linux distributions vanishes.

To make this a little less painful to deal with,
`rpmutils` allows you to select different PGP implementations
depending on what kind of signatures you need to consume.
However, it always generates V4 or newer signatures,
regardless of which library you select.
All current RPM-based distributions can verify V4 signatures
even if the base OS ships with V3 by default.

### golang.org/x/crypto/openpgp

This implementation was formerly part of the standard library and is now deprecated.
It supports parsing V3 and V4 signatures.

```go
import (
    "github.com/sassoftware/go-rpmutils"
    "github.com/sassoftware/go-rpmutils/stdpgp"
    "golang.org/x/crypto/openpgp"
)

func main() {
    kf, err := os.Open("trusted.pgp")
    keyring, err := openpgp.ReadArmoredKeyRing(kf)
    f, err := os.Open("foo.rpm")
    hdr, sigs, err := rpmutils.VerifyStream(f, stdpgp.Verifier{EntityList: keyring})
}
```

### github.com/ProtonMail/go-crypto/openpgp

An actively maintained fork from the standard library.
It supports V4 and newer signatures.

```go
import (
    "github.com/sassoftware/go-rpmutils"
    "github.com/sassoftware/go-rpmutils/pmpgp"
    "github.com/ProtonMail/go-crypto/openpgp"
)

func main() {
    kf, err := os.Open("trusted.pgp")
    keyring, err := openpgp.ReadArmoredKeyRing(kf)
    f, err := os.Open("foo.rpm")
    hdr, sigs, err := rpmutils.VerifyStream(f, pmpgp.Verifier{EntityList: keyring})
}
```

### github.com/pgpkeys-eu/go-crypto

This is a soft fork of the ProtonMail repository that adds back support for V3 signatures.
Because it uses the same module name as the ProtonMail version,
it must be consumed using a `replace` directive.
Additionally, if you want to validate V3 signatures you must build with a `sigv3` tag
in order to enable the `rpmutils` portion of the signature implementation.

Building the same example code as above:
```
go mod edit -replace github.com/ProtonMail/go-crypto=github.com/pgpkeys-eu/go-crypto@main
go mod tidy
go build -t sigv3
```

## Contributing

1. Read contributor agreement
2. Fork it
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes (`git commit -a`). Make sure to include a Signed-off-by line per the contributor agreement.
5. Push to the branch (`git push origin my-new-feature`)
6. Create new Pull Request

## License

go-rpmutils is released under the Apache 2.0 license. See [LICENSE](https://github.com/sassoftware/go-rpmutils/blob/master/LICENSE).
