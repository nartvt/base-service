package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	ContextTimeout     = 30 * time.Second
	DefaultTierNewUser = 1
)

func StringToInt64(s string) int64 {
	fmt.Println(s)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found. Using exists environment variables")
	}
}

func SliceToJson(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	resp, err := json.Marshal(slice)
	if err != nil {
		return ""
	}
	return string(resp)
}

func StructToJson(structAny any) string {
	resp, err := json.Marshal(structAny)
	if err != nil {
		return ""
	}
	return string(resp)
}

func UUID() string {
	id := uuid.New()
	return id.String()
}

func UUIDFunc() func() string {
	return func() string {
		return UUID()
	}
}

func ContextwithTimeout() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	go cancelContext(ContextTimeout, cancel)
	return ctx
}

func cancelContext(timeout time.Duration, cancel context.CancelFunc) {
	time.Sleep(timeout)
	cancel()
	slog.Info("context canceled by timeout")
}

func ToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint:
		return int64(v)
	case uint64:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return i
		}
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
	default:
		return 0
	}
	return 0
}

func ToTimePtr(value interface{}) *time.Time {
	switch v := value.(type) {
	case time.Time:
		return &v
	case *time.Time:
		return v
	case string:
		parsedTime, err := time.Parse("2006-01-02 15:04:05", v)
		if err == nil {
			return &parsedTime
		}
	}
	return nil
}

func Format(name string) string {
	if p := strings.Split(name, "."); len(p) > 1 {
		return p[len(p)-1]
	}
	return ""
}

func GetStringOrEmpty(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func GetFileUrl(fileUrl, bucketName, folderPath, fileName string) string {
	if fileName == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s%s%s", fileUrl, bucketName, folderPath, fileName)
}
