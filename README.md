# Recovery [![GoDoc](https://godoc.org/github.com/unrolled/recovery?status.svg)](http://godoc.org/github.com/unrolled/recovery) [![Build Status](https://travis-ci.org/unrolled/recovery.svg)](https://travis-ci.org/unrolled/recovery)

Recovery is a HTTP middleware that catches any panics and serves a proper error response. It's a standard net/http [Handler](http://golang.org/pkg/net/http/#Handler), and can be used with many frameworks or directly with Go's net/http package.

## Usage

~~~ go
// main.go
package main

import (
    "log"
    "net/http"

     "github.com/unrolled/recovery"
)

var myPanicHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    panic("you should not have a handler that just panics ;)")
})

func main() {
    recoveryMiddleware := recovery.New(recovery.Options{
        Prefix: "MySampleWebApp",
        OutputFlags: log.LstdFlags,
    })

    // recoveryWithDefaults := recovery.New()

    app := recoveryMiddleware.Handler(myPanicHandler)
    http.ListenAndServe("0.0.0.0:3000", app)
}
~~~

A simple GET request to "/" will output:
~~~ bash
[MySampleWebApp] 2014/12/05 23:15:11 Recovering from Panic: you should not have a handler that just panics ;)
goroutine 5 [running]:
github.com/unrolled/recovery.func·001()
    /$GOPATH/src/github.com/unrolled/recovery/recovery.go:86 +0x12a
main.func·001(0x4a1008, 0xc20801e6c0, 0xc2080324e0)
    /$GOPATH/src/thisapp.go:12 +0x64
...
~~~

If you are using a logging middleware like [Logger](https://github.com/unrolled/logger) (which you should be), be sure the logger is first followed by Recovery. This will ensure that recovered handlers will still be logged (ie. you can see a 500 in your log files).

### Available Options
Recovery comes with a variety of configuration options (Note: these are not the default option values. See the defaults below.):

~~~ go
// ...
r := recovery.New(recovery.Options{
    IncludeFullStack: false, // IncludeFullStack if set to true, will dump the complete stack instead of the single goroutine that panicked. Default is false (single goroutine only).
    StackSize: 8 * 1024, // StackSize sets how large the []byte buffer is for the stack dump. Default is 8192.
    Prefix: "myAppRecov", // Prefix is the outputted keyword in front of the log message. Logger automatically wraps the prefix in square brackets (ie. [myApp] ) unless the `DisableAutoBrackets` is set to true. A blank value will not have brackets added. Default is blank (with no brackets).
    DisableAutoBrackets: false, // DisableAutoBrackets if set to true, will remove the prefix and square brackets. Default is false.
    Out: os.Stderr, // Out is the destination to which the logged data will be written too. Default is `os.Stderr`.
    OutputFlags: log.Ldate | log.lTime, // OutputFlags defines the logging properties. See http://golang.org/pkg/log/#pkg-constants. To disable all flags, set this to `-1`. Defaults to log.LstdFlags (2009/01/23 01:23:23).
})
// ...
~~~

### Default Options
These are the preset options for Recovery:

~~~ go
r := recovery.New()

// Is the same as the default configuration options:

r := recovery.New(recovery.Options{
    IncludeFullStack: false,
    StackSize: 8 * 1024,      
    Prefix: "",
    DisableAutoBrackets: false,
    Out: os.Stderr,
    OutputFlags log.LstdFlags,
})
~~~

### Include Full Stack
Be aware that including the full stack could produce a very large dump. If `IncludeFullStack` is true, Recovery logs stack traces of all other goroutines after the the current goroutine is logged. So if you do need a complete stack trace be sure to increase the `StackSize` to something huge like `256 * 1024`.
