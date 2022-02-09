// Package testgrp is for learning.
package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/ardanlabs/service/business/web/trusted"
	"github.com/ardanlabs/service/foundation/web"
)

func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Int31n(100); n%2 == 0 {
		return trusted.NewRequestError(errors.New("TRUST ME"), http.StatusBadRequest)
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
