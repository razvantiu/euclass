package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/business/web/trusted"
	"github.com/ardanlabs/service/foundation/web"
	"go.uber.org/zap"
)

// Error handles errors from the core handlers.
func Error(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			// Run the next handler and catch any propagated error.
			if err := handler(ctx, w, r); err != nil {

				// Log the error.
				log.Errorw("ERROR", "traceid", v.TraceID, "ERROR", err)

				// Build out the error response.
				var er trusted.ErrorResponse
				var status int
				switch {
				case trusted.IsRequestError(err):
					reqErr := trusted.GetRequestError(err)
					er = trusted.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = trusted.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// Respond with the error back to the client.
				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
