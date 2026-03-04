package port

import errutil "github.com/vocbl/shared/errors"

var (
	ErrRegisterHandler = errutil.NewHandlerSt("port.register", func(err error) error {
		return err
	})
)
