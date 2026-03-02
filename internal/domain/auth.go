package domain

// import (
// 	"strings"
// 	"time"

// 	"github.com/oklog/ulid"
// 	ulidutil "github.com/vocbl/shared/utils/ulid"
// )

// type deviceType string

// var (
// 	DeviceTypeMobile  deviceType = "mobile"
// 	DeviceTypeTablet  deviceType = "tablet"
// 	DeviceTypeDesktop deviceType = "desktop"
// 	DeviceTypeUnknown deviceType = "unknown"
// )

// func stringToDeviceType(s string) deviceType {
// 	switch deviceType := deviceType(strings.ToLower(strings.TrimSpace(s))); deviceType {
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

// func StringToDevicePlatform(s string) devicePlatform {
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
// 	Type     deviceType
// 	Platform devicePlatform
// }

// func NewDevice(deviceType, devicePlatform string) device {
// 	return device{
// 		Type:     stringToDeviceType(deviceType),
// 		Platform: StringToDevicePlatform(devicePlatform),
// 	}
// }

// type UserJwtSession struct {
// 	UserID         ulid.ULID
// 	RefreshTokenID ulid.ULID
// 	Device         device
// 	IsRevoked      bool
// 	ExpiresAt      time.Time
// }

// func NewUserJwtSession(userID, refreshTokenID ulid.ULID, duration time.Duration, deviceType, dedevicePlatform string) (*UserJwtSession, error) {
// 	if duration < time.Minute || duration > 24*time.Hour {
// 		return nil, ErrInvalidSessionDuration
// 	}

// 	if ulidutil.IsNil(userID) || ulidutil.IsNil(refreshTokenID) {
// 		return nil, ErrNilULID
// 	}

// 	return &UserJwtSession{
// 		RefreshTokenID: ulidutil.NewULID(),
// 		UserID:         userID,
// 		ExpiresAt:      time.Now().UTC().Add(duration),
// 		Device: device{
// 			Type:     stringToDeviceType(deviceType),
// 			Platform: StringToDevicePlatform(dedevicePlatform),
// 		},
// 	}, nil
// }

// type SessionPolicy struct {
// 	RefreshDuration time.Duration
// 	AccessDuration  time.Duration
// }
