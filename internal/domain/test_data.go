package identity

import (
	"time"
)

const (
	//User
	TestValidPassword     = "TestPassword123"
	TestValidPasswordHash = "$2a$10$8K1p/a0DX1.A8At9.S8ObeS8ObeS8ObeS8ObeS8ObeS8ObeS8ObeS"
	TestValidEmail        = "valid@email.test"
	TestValidFirstName    = "John"
	TestValidLastName     = "Doe"
	TestValidUsername     = "TestJohn"

	TestInvalidPassword     = "short"
	TestInvalidPasswordHash = "invalid-hash"
	TestInvalidEmail        = "not-an-email"
	TestInvalidFirstName    = ""
	TestInvalidLastName     = ""
	TestInvalidUsername     = ""

	//ID
	TestInvalidID = "not-a-ulid"
)

func TestValidUser() *User {
	return &User{
		id:    NewUserID(),
		email: TestValidEmail,
		creds: UserCreds{
			firstName: TestValidFirstName,
			lastName:  TestValidLastName,
			username:  TestValidUsername,
		},
		createdAt: time.Now().Add(-time.Minute * 3),
	}
}

func TestValidUserWithExteralIdentity() (*User, ExternalIdentity) {
	externalIdentity := ExternalIdentity{
		provider:   ExternalProviderGoogle,
		externalID: "external-id",
	}

	return &User{
		id:    NewUserID(),
		email: TestValidEmail,
		creds: UserCreds{
			firstName: TestValidFirstName,
			lastName:  TestValidLastName,
			username:  TestValidUsername,
		},
		externalIdentities: []ExternalIdentity{externalIdentity},
		createdAt:          time.Now().Add(-time.Minute * 3),
	}, externalIdentity
}
