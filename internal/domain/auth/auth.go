package domain

// type deviceNature string

// var (
// 	DeviceTypeMobile  deviceNature = "mobile"
// 	DeviceTypeTablet  deviceNature = "tablet"
// 	DeviceTypeDesktop deviceNature = "desktop"
// 	DeviceTypeUnknown deviceNature = "unknown"
// )

// func parseDeviceNature(s string) deviceNature {
// 	switch deviceType := deviceNature(strings.ToLower(strings.TrimSpace(s))); deviceType {
// 	case
// 		DeviceTypeMobile,
// 		DeviceTypeTablet,
// 		DeviceTypeDesktop:
// 		return deviceType
// 	}

// 	return DeviceTypeUnknown
// }

// type devicePlatform string

// const (
// 	DevicePlatformIOS     devicePlatform = "ios"
// 	DevicePlatformAndroid devicePlatform = "android"
// 	DevicePlatformWeb     devicePlatform = "web"
// 	DevicePlatformMacOS   devicePlatform = "macos"
// 	DevicePlatformWindows devicePlatform = "windows"
// 	DevicePlatformLinux   devicePlatform = "linux"
// 	DevicePlatformUnknown devicePlatform = "unknown"
// )

// func parseDevicePlatform(s string) devicePlatform {
// 	switch devicePlatform := devicePlatform(strings.ToLower(strings.TrimSpace(s))); devicePlatform {
// 	case
// 		DevicePlatformIOS,
// 		DevicePlatformAndroid,
// 		DevicePlatformWeb,
// 		DevicePlatformMacOS,
// 		DevicePlatformWindows,
// 		DevicePlatformLinux:
// 		return devicePlatform

// 	}

// 	return DevicePlatformUnknown
// }

// type device struct {
// 	nature   deviceNature
// 	platform devicePlatform
// }

// func newDevice(deviceNature, devicePlatform string) device {
// 	return device{
// 		nature:   parseDeviceNature(deviceNature),
// 		platform: parseDevicePlatform(devicePlatform),
// 	}
// }

// type UserAuthSession struct {
// 	ID        AuthSessionID
// 	UserID    UserID
// 	Device    device
// 	IsRevoked bool
// 	ExpiresAt time.Time
// }

// func (uas *UserAuthSession) GenerateAccessToken(policy *AuthPolicy) (AccessToken, error) {
// 	if uas.IsRevoked {
// 		return AccessToken{}, ErrAuthSessionRevoked
// 	}

// 	if uas.ExpiresAt.Before(time.Now().UTC()) {
// 		return AccessToken{}, ErrAuthSessionExpired
// 	}

// 	return policy.generateAccessToken(uas.UserID)
// }

// type RefreshToken struct {
// 	sessionID AuthSessionID
// 	expiresAt time.Time
// 	token     string
// }

// type AccessToken struct {
// 	token string
// }

// func NewUserAuthSession(policy *AuthPolicy, userID UserID, deviceNature, devicePlatform string) (*UserAuthSession, RefreshToken, error) {
// 	sessionID := newAuthSessionID()

// 	refreshToken, err := policy.generateRefreshToken(userID, sessionID)
// 	if err != nil {
// 		return nil, RefreshToken{}, err
// 	}

// 	return &UserAuthSession{
// 		ID:        sessionID,
// 		UserID:    userID,
// 		ExpiresAt: refreshToken.expiresAt,
// 		Device:    newDevice(deviceNature, devicePlatform),
// 	}, refreshToken, nil
// }

// type AuthPolicy struct {
// 	refreshTokenDuration time.Duration
// 	accessTokenDuration  time.Duration
// 	maxEttempts          int
// 	generateRefreshToken func(userID UserID, sessionID AuthSessionID) (RefreshToken, error)
// 	generateAccessToken  func(userID UserID) (AccessToken, error)
// 	parseRefreshToken    func(token RefreshToken) (AuthSessionID, UserID, error)
// 	hashPassword         func(password Password) (PasswordHash, error)
// }
