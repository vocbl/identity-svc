package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"
	ulidutil "github.com/vocbl/shared/utils/ulid"
	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
	"github.com/vocbl/users-svc/internal/shared/security"
)

type OnboardService struct {
	repo               OnboardRepo
	verificationPolicy domain.VerificationPolicy
	hashPassword       security.PasswordHasher
	generateToken      security.TokenGenerator
}

func NewOnboardService(
	repo OnboardRepo,
	verificationPolicy domain.VerificationPolicy,
	passwordHasher security.PasswordHasher,
	tokenGenerator security.TokenGenerator,
) (*OnboardService, error) {

	if err := verificationPolicy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid verification policy config: %w", err)
	}

	token, tokenHash := tokenGenerator()
	if err := domain.ValidateTokenHash(tokenHash); err != nil {
		return nil, fmt.Errorf("invalid tokenGenerator: %w", err)
	}
	if err := domain.ValidateToken(token); err != nil {
		return nil, fmt.Errorf("invalid tokenGenerator: %w", err)
	}

	passwordHash, err := passwordHasher(domain.TestValidPassword)
	if err != nil {
		return nil, fmt.Errorf("invalid passwordHasher: failed to hash password: %w", err)
	}
	if err := domain.ValidatePasswordHash(passwordHash); err != nil {
		return nil, fmt.Errorf("invalid passwordHasher: %w", err)
	}

	return &OnboardService{
		repo:               repo,
		generateToken:      tokenGenerator,
		hashPassword:       passwordHasher,
		verificationPolicy: verificationPolicy,
	}, nil
}

type NewUser struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Username  string
}

func (nu NewUser) ToVerificationSession() (*domain.UserVerificationSession, string) {
	return &domain.UserVerificationSession{
		Email: nu.Email,
		Creds: domain.UserCreds{
			FirstName: nu.FirstName,
			LastName:  nu.LastName,
			Username:  nu.Username,
		},
	}, nu.Password
}

func (s *OnboardService) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, ErrCheckUsernameAvailabilityOperation.Wrap(domain.ErrInvalidUsername)
	}

	exists, err := s.repo.CheckUsernameExistance(ctx, username)
	if err != nil {
		return false, ErrCheckUsernameAvailabilityOperation.Wrap(err)
	}

	return !exists, nil
}

func (s *OnboardService) CreateVerificationSession(ctx context.Context, user NewUser) (ulid.ULID, time.Time, error) {
	session, password := user.ToVerificationSession()

	if errs := session.Validate(password); errs != nil {
		return ulid.ULID{}, time.Time{}, ErrCreateVerificationSessionOperationValidation.Wrap(errs...)
	}

	token, tokenHash := s.generateToken()
	session.Enrich(tokenHash, s.verificationPolicy.Duration, time.Now().UTC().Add(s.verificationPolicy.RestartDuration))

	err := s.repo.WithinTransaction(ctx, func(txRepo OnboardRepo) error {
		err := txRepo.SaveUserVerificationSession(ctx, session)
		if err != nil {
			return err
		}

		return txRepo.EmitUserCreatedEvent(ctx, session.SessionID.String(), session.Email, password, token)
	})

	if err != nil {
		var sessionID ulid.ULID
		var restartableSince time.Time

		switch {
		case errors.Is(err, db.ErrDublicateEmail):
			err = ErrEmailAlreadyExists
		case errors.Is(err, db.ErrDublicateUsername):
			err = ErrUsernameTaken
		default:
			var upvErr db.UserPendingVerificationError
			if errors.As(err, &upvErr) {
				err = ErrUserPendingVerification
				sessionID = upvErr.SessionID
				restartableSince = upvErr.RestartableSince
			}

		}
		return sessionID, restartableSince, ErrCreateVerificationSessionOperation.Wrap(err)
	}

	return session.SessionID, session.RestartableSince, nil
}

func (s *OnboardService) StartVerificationSession(ctx context.Context, sessionIdStr string, email, password, token string) error {
	sessionID, err := ulid.Parse(sessionIdStr)
	if err != nil {
		return ErrStartVerificationSessionOperation.Wrap(ErrInvalidULID)
	}

	var errs []error
	if !domain.ValidateEmail(email) {
		errs = append(errs, domain.ErrInvalidEmail)
	}
	if !domain.ValidatePassword(password) {
		errs = append(errs, domain.ErrInvalidPassword)
	}
	if err = domain.ValidateToken(token); err != nil {
		errs = append(errs, err)
	}

	if errs != nil {
		return ErrStartVerificationSessionOperationValidation.Wrap(errs...)
	}

	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return ErrStartVerificationSessionOperation.Wrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo OnboardRepo) error {
		err := txRepo.SetUserVerificationSessionPasswordHash(ctx, sessionID, passwordHash)
		if err != nil {
			return err
		}

		return txRepo.EmitVerifyUserEvent(ctx, email, token)
	})

	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNonExistingVerificationSession
		}
		return ErrStartVerificationSessionOperation.Wrap(err)
	}

	return nil
}

