package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// DecodeYAMLSecure は信頼されないYAMLを安全にデコードする
func DecodeYAMLSecure(r io.Reader, v interface{}, maxSize int64) error {
	// サイズ制限のデフォルト値設定
	if maxSize <= 0 {
		maxSize = MaxInputSize
	}

	// サイズ制限付き読み込み
	limitedReader := io.LimitReader(r, maxSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return NewValidationError(ErrCodeYAMLInvalid, fmt.Sprintf("failed to read input: %v", err))
	}

	// サイズ超過チェック
	if int64(len(data)) > maxSize {
		return NewValidationError(ErrCodeFileSizeExceeded, "input exceeds size limit")
	}

	// 空入力を拒否
	trimmed := bytes.TrimLeft(data, " \t\n\r")
	if len(trimmed) == 0 {
		return NewValidationError(ErrCodeYAMLInvalid, "empty input")
	}

	// yaml.NewDecoderでNodeにデコード
	nodeDecoder := yaml.NewDecoder(bytes.NewReader(data))
	var node yaml.Node
	if err := nodeDecoder.Decode(&node); err != nil {
		if err == io.EOF {
			// コメントのみのYAMLは空ドキュメント扱い
			return NewValidationError(ErrCodeYAMLInvalid, "YAML contains no document (only comments/whitespace)")
		}
		return NewValidationError(ErrCodeYAMLInvalid, fmt.Sprintf("failed to parse YAML: %v", err))
	}

	// 空ドキュメント（Kind==0）を拒否
	if node.Kind == 0 {
		return NewValidationError(ErrCodeYAMLInvalid, "YAML document is empty")
	}

	// 複数ドキュメントチェック
	var dummy yaml.Node
	err = nodeDecoder.Decode(&dummy)
	if err == nil {
		// 2つ目のドキュメントが正常にデコードできた = 複数ドキュメント
		return NewValidationError(ErrCodeYAMLInvalid, "YAML file contains multiple documents")
	} else if err != io.EOF {
		// EOF以外のエラーは末尾の不正データ
		return NewValidationError(ErrCodeYAMLInvalid, fmt.Sprintf("trailing data after YAML document: %v", err))
	}
	// EOF = 正常（単一ドキュメント）

	// ノードをウォークして検証（エイリアス展開前に実行）
	if err := validateYAMLNode(&node); err != nil {
		return err
	}

	// KnownFields(true) で構造体にデコード
	// 重要: nodeDecoderは既に消費済みなので、新しいReaderで新しいDecoderを作成する必要がある
	structDecoder := yaml.NewDecoder(bytes.NewReader(data))
	structDecoder.KnownFields(true)
	if err := structDecoder.Decode(v); err != nil {
		return NewValidationError(ErrCodeYAMLInvalid, fmt.Sprintf("failed to decode YAML: %v", err))
	}
	return nil
}

// 定数
const (
	MaxYAMLDepth = 50
	MaxYAMLNodes = 10000
)

func validateYAMLNode(node *yaml.Node) error {
	nodeCount := 0
	if err := validateYAMLNodeRecursive(node, 0, &nodeCount); err != nil {
		// 全てのエラーをErrCodeYAMLInvalidでラップ
		return NewValidationError(ErrCodeYAMLInvalid, err.Error())
	}
	return nil
}

func validateYAMLNodeRecursive(node *yaml.Node, depth int, nodeCount *int) error {
	// 深さ制限
	if depth > MaxYAMLDepth {
		return errors.New("YAML nesting too deep")
	}

	// ノード数制限
	(*nodeCount)++
	if *nodeCount > MaxYAMLNodes {
		return errors.New("YAML has too many nodes")
	}

	// エイリアスノードを拒否（展開前に検出）
	if node.Kind == yaml.AliasNode {
		return errors.New("YAML aliases are not allowed")
	}

	// アンカーを拒否（エイリアスがなくてもアンカー自体を拒否）
	if node.Anchor != "" {
		return errors.New("YAML anchors are not allowed")
	}

	// フロースタイルを拒否（ビットフラグなので&でチェック）
	// 例外: 空のシーケンス/マッピングはyaml.Marshalがフロースタイルで出力するため許可
	// ただし、空でもタグ/アンカーは既に上でチェック済み
	if node.Style&yaml.FlowStyle != 0 {
		isEmpty := len(node.Content) == 0
		if !isEmpty {
			return errors.New("YAML flow style is not allowed")
		}
		// 空でもタグ/アンカーは既に上でチェック済み
	}

	// カスタムタグを拒否（標準タグ以外）
	if node.Tag != "" && !isStandardYAMLTag(node.Tag) {
		return fmt.Errorf("custom YAML tags are not allowed: %s", node.Tag)
	}

	// マップノードの検証
	if node.Kind == yaml.MappingNode {
		// マッピングは偶数長のContentを持つべき（キー/値ペア）
		if len(node.Content)%2 != 0 {
			return errors.New("malformed YAML mapping: odd number of elements")
		}

		keys := make(map[string]bool)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]

			// 非スカラーキーを拒否（fail closed）
			if keyNode.Kind != yaml.ScalarNode {
				return errors.New("non-scalar mapping keys are not allowed")
			}

			// マージキー（<<）を拒否
			if keyNode.Value == "<<" {
				return errors.New("YAML merge keys (<<) are not allowed")
			}

			// 重複キーをチェック
			if keys[keyNode.Value] {
				return fmt.Errorf("duplicate key: %s", keyNode.Value)
			}
			keys[keyNode.Value] = true
		}
	}

	// 子ノードを再帰的に検証
	for _, child := range node.Content {
		if err := validateYAMLNodeRecursive(child, depth+1, nodeCount); err != nil {
			return err
		}
	}

	return nil
}

// isStandardYAMLTag は標準YAMLタグかどうかを判定
// yaml.v3は短縮表記（!!str）またはロング形式（tag:yaml.org,2002:str）を返す
// 注: types.goにmapフィールドやtime.Timeフィールドはないため、KnownFields(true)で十分
// 注: YAMLの型推論（on/offがboolになる等）は構造体のフィールド型で強制されるため問題なし
//
//	validator.goでの追加の型検証は不要
var standardYAMLTags = map[string]bool{
	"": true, // 暗黙的タグ
	// 短縮表記
	"!!str":       true,
	"!!int":       true,
	"!!float":     true,
	"!!bool":      true,
	"!!null":      true,
	"!!map":       true,
	"!!seq":       true,
	"!!timestamp": true,
	// ロング形式
	"tag:yaml.org,2002:str":       true,
	"tag:yaml.org,2002:int":       true,
	"tag:yaml.org,2002:float":     true,
	"tag:yaml.org,2002:bool":      true,
	"tag:yaml.org,2002:null":      true,
	"tag:yaml.org,2002:map":       true,
	"tag:yaml.org,2002:seq":       true,
	"tag:yaml.org,2002:timestamp": true,
}

func isStandardYAMLTag(tag string) bool {
	return standardYAMLTags[tag]
}
