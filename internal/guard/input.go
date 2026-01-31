package guard

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadPostInputFromFile はYAMLファイルを読み込みPostInputを返します
func ReadPostInputFromFile(path string) (*PostInput, error) {
	// 相対パスをcwdから解決
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// symlinkを解決
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve symlink: %w", err)
	}

	// ファイルを開く
	file, err := os.Open(realPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 開いたFDに対してFstat（TOCTOU対策）
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// 通常ファイルかチェック
	if !fileInfo.Mode().IsRegular() {
		return nil, NewValidationError(ErrCodeNotRegularFile, fmt.Sprintf("file is not a regular file: %s", realPath))
	}

	// DecodeYAMLSecureでデコード（サイズチェックも含む）
	var input PostInput
	if err := DecodeYAMLSecure(file, &input, 0); err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	return &input, nil
}
