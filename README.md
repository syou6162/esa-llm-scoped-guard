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
  "create_new": true,
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
        "summary": ["データ分析の要件を整理", "必要なデータソースを特定"],
        "description": "データ分析の要件を定義する"
      },
      {
        "id": "task-2",
        "title": "実装",
        "status": "in_progress",
        "summary": ["データ収集APIを実装中"],
        "description": "データ分析機能を実装する"
      },
      {
        "id": "task-3",
        "title": "テスト",
        "status": "not_started",
        "summary": ["単体テストと統合テストを実施"],
        "description": "実装した機能をテストする"
      }
    ]
  }
}
```

**注意**: タグは自動的にGitリポジトリ名が設定されます（gitリポジトリでない場合はタグなし）。

### 生成されるマークダウン

上記のJSONから以下のマークダウンが生成されます：

```markdown
## サマリー
- [x] 要件定義
- [ ] 実装
- [ ] テスト

## 背景
関連リンク:
- https://github.com/owner/repo/issues/123

このタスクではデータ分析機能を実装します。

## タスク

### 要件定義
- Status: `completed`

- 要約:
  - データ分析の要件を整理
  - 必要なデータソースを特定

<details><summary>詳細を開く</summary>

データ分析の要件を定義する

</details>

### 実装
- Status: `in_progress`

- 要約:
  - データ収集APIを実装中

<details><summary>詳細を開く</summary>

データ分析機能を実装する

</details>

### テスト
- Status: `not_started`

- 要約:
  - 単体テストと統合テストを実施

<details><summary>詳細を開く</summary>

実装した機能をテストする

</details>
```

**重要**:
- `description`フィールドには、タイトルやステータス情報を含めないでください。これらは自動的に生成されます。
- `background`と`description`には、行頭に見出しマーカー（`#`など）を含めることができません。`background`は`#`、`##`が禁止、`description`は`#`、`##`、`###`が禁止です（`####`以下は使用可能）。

**フィールド仕様**:

| フィールド | 必須 | 説明 | 制限 |
|-----------|------|------|------|
| `create_new` | No | 新規作成フラグ（**trueで新規作成。post_numberと同時指定不可**） | boolean |
| `post_number` | No | esa記事番号（**既存記事の更新時に指定。create_newと同時指定不可**） | 1以上の整数 |
| `name` | Yes | 記事タイトル | 最大255バイト、制御文字・`/`・全角括弧`（）`・全角コロン`：`不可 |
| `category` | Yes | カテゴリパス | 許可カテゴリ配下で、必ず`/yyyy/mm/dd`形式の日付で終わること（例: `LLM/Tasks/2025/01/18`） |
| `body` | Yes | 本文（構造化形式） | backgroundフィールド必須、tasksフィールド必須、related_links配列は任意（URI形式） |
| `body.background` | Yes | 背景説明（プレーンテキスト） | 「## 背景」ヘッダーは含めない（自動追加される）。行頭に`#`または`##`を含めることはできない（`####`以下は可） |
| `body.related_links` | No | 関連リンク配列 | URI形式の文字列配列 |
| `body.tasks` | Yes | タスク配列 | Task配列（最低1つ必要） |

**Taskフィールド仕様**:

| フィールド | 必須 | 説明 | 制限 |
|-----------|------|------|------|
| `id` | Yes | タスクの一意識別子 | ユニークである必要あり |
| `title` | Yes | タスクのタイトル | マークダウンで「### {title}」として自動生成される |
| `status` | Yes | タスクのステータス | `not_started`, `in_progress`, `in_review`, `completed` のいずれか。マークダウンで「Status: {status}」として自動生成される |
| `summary` | Yes | タスクの要約 | 1-3行の配列。各行は140字以内。マークダウンで「- 要約:」セクションとして出力される |
| `description` | Yes | タスクの詳細説明 | プレーンテキスト。`<details><summary>詳細を開く</summary>`で囲まれて折りたたみ可能になる。行頭に`#`、`##`、`###`を含めることはできない（`####`以下は可） |
| `github_urls` | No | GitHub PR/IssueのURL配列 | `https://github.com/...`形式のURL。単数の場合「Pull Request: URL」、複数の場合「Pull Requests:」+リスト形式で出力 |

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
