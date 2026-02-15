package identity

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
)

type ID struct {
	ulid.ULID
}

func (id ID) IsNil() bool {
	return id.Compare(ulid.ULID{}) == 0
}

func ParseID(s string) (ID, error) {
	val, err := ulid.Parse(s)
	if err != nil {
		return ID{}, ErrInvalidIdentifier
	}
	return ID{val}, nil
}

func NewID() ID {
	return ID{
		ULID: ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.Reader, 0)),
	}
}

func RebuildID(id string) (ID, error) {
	val, err := ulid.Parse(id)
	if err != nil {
		return ID{}, err
	}

	return ID{val}, nil
}
