package parse

import (
	"strconv"
	"time"
)

//GetInt64 Returns an int64 number if valid, else returns 0
func GetInt64(val string) (int64, error) {
	if val == "" {
		return 0, nil
	}
	return MustGetInt64(val)
}

//MustGetInt64 returns
func MustGetInt64(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}

//GetTime returns nil if val is empty, else does a time.Parse with RFC3339 format
func GetTime(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}
	return MustGetTime(val)
}

//MustGetTime just calls time.Parse with RFC3339 format
func MustGetTime(val string) (*time.Time, error) {
	t, err := time.Parse(time.RFC3339, val)
	return &t, err
}
