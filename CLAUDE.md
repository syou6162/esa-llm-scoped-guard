# esa-llm-scoped-guard - Claude Code向けプロジェクト固有指示

## プロジェクト概要

esa.ioへの書き込みを特定カテゴリのみに制限するセキュアなGo製CLIツールです。

## 開発時の注意事項

### セキュリティ重視の設計

このツールはLLMエージェントが任意の記事を編集できないようにするセキュリティツールです。以下の原則を厳守してください：

1. **fail closed**: エラー時は常に操作を拒否する
2. **境界チェック**: カテゴリマッチは `/` 区切りで厳密に検証
3. **パス正規化**: `..`, `.`, 空セグメント、先頭/末尾スラッシュを拒否
4. **日付形式必須**: カテゴリは必ず `/yyyy/mm/dd` 形式で終わること（年: 2000-2099, 月: 01-12, 日: 01-31）
5. **サイズ制限**: 全入力に制限を設ける（10MB上限）
6. **文字制限**: nameフィールドは `/`、全角括弧 `（）`、全角コロン `：` を禁止

### コーディング規約

- **標準ライブラリ優先**: cobra/viperなどの外部ライブラリは使用せず、標準ライブラリ中心
- **テスト駆動**: 新機能追加時は必ずテストを先に書く（t_wada式TDD）
- **エラーハンドリング**: 全てのエラーに適切なコンテキストを付与

### テストの方針

- **Protocolベース**: `EsaClientInterface` を使ってモック可能な設計
- **境界値テスト**: セキュリティ機能は境界値を必ずテスト
- **統合テスト**: スタブクライアントで全体フローを検証

### コミット規約

- **Conventional Commits形式**: `feat:`, `fix:`, `test:` などを使用
- **semantic-committing**: 変更は意味のある最小単位でコミット
- **小まめにコミット**: 各フェーズ完了時にコミット

## セキュリティレビューのポイント

新機能追加時は以下を確認してください：

1. **入力検証**: 全ての入力に適切なバリデーションがあるか
2. **サイズ制限**: メモリ枯渇攻撃を防ぐ制限があるか
3. **パストラバーサル**: ファイル操作でパストラバーサルが可能でないか
4. **TOCTOU**: レースコンディションのリスクがあるか
5. **エラーメッセージ**: 機密情報が漏洩しないか
6. **日付形式検証**: categoryの日付形式が正しく検証されているか（正規表現 + 範囲チェック）

## よくある修正パターン

### カテゴリ検証の追加

```go
// 悪い例
if strings.HasPrefix(category, allowed) {
    // LLM/Tasks-evil が LLM/Tasks にマッチしてしまう
}

// 良い例
normalized, err := guard.NormalizeCategory(category)
if err != nil {
    return false, err
}
allowed, err := guard.IsAllowedCategory(normalized, allowedCategories)
if err != nil || !allowed {
    return fmt.Errorf("category not allowed")
}
```

### ファイル読み込みのセキュリティ

```go
// 悪い例
data, _ := os.ReadFile(path) // サイズ無制限

// 良い例
file, err := os.Open(realPath)
if err != nil {
    return nil, err
}
defer file.Close()

limitedReader := io.LimitReader(file, 10*1024*1024+1)
data, err := io.ReadAll(limitedReader)
if len(data) > 10*1024*1024 {
    return nil, fmt.Errorf("file size exceeds 10MB")
}
```

## LLMエージェント向け推奨ワークフロー

esa.ioへ投稿する前に、以下のワークフローを推奨します：

1. **validate**: JSONの妥当性を検証
   ```bash
   esa-llm-scoped-guard validate -json ./tasks/123.json
   ```

2. **preview**: 生成されるMarkdownを確認
   ```bash
   esa-llm-scoped-guard preview -json ./tasks/123.json
   ```
   - 意図しないHTMLタグ（`<details>`や`<summary>`など）が含まれていないか確認
   - Markdown構造が正しいか確認

3. **diff**: 既存記事の更新時は差分を確認
   ```bash
   esa-llm-scoped-guard diff -json ./tasks/123.json
   ```
   - 意図した変更になっているか検証
   - 不要な変更が含まれていないか確認

4. **post**: 最終確認後に投稿
   ```bash
   esa-llm-scoped-guard post -json ./tasks/123.json
   ```

## トラブルシューティング

### テスト失敗時

1. `go test ./... -v` で詳細を確認
2. 境界値テストが失敗している場合は正規化ロジックを確認
3. mock/stubの設定が正しいか確認

### ビルドエラー時

1. `go mod tidy` で依存関係を整理
2. `go fmt ./...` でフォーマットを統一
3. `staticcheck ./...` で静的解析を実行

### セキュリティ警告時

Codex MCPに相談して、セキュリティ専門家の視点でレビューを依頼してください。

## 関連ファイル

- `README.md`: ユーザー向けドキュメント
- `.claude_work/plans/*.md`: 実装計画書
- `internal/guard/guard.go`: カテゴリ権限チェックのコアロジック
- `internal/guard/validator.go`: 入力バリデーション
- `internal/guard/input.go`: JSON読み込みロジック
- `config.go`: 設定ファイル読み込み
