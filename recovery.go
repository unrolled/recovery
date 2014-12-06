package recovery

import (
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

// Options is a struct for specifying configuration parameters for the Recovery middleware.
type Options struct {
	// IncludeFullStack if set to true, will dump the complete stack instead of the single goroutine that panicked. Default is false (single goroutine only).
	IncludeFullStack bool
	// StackSize sets how large the []byte buffer is for the stack dump. Default is 8192.
	StackSize int
	// Prefix is the outputted keyword in front of the log message. Logger automatically wraps the prefix in square brackets (ie. [myApp] ) unless the `DisableAutoBrackets` is set to true. A blank value will not have brackets added. Default is blank (with no brackets).
	Prefix string
	// DisableAutoBrackets if set to true, will remove the prefix and square brackets. Default is false.
	DisableAutoBrackets bool
	// Out is the destination to which the logged data will be written too. Default is `os.Stderr`.
	Out io.Writer
	// OutputFlags defines the logging properties. See http://golang.org/pkg/log/#pkg-constants. To disable all flags, set this to `-1`. Defaults to log.LstdFlags (2009/01/23 01:23:23).
	OutputFlags int
}

// Recovery is a HTTP middleware that catches any panics and serves a proper error response.
type Recovery struct {
	*log.Logger
	opt          Options
	panicHandler http.Handler
}

// New returns a new Recovery instance.
func New(opts ...Options) *Recovery {
	var o Options
	if len(opts) == 0 {
		o = Options{}
	} else {
		o = opts[0]
	}

	// Stacksize
	if o.StackSize <= 0 {
		o.StackSize = 8 * 1024
	}

	// Determine prefix.
	prefix := o.Prefix
	if len(prefix) > 0 && o.DisableAutoBrackets == false {
		prefix = "[" + prefix + "] "
	}

	// Determine output writer.
	var output io.Writer
	if o.Out != nil {
		output = o.Out
	} else {
		// Default is stderr.
		output = os.Stderr
	}

	// Determine output flags.
	flags := log.LstdFlags
	if o.OutputFlags == -1 {
		flags = 0
	} else if o.OutputFlags != 0 {
		flags = o.OutputFlags
	}

	return &Recovery{
		Logger:       log.New(output, prefix, flags),
		opt:          o,
		panicHandler: http.HandlerFunc(defaultPanicHandler),
	}
}

// Handler wraps an HTTP handler and recovers any panics from up stream.
func (r *Recovery) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				r.panicHandler.ServeHTTP(w, req)

				stack := make([]byte, r.opt.StackSize)
				stack = stack[:runtime.Stack(stack, r.opt.IncludeFullStack)]

				r.Printf("Recovering from Panic: %s\n%s", err, stack)
			}
		}()

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(fn)
}

// SetPanicHandler sets the handler to call when Recovery encounters a panic.
func (r *Recovery) SetPanicHandler(handler http.Handler) {
	r.panicHandler = handler
}

func defaultPanicHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
