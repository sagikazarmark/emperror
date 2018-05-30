/*
Package airbrake provides Airbrake integration.

Create a new handler as you would create a gobrake.Notifier:

	projectID  := int64(1)
	projectKey := "key"

	handler := airbrake.NewHandler(projectID, projectKey)

If you need access to the underlying Notifier instance (or need more advanced construction), you can access it from the handler:

	handler.Notifier.SetHost("https://errbit.domain.com")

By default Gobrake sends errors asynchronously and expects to be closed before the program finishes:

	func main() {
		defer handler.Close()
	}

If you want to Flush notices you can do it as you would with Gobrake's notifier or you can configure the handler to send notices synchronously:

	handler.SendSynchronously = true
*/
package airbrake

import (
	"github.com/airbrake/gobrake"
	"github.com/goph/emperror"
	"github.com/goph/emperror/internal/keyvals"
)

// Handler is responsible for sending errors to Airbrake/Errbit.
type Handler struct {
	Notifier          *gobrake.Notifier
	SendSynchronously bool
}

// NewHandler creates a new Airbrake handler.
func NewHandler(projectID int64, projectKey string) *Handler {
	return &Handler{
		Notifier: gobrake.NewNotifier(projectID, projectKey),
	}
}

// Handle calls the underlying Airbrake notifier.
func (h *Handler) Handle(err error) {
	// Get HTTP request (if any)
	req, _ := emperror.HttpRequest(err)

	// Expose the stackTracer interface on the outer error (if there is stack trace in the error)
	err = emperror.ExposeStackTrace(err)

	notice := h.Notifier.Notice(err, req, 1)

	// Extract context from the error and attach it to the notice
	if kvs := emperror.Context(err); len(kvs) > 0 {
		notice.Params = keyvals.ToMap(kvs)
	}

	if h.SendSynchronously {
		h.Notifier.SendNotice(notice)
	} else {
		h.Notifier.SendNoticeAsync(notice)
	}

}

// Close closes the underlying Airbrake instance.
func (h *Handler) Close() error {
	return h.Notifier.Close()
}
