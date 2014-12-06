/*Package recovery is a HTTP middleware that catches any panics and serves a proper error response.

  package main

  import (
      "log"
      "net/http"

      "github.com/unrolled/recovery"  // or "gopkg.in/unrolled/recovery.v1"
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

A GET request to "/" will output:

  [MySampleWebApp] 2014/12/05 23:15:11 Recovering from Panic: you should not have a handler that just panics ;)
  goroutine 5 [running]:
  github.com/unrolled/recovery.func·001()
      /$GOPATH/src/github.com/unrolled/recovery/recovery.go:86 +0x12a
  main.func·001(0x4a1008, 0xc20801e6c0, 0xc2080324e0)
      /$GOPATH/src/thisapp.go:12 +0x64
  ...
*/
package recovery
