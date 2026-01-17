# esa-llm-scoped-guard

Claude CodeなどのLLMエージェントが、esa.ioの**特定カテゴリ配下の記事のみ**を安全に編集できるようにするGo製CLIツール。

## 特徴

- **書き込み専用ツール**: 読み取りはesa MCPサーバーに任せ、書き込みのみを制限
- **カテゴリベースの権限管理**: 許可されたカテゴリ配下のみ編集可能

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
  "category": "LLM/Tasks/2025/01/18",
  "body": {
    "background": "このタスクではデータ分析機能を実装します。",
    "related_links": ["https://github.com/owner/repo/issues/123"],
    "tasks": [
      {
        "id": "task-1",
        "title": "要件定義",
        "status": "completed",
        "description": "データ分析の要件を定義する"
      },
      {
        "id": "task-2",
        "title": "実装",
        "status": "in_progress",
        "description": "データ分析機能を実装する"
      },
      {
        "id": "task-3",
        "title": "テスト",
        "status": "not_started",
        "description": "実装した機能をテストする"
      }
    ]
  }
}
```

**注意**: タグは自動的にGitリポジトリ名が設定されます（gitリポジトリでない場合はタグなし）。

**フィールド仕様**:

| フィールド | 必須 | 説明 | 制限 |
|-----------|------|------|------|
| `post_number` | 更新時のみ | esa記事番号 | 1以上の整数 |
| `name` | Yes | 記事タイトル | 最大255バイト、制御文字・`/`・全角括弧`（）`・全角コロン`：`不可 |
| `category` | Yes | カテゴリパス | 許可カテゴリ配下で、必ず`/yyyy/mm/dd`形式の日付で終わること（例: `LLM/Tasks/2025/01/18`） |
| `body` | Yes | 本文（構造化形式） | backgroundフィールド必須、tasksフィールド必須、related_links配列は任意（URI形式） |
| `body.background` | Yes | 背景説明（プレーンテキスト） | 「## 背景」ヘッダーは含めない（自動追加される） |
| `body.related_links` | No | 関連リンク配列 | URI形式の文字列配列 |
| `body.tasks` | Yes | タスク配列 | Task配列（最低1つ必要） |

**Taskフィールド仕様**:

| フィールド | 必須 | 説明 | 制限 |
|-----------|------|------|------|
| `id` | Yes | タスクの一意識別子 | ユニークである必要あり |
| `title` | Yes | タスクのタイトル | - |
| `status` | Yes | タスクのステータス | `not_started`, `in_progress`, `in_review`, `completed` のいずれか |
| `description` | Yes | タスクの詳細説明 | - |

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
