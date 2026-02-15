package port

import (
	"errors"

	errutil "github.com/vocbl/shared/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	ErrRegisterHandler = errutil.NewHandlerSt("port.register", func(err error) error {
		if errors.Is(err, errutil.ErrValidation) {
			br := &errdetails.BadRequest{}
		}
	})
)
