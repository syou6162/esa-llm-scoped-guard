package guard

// ValidationErrorCode はバリデーションエラーの種類を表す
type ValidationErrorCode string

const (
	// Category errors
	ErrCodeCategoryEmpty             ValidationErrorCode = "category_empty"
	ErrCodeCategoryInvalidPath       ValidationErrorCode = "category_invalid_path"
	ErrCodeCategoryNotAllowed        ValidationErrorCode = "category_not_allowed"
	ErrCodeCategoryChangeNotAllowed  ValidationErrorCode = "category_change_not_allowed"
	ErrCodeCategoryInvalidDateSuffix ValidationErrorCode = "category_invalid_date_suffix"

	// Field errors
	ErrCodeFieldEmpty         ValidationErrorCode = "field_empty"
	ErrCodeFieldTooLong       ValidationErrorCode = "field_too_long"
	ErrCodeFieldInvalidChars  ValidationErrorCode = "field_invalid_chars"
	ErrCodeFieldInvalidFormat ValidationErrorCode = "field_invalid_format"

	// Reference errors
	ErrCodeDuplicateID        ValidationErrorCode = "duplicate_id"
	ErrCodeNonExistentRef     ValidationErrorCode = "non_existent_ref"
	ErrCodeSelfReference      ValidationErrorCode = "self_reference"
	ErrCodeCircularDependency ValidationErrorCode = "circular_dependency"

	// Input errors
	ErrCodeMutuallyExclusive ValidationErrorCode = "mutually_exclusive"
	ErrCodeMissingRequired   ValidationErrorCode = "missing_required"
	ErrCodeInvalidValue      ValidationErrorCode = "invalid_value"

	// File errors
	ErrCodeFileSizeExceeded ValidationErrorCode = "file_size_exceeded"
	ErrCodeNotRegularFile   ValidationErrorCode = "not_regular_file"
	ErrCodeJSONInvalid      ValidationErrorCode = "json_invalid"
)

// ValidationError はバリデーションエラーを表す構造体
// フィールドはunexportedで外部から変更不可能（Go標準ライブラリと同じパターン）
type ValidationError struct {
	code    ValidationErrorCode
	field   string
	index   int
	message string
	err     error
}

// Getters
func (e *ValidationError) Code() ValidationErrorCode { return e.code }
func (e *ValidationError) Field() string             { return e.field }
func (e *ValidationError) Index() int                { return e.index }
func (e *ValidationError) Message() string           { return e.message }
func (e *ValidationError) Unwrap() error             { return e.err }

// Error は既存メッセージを返す（後方互換）
func (e *ValidationError) Error() string {
	return e.message
}

// Is はエラーコード（code）のみで比較
func (e *ValidationError) Is(target error) bool {
	t, ok := target.(*ValidationError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// NewValidationError は新しいValidationErrorを作成（index=-1で初期化）
func NewValidationError(code ValidationErrorCode, message string) *ValidationError {
	return &ValidationError{
		code:    code,
		index:   -1,
		message: message,
	}
}

// WithField はフィールド名を設定した新しいコピーを返す（不変）
func (e ValidationError) WithField(field string) *ValidationError {
	e.field = field
	return &e
}

// WithIndex は配列インデックスを設定した新しいコピーを返す（不変）
func (e ValidationError) WithIndex(index int) *ValidationError {
	e.index = index
	return &e
}

// Wrap は元のエラーをラップした新しいコピーを返す（不変）
func (e ValidationError) Wrap(err error) *ValidationError {
	e.err = err
	return &e
}

// センチネルエラー（比較用のみ・直接returnしない）
// errors.Is() で種類を判定するための比較用インスタンス
// フィールドがunexportedなので外部から変更不可能
//
// 重要: センチネルは message が空なので、直接 return しないこと。
// 必ず NewValidationError() で新規作成して返す。センチネルは errors.Is() での比較専用。
var (
	// Category errors
	ErrCategoryEmpty             = &ValidationError{code: ErrCodeCategoryEmpty, index: -1}
	ErrCategoryInvalidPath       = &ValidationError{code: ErrCodeCategoryInvalidPath, index: -1}
	ErrCategoryNotAllowed        = &ValidationError{code: ErrCodeCategoryNotAllowed, index: -1}
	ErrCategoryChangeNotAllowed  = &ValidationError{code: ErrCodeCategoryChangeNotAllowed, index: -1}
	ErrCategoryInvalidDateSuffix = &ValidationError{code: ErrCodeCategoryInvalidDateSuffix, index: -1}

	// Field errors
	ErrFieldEmpty         = &ValidationError{code: ErrCodeFieldEmpty, index: -1}
	ErrFieldTooLong       = &ValidationError{code: ErrCodeFieldTooLong, index: -1}
	ErrFieldInvalidChars  = &ValidationError{code: ErrCodeFieldInvalidChars, index: -1}
	ErrFieldInvalidFormat = &ValidationError{code: ErrCodeFieldInvalidFormat, index: -1}

	// Reference errors
	ErrDuplicateID        = &ValidationError{code: ErrCodeDuplicateID, index: -1}
	ErrNonExistentRef     = &ValidationError{code: ErrCodeNonExistentRef, index: -1}
	ErrSelfReference      = &ValidationError{code: ErrCodeSelfReference, index: -1}
	ErrCircularDependency = &ValidationError{code: ErrCodeCircularDependency, index: -1}

	// Input errors
	ErrMutuallyExclusive = &ValidationError{code: ErrCodeMutuallyExclusive, index: -1}
	ErrMissingRequired   = &ValidationError{code: ErrCodeMissingRequired, index: -1}
	ErrInvalidValue      = &ValidationError{code: ErrCodeInvalidValue, index: -1}

	// File errors
	ErrFileSizeExceeded = &ValidationError{code: ErrCodeFileSizeExceeded, index: -1}
	ErrNotRegularFile   = &ValidationError{code: ErrCodeNotRegularFile, index: -1}
	ErrJSONInvalid      = &ValidationError{code: ErrCodeJSONInvalid, index: -1}
)
