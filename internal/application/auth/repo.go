package app

// type AuthRepo interface {
// 	SaveSession(ctx context.Context, session *domain.UserAuthSession) error
// 	UpdateRefreshToken(ctx context.Context, refreshTokenID ulid.ULID, token string) error
// 	RevokeSession(ctx context.Context, userID, refreshTokenID ulid.ULID) error
// 	GetSessions(ctx context.Context, userID ulid.ULID) ([]domain.UserAuthSession, error)
// 	WithinTransaction(ctx context.Context, fn func(txRepo AuthRepo) error) error
// 	EmitNewSessionEvent(ctx context.Context, sessionID domain.AuthSessionID) error

// 	//returns password, userID and error
// 	GetAuthData(ctx context.Context, email domain.Email) (domain.PasswordHash, domain.UserID, bool, error)
// 	GetUserPasswordHash(ctx context.Context, userID ulid.ULID) (string, error)
// 	GetVerificationSessionID(ctx context.Context, email string) (ulid.ULID, error)
// 	EmitSessionRevokationEvent(ctx context.Context, refreshTokenID ulid.ULID) error

// 	IsActive(ctx context.Context, refreshTokenID ulid.ULID) (bool, error)
// }

// type authCache interface {
// 	IsActive(ctx context.Context, refreshTokenID ulid.ULID) (bool, error)
// }