func (s *OnboardService) DeleteVerificationSession(ctx context.Context, sessionIdStr string) error {
	sessionID, err := ulid.Parse(sessionIdStr)
	if err != nil {
		return ErrStartVerificationSessionOperation.Wrap(ErrInvalidULID)
	}

	return ErrStartVerificationSessionOperation.WrapCheck(s.repo.DeleteUserVerificationSession(ctx, sessionID))
}

func (s *OnboardService) RestartVerification(ctx context.Context, strSessionID string) (time.Time, error) {
	sessionID, err := ulid.Parse(strSessionID)
	if err != nil {
		return time.Time{}, ErrRestartVerificationSessionOperation.Wrap(ErrInvalidULID)
	}

	token, tokenHash := s.generateToken()

	restartableSince := time.Now().UTC().Add(s.verificationPolicy.RestartDuration)
	err = s.repo.WithinTransaction(ctx, func(txRepo OnboardRepo) error {
		email, attempts, err := txRepo.RestartUserVerificationSession(ctx, sessionID, restartableSince, tokenHash)
		if err != nil {
			return err
		}

		if s.verificationPolicy.MaxAttempts < attempts {
			err = txRepo.DeleteUserVerificationSession(ctx, sessionID)
			if err != nil {
				return err
			}
			return db.ErrUserVerificationRestartAttempts
		}

		return txRepo.EmitVerifyUserEvent(ctx, email, token)
	})

	if err != nil {
		switch {
		case errors.Is(err, db.ErrNonExistingData):
			err = ErrNonExistingVerificationSession
		case errors.Is(err, db.ErrUserVerificationCompleted):
			err = ErrVerificationSessionAlreadyCompleted
		case errors.Is(err, db.ErrUserVerificationRestartSince):
			err = ErrUserVerificationRestartSince
		case errors.Is(err, db.ErrUserVerificationRestartAttempts):
			err = ErrUserVerificationRestartAttempts
		}

		return time.Time{}, ErrRestartVerificationSessionOperation.Wrap(err)
	}

	return restartableSince, nil
}

func (s *OnboardService) CompleteVerification(ctx context.Context, token string) error {
	err := domain.ValidateToken(token)
	if err != nil {
		return ErrCompleteVerificationOperation.Wrap(domain.ErrInvalidToken)
	}

	err = s.repo.WithinTransaction(ctx, func(tx OnboardRepo) error {
		session, err := tx.CompleteUserVerificationSession(ctx, security.HashToken(token))
		if err != nil {
			return err
		}

		if err = tx.SaveUser(ctx, session.ToUser(ulidutil.NewULID())); err != nil {
			return err
		}

		return tx.EmitUserVerifiedEvent(ctx, session.SessionID)
	})

	if err != nil {
		switch {
		case errors.Is(err, db.ErrNonExistingData):
			err = ErrNonExistingVerificationSession
		case errors.Is(err, db.ErrUserVerificationCompleted):
			err = ErrVerificationSessionAlreadyCompleted
		case errors.Is(err, db.ErrSessionExpired):
			err = ErrUserVerificationSessionExpired
		}

		return ErrCompleteVerificationOperation.Wrap(err)
	}

	return nil
}

func (s *OnboardService) CleanUpVerificationSessions(ctx context.Context) error {
	return ErrCleanUpVerificationSessionsOperation.WrapCheck(s.repo.CleanUpVerificationSessions(ctx, s.verificationPolicy.CleanUpDuration))
}

type OAuthUser struct {
	Email      string
	FirstName  string
	LastName   string
	ExternalID string
}

func (ou OAuthUser) ToDomainUser(provider domain.AuthProvider) *domain.User {
	return &domain.User{
		Email: ou.Email,
		Creds: domain.UserCreds{
			FirstName: ou.FirstName,
			LastName:  ou.LastName,
		},
		ExternalIdentities: []domain.ExternalIdentity{
			{
				ID:       ou.ExternalID,
				Provider: provider,
			},
		},
	}
}

func (s *OnboardService) CreateOauthUser(ctx context.Context, provider domain.AuthProvider, ou OAuthUser) (ulid.ULID, error) {
	user := ou.ToDomainUser(provider)
	errs := user.OauthValidateAndEnrich()
	if errs != nil {
		return ulid.ULID{}, ErrCreateOauthUserOperationValidation.Wrap(errs...)
	}

	var err error
	for i := 1; i < 6; i++ {
		user.Creds.GenerateUsername(i)
		err = s.repo.SaveUser(ctx, user)
		if err != nil {
			switch {
			case errors.Is(err, db.ErrDublicateEmail):
				userID, err := s.repo.SaveUserExternalIdentity(ctx, user.Email, user.ExternalIdentities[0])
				if err != nil {
					if errors.Is(err, db.ErrDublicateData) {
						err = ErrExternalIdentityAlreadyExists
					}
					return ulid.ULID{}, ErrCreateOauthUserOperation.Wrap(err)
				}
				return userID, err
			case errors.Is(err, db.ErrDublicateUsername):
				continue
			default:
				return ulid.ULID{}, ErrCreateOauthUserOperation.Wrap(err)
			}
		}
		return user.ID, nil
	}

	return ulid.ULID{}, ErrCreateOauthUserOperation.Wrap(ErrUniqueUsernameGeneration)
}
