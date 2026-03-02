package app

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid"
	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
	"github.com/vocbl/users-svc/internal/shared/security"
	"go.uber.org/zap"
)

type AuthService struct {
	repo               AuthRepo
	logger             *zap.Logger
	cache              authCache
	jwt                JwtCfg
	generateRefreshJWT func(userID string, duration time.Duration) (string, ulid.ULID, error)
	generateAccessJWT  func(userID string, duration time.Duration) (string, error)
	//returns tokenID, userID and error
	verifyRefreshJWT func(token string) (ulid.ULID, ulid.ULID, error)
	verifyPassword   func(string, string) bool
}

type JwtCfg struct {
	RefreshDuration time.Duration
	AccessDuration  time.Duration
}

func NewAuthService(repo AuthRepo, jwt JwtCfg) (*AuthService, error) {
	err := domain.ValidateRefreshTokenDuration(je)
	return &AuthService{
		repo: repo,
		jwt:  jwt,
	}
}

// returns refresh token, access token, verificationSessionID and error
func (s *AuthService) Login(ctx context.Context, email, password, deviceType, devicePlatform string) (string, string, error) {
	if !domain.ValidateEmail(email) {
		return "", "", ErrLoginOperation.Wrap(ErrInvalidEmail)
	}

	hashedPassword, userID, err := s.repo.GetAuthData(ctx, email)
	if err != nil {
		var upvErr db.UserPendingVerificationError
		if errors.As(err, &upvErr) {
			if !s.verifyPassword(password, hashedPassword) {
				return "", "", ErrLoginOperation.Wrap(ErrInvalidPassword)
			}

			return "", "", &ErrUserPendingVerification{upvErr.SessionID, upvErr.RestartableSince}
		} else if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNonExistingUser
		}
		return "", "", ErrLoginOperation.Wrap(err)
	}

	if !s.verifyPassword(password, hashedPassword) {
		return "", "", ErrLoginOperation.Wrap(ErrInvalidPassword)
	}

	refreshToken, refreshTokenID, err := s.generateRefreshJWT(userID.String(), s.jwt.RefreshDuration)
	if err != nil {
		return "", "", ErrLoginOperation.Wrap(err)
	}

	accessToken, err := s.generateAccessJWT(userID.String(), s.jwt.AccessDuration)
	if err != nil {
		return "", "", ErrLoginOperation.Wrap(err)
	}

	userJwtSession, err := domain.NewUserJwtSession(userID, refreshTokenID, s.jwt.RefreshDuration, deviceType, devicePlatform)
	if err != nil {
		return "", "", ErrLoginOperation.Wrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo AuthRepo) error {
		err := txRepo.SaveSession(ctx, userJwtSession)
		if err != nil {
			return err
		}

		return txRepo.EmitNewSessionEvent(ctx, refreshTokenID, s.jwt.RefreshDuration)
	})

	if err != nil {
		return "", "", ErrLoginOperation.Wrap(err)
	}

	return refreshToken, accessToken, nil
}

// func (s *AuthService) GetSessions(ctx context.Context, stringUserID, password string) ([]domain.UserJwtSession, error) {
// 	userID, err := ulid.Parse(stringUserID)
// 	if err != nil {
// 		return nil, ErrGetSessionsOperation.Wrap(ErrInvalidRefreshTokenID)
// 	}

// 	hashedPassword, err := s.repo.GetUserPasswordHash(ctx, userID)
// 	if err != nil {
// 		var upvErr db.UserPendingVerificationError
// 		if errors.As(err, &upvErr) {
// 			if !s.verifyPassword(password, hashedPassword) {
// 				return "", "", ErrLoginOperation.Wrap(ErrInvalidPassword)
// 			}

// 			return "", "", &ErrUserPendingVerification{upvErr.SessionID, upvErr.RestartableSince}
// 		} else if errors.Is(err, db.ErrNonExistingData) {
// 			err = ErrNonExistingUser
// 		}
// 		return "", "", ErrLoginOperation.Wrap(err)
// 	}

// 	sessions, err := s.repo.GetSessions(ctx, userID)
// 	if err != nil {
// 		if errors.Is(err, db.ErrNonExistingData) {
// 			return nil, ErrGetSessionsOperation.Wrap(ErrNonExistingUser)
// 		}
// 		return nil, ErrGetSessionsOperation.Wrap(err)
// 	}

// 	return sessions, nil
// }

func (s *AuthService) RevokeSession(ctx context.Context, refreshJWT string) error {
	tokenID, userID, err := s.verifyRefreshJWT(refreshJWT)
	if err != nil {
		return ErrRevokeSessionOperation.Wrap(ErrInvalidRefreshToken)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo AuthRepo) error {
		err := txRepo.RevokeSession(ctx, userID, tokenID)
		if err != nil {
			return err
		}

		return txRepo.EmitSessionRevokationEvent(ctx, tokenID)
	})

	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNonExistingSession
		}
		return ErrRevokeSessionOperation.Wrap(err)
	}

	return nil
}

func (s *AuthService) IssueAccessToken(ctx context.Context, refreshJWT string) (string, error) {
	refreshTokenID, userID, err := s.verifyRefreshJWT(refreshJWT)
	if err != nil {
		return "", ErrIssueAccessTokenOperation.Wrap(ErrInvalidRefreshToken)
	}

	isActive, err := s.cache.IsActive(ctx, refreshTokenID)
	if err != nil {
		isActive, err = s.repo.IsActive(ctx, refreshTokenID)
		if err != nil {
			return "", ErrIssueAccessTokenOperation.Wrap(err)
		}
	}

	if !isActive {
		return "", ErrIssueAccessTokenOperation.Wrap(ErrRevokedRefreshToken)
	}

	accessToken, err := security.GenerateAccessJWT(userID, s.jwt.AccessDuration, s.jwt.AccessSecretKey)
	return accessToken, ErrIssueAccessTokenOperation.WrapCheck(err)
}
