package util

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"resty.dev/v3"
)

const (
	cookieEnvName = "TERABOX_COOKIE"
)

func FindFilesRecursive(rootPath string) ([]string, error) {
	var fileList []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path '%s': %v\n", path, err)
			return nil // Skip this path and continue walking
		}
		// Skip the root directory itself
		if path == rootPath {
			return nil
		}
		// Check for hidden files (starts with .) and skip
		if !info.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}

func GetResponse(resp *resty.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode())
	}

	return resp.Bytes(), nil
}

func getCookieFromEnv() (string, error) {
	cookie := os.Getenv(cookieEnvName)
	if cookie == "" {
		return "", fmt.Errorf("TeraBox Cookie not found")
	}
	return cookie, nil
}

func GetCookie(filePath string) (string, error) {
	if filePath == "" {
		return getCookieFromEnv()
	}

	fp, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, filePath, "not found")
		return getCookieFromEnv()
	}
	defer fp.Close()

	bytes, err := io.ReadAll(fp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read cookie file")
		return getCookieFromEnv()
	}

	cookie := string(bytes)
	return cookie, nil
}

func GetAbsPath(cwd, relPath string) string {
	if !strings.HasPrefix(relPath, "/") {
		return path.Join(cwd, relPath)
	}
	return path.Clean(relPath)
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func ParseChunkSize(arg string) (int64, error) {
	if arg == "" {
		return 0, fmt.Errorf("chunk-size cannot be empty")
	}

	numStr := arg
	unit := strings.ToLower(string(numStr[len(numStr)-1]))

	if unit < "0" || unit > "9" {
		numStr = arg[:len(arg)-1]
		if numStr == "" {
			return 0, fmt.Errorf("invalid format: %s", arg)
		}
	} else {
		unit = ""
	}

	base, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", numStr)
	}

	if base < 1 {
		return 0, fmt.Errorf("value must be greater than 0")
	}

	switch unit {
	case "k":
		base *= 1024
	case "m":
		base *= 1024 * 1024
	case "g":
		base *= 1024 * 1024 * 1024
	case "":
	default:
		return 0, fmt.Errorf("invalid unit: %s (use k/K/m/M/g/G)", unit)
	}

	base = min(1024*1024*1024*1024, base)

	return base, nil
}
