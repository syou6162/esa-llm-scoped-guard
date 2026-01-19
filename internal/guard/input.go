package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadPostInputFromFile はJSONファイルを読み込みPostInputを返します
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

	// サイズ制限付きで読み込み（10MB+1バイト読んで超過を検出）
	const maxSize = 10 * 1024 * 1024
	limitedReader := io.LimitReader(file, maxSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// サイズ超過チェック
	if len(data) > maxSize {
		return nil, NewValidationError(ErrCodeFileSizeExceeded, "file size exceeds 10MB")
	}

	// 読み込んだデータをデコード
	var input PostInput
	decoder := json.NewDecoder(io.NopCloser(strings.NewReader(string(data))))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return nil, NewValidationError(ErrCodeJSONInvalid, fmt.Sprintf("failed to parse JSON: %v", err)).Wrap(err)
	}

	// JSON EOF確認（追加データがないことを確認）
	if decoder.More() {
		return nil, NewValidationError(ErrCodeJSONInvalid, "JSON file contains multiple values")
	}

	// 2回目のDecodeでEOFを確認
	var dummy interface{}
	if err := decoder.Decode(&dummy); err != io.EOF {
		return nil, NewValidationError(ErrCodeJSONInvalid, "JSON file contains trailing data")
	}

	return &input, nil
}
