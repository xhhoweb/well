package util

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// GenerateRandomString Generate random string with specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// StrToInt64 Convert string to int64
func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StrToInt Convert string to int
func StrToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// Int64ToStr Convert int64 to string
func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

// IntToStr Convert int to string
func IntToStr(i int) string {
	return strconv.Itoa(i)
}

// UnixToTime Convert unix timestamp to time.Time
func UnixToTime(ts int64) time.Time {
	return time.Unix(int64(ts), 0)
}

// TimeToUnix Convert time.Time to unix timestamp
func TimeToUnix(t time.Time) int64 {
	return t.Unix()
}

// DefaultIfEmpty Return default value if string is empty
func DefaultIfEmpty(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}
