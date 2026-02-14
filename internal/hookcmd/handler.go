package hookcmd

import (
	"context"
	"fmt"
	"io"
)

// Handler processes a hook event.
type Handler interface {
	Name() string
	Run(ctx context.Context, input *HookInput, out io.Writer, errOut io.Writer) error
}

// RunHandlers executes handlers sequentially. Errors are logged to errOut
// but do not stop subsequent handlers. Panics are recovered.
func RunHandlers(ctx context.Context, input *HookInput, handlers []Handler, out, errOut io.Writer) {
	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(errOut, "[%s] panic recovered: %v\n", h.Name(), r)
				}
			}()
			if err := h.Run(ctx, input, out, errOut); err != nil {
				fmt.Fprintf(errOut, "[%s] error: %v\n", h.Name(), err)
			}
		}()
	}
}
