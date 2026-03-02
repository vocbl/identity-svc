package domain

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
)

type id struct {
	ulid.ULID
}

func parse(s string) (id, error) {
	val, err := ulid.Parse(s)
	if err != nil {
		return id{}, ErrInvalidIdentifier
	}
	return id{val}, nil
}

func newID() id {
	return id{
		ULID: ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.Reader, 0)),
	}
}

type VerificationSessionID struct {
	id
}

func ParseVerificationSessionID(s string) (VerificationSessionID, error) {
	val, err := parse(s)
	return VerificationSessionID{val}, err
}

func newVerificationSessionID() VerificationSessionID {
	return VerificationSessionID{newID()}
}

type UserID struct {
	id
}

func ParseUserID(s string) (UserID, error) {
	val, err := parse(s)
	return UserID{val}, err
}

func newUserID() UserID {
	return UserID{newID()}
}
