package app

// type AuthService struct {
// 	repo   AuthRepo
// 	cache  authCache
// 	policy *domain.AuthPolicy
// }

// func (s *AuthService) Login(ctx context.Context, emailStr, passwordStr, deviceType, devicePlatform string) (domain.RefreshToken, domain.AccessToken, error) {
// 	email, err := domain.ParseEmail(emailStr)
// 	if err != nil {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	password, err := domain.ParsePassword(passwordStr)
// 	if err != nil {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	hashedPassword, userID, verified, err := s.repo.GetAuthData(ctx, email)
// 	if err != nil {
// 		if errors.Is(err, db.ErrNonExistingData) {
// 			err = ErrNonExistingUser
// 			s.policy.DummyPasswordCheck()
// 		}
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	if !verified {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(ErrUserPendingVerification)
// 	}

// 	err = hashedPassword.Verify(s.policy, password)
// 	if err != nil {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	authSession, refreshToken, err := domain.NewUserAuthSession(s.policy, userID, deviceType, devicePlatform)
// 	if err != nil {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	accessToken, err := authSession.GenerateAccessToken(s.policy)
// 	if err != nil {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	err = s.repo.WithinTransaction(ctx, func(txRepo AuthRepo) error {
// 		err := txRepo.SaveSession(ctx, authSession)
// 		if err != nil {
// 			return err
// 		}

// 		return txRepo.EmitNewSessionEvent(ctx, authSession.ID)
// 	})

// 	if err != nil {
// 		return domain.RefreshToken{}, domain.AccessToken{}, ErrLoginOperation.Wrap(err)
// 	}

// 	return refreshToken, accessToken, nil
// }

// func (s *AuthService) IssueAccessToken(ctx context.Context, refreshJWT string) (string, error) {
// 	refreshTokenID, userID, err := s.verifyRefreshJWT(refreshJWT)
// 	if err != nil {
// 		return "", ErrIssueAccessTokenOperation.Wrap(ErrInvalidRefreshToken)
// 	}

// 	isActive, err := s.cache.IsActive(ctx, refreshTokenID)
// 	if err != nil {
// 		isActive, err = s.repo.IsActive(ctx, refreshTokenID)
// 		if err != nil {
// 			return "", ErrIssueAccessTokenOperation.Wrap(err)
// 		}
// 	}

// 	if !isActive {
// 		return "", ErrIssueAccessTokenOperation.Wrap(ErrRevokedRefreshToken)
// 	}

// 	accessToken, err := security.GenerateAccessJWT(userID, s.jwt.AccessDuration, s.jwt.AccessSecretKey)
// 	return accessToken, ErrIssueAccessTokenOperation.WrapCheck(err)
// }
