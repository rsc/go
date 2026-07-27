## Tools {#tools}

### Go command {#go-command}

The `go test` command now applies a per-test timeout in addition to the
existing whole-binary `-timeout`. By default each test function is limited to
one minute of active running time; this can be changed with the new
`-testtimeout` flag, or per-test with the [testing.T.SetTimeout] method. When
`-testtimeout` is not set but `-timeout` is set explicitly, the `-timeout`
value is used as the per-test default.

### Cgo {#cgo}

### Vet {#vet}

The new [`scannererr`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/scannererr)
analyzer checks for failure to handle scanner errors after a loop
around [bufio.Scanner.Scan], which may cause scanning or I/O errors to
go unreported. <!-- /issue/17747/ -->

The [`sqlrowserr`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/sqlrowserr)
analyzer performs a similar check for loops around [sql.Rows.Next],
so that iteration errors are correctly distinguished from a smaller result.
