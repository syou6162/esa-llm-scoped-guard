# esa-llm-scoped-guard

Claude CodeなどのLLMエージェントが、esa.ioの**特定カテゴリ配下の記事のみ**を安全に編集できるようにするGo製CLIツール。

## 特徴

- **書き込み専用ツール**: 読み取りはesa MCPサーバーに任せ、書き込みのみを制限
- **カテゴリベースの権限管理**: 許可されたカテゴリ配下のみ編集可能
- **包括的なセキュリティ対策**: 境界チェック、パストラバーサル防止、サイズ制限など
- **日本語カテゴリ対応**: Unicode文字を含むカテゴリ名をサポート

## インストール

```bash
go install github.com/syou6162/esa-llm-scoped-guard@latest
```

## 設定

### 1. 設定ファイルの作成

`~/.config/esa-llm-scoped-guard/config.yaml` を作成します：

```yaml
esa:
  team_name: "my-team"

allowed_categories:
  - "LLM/Tasks"
  - "Draft/AI-Generated"
```

**重要**: 設定ファイルとディレクトリのパーミッションを `0600` (ファイル) / `0700` (ディレクトリ) に設定してください：

```bash
chmod 600 ~/.config/esa-llm-scoped-guard/config.yaml
chmod 700 ~/.config/esa-llm-scoped-guard
```

### 2. 環境変数の設定

```bash
export ESA_ACCESS_TOKEN="your-esa-access-token"
```

## 使い方

### JSONファイルの作成

```json
{
  "name": "タスク: データ分析の実装",
  "category": "LLM/Tasks/2025-01",
  "body_md": "## 概要\n\nこのタスクでは...\n\n## 進捗\n\n- [x] 要件定義\n- [ ] 実装\n- [ ] テスト"
}
```

**注意**: タグは自動的にGitリポジトリ名が設定されます（gitリポジトリでない場合はタグなし）。

**フィールド仕様**:

| フィールド | 必須 | 説明 | 制限 |
|-----------|------|------|------|
| `post_number` | 更新時のみ | esa記事番号 | 1以上の整数 |
| `name` | Yes | 記事タイトル | 最大255バイト、制御文字不可 |
| `category` | Yes | カテゴリパス | ASCII文字のみ、許可カテゴリのマッチが必須 |
| `body_md` | Yes | 本文（Markdown） | 最大1MB、`---`で始まる場合はエラー |

### コマンド実行

```bash
# 新規作成
esa-llm-scoped-guard -json ./tasks/new-task.json

# 更新（post_numberを指定）
esa-llm-scoped-guard -json ./tasks/update-task.json
```

### ヘルプ表示

```bash
esa-llm-scoped-guard -help
```

## セキュリティ機能

### カテゴリ制限

- **境界チェック**: `LLM/Tasks` が `LLM/Tasks-evil` にマッチしないよう `/` 区切りで検証
- **パス正規化**: `..`、空セグメント、先頭/末尾スラッシュ、連続スラッシュを拒否
- **ASCII-only**: Unicode confusables回避のため、カテゴリは `[A-Za-z0-9/_-]` のみ許可

### 更新時の検証

- **TOCTOU対策**: 更新直前に既存記事を再取得してカテゴリ検証
- **カテゴリホッピング防止**: 既存カテゴリと入力カテゴリが一致しない場合は拒否

### ファイル読み込み

- **サイズ制限**: `io.LimitedReader` で10MB上限を強制
- **未知フィールド拒否**: `DisallowUnknownFields()` でスキーマ外のフィールドを拒否
- **EOF確認**: JSON終端後の追加データを検出

### API通信

- **HTTPS強制**: TLS 1.2以上を要求
- **リダイレクト禁止**: ホスト変更やHTTPダウングレードを防止
- **プロキシ無効化**: トークン漏洩防止のため `HTTP_PROXY`/`HTTPS_PROXY` を無視
- **エラーサニタイズ**: エラーメッセージを最大500文字に制限、制御文字を除去

## 開発

### テスト実行

```bash
go test ./...
```

### ビルド

```bash
go build -v ./...
```

### pre-commit設定

```bash
pre-commit install
pre-commit run --all-files
```

## ライセンス

MIT License
