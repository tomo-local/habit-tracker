# habit-tracker

Googleカレンダーを使って習慣の継続状況を管理するCLIツール。

- CLIでGitHub contributions風のheatmapを表示
- カレンダーへの習慣記録を1コマンドで追加
- 複数Googleアカウントの切り替えに対応
- Web UIで習慣・記録を管理（`serve`、低優先度）

## コマンド一覧

```sh
habit-tracker                        # view と同じ（デフォルト習慣を表示）

habit-tracker view                   # デフォルト習慣のheatmapをCLIで表示
habit-tracker view <habit名>          # 指定した習慣を表示

habit-tracker auth login             # Googleアカウントを認証してトークンを保存
habit-tracker auth list              # 認証済みアカウント一覧
habit-tracker auth switch            # アクティブアカウントを切り替え
habit-tracker auth remove            # アカウントを削除

habit-tracker setup                  # カレンダー選択・習慣登録（インタラクティブ）

habit-tracker add                    # 登録済み習慣をインタラクティブに選んで記録
habit-tracker add <habit名>           # 指定した習慣を直接記録
habit-tracker add <habit名> -d 60    # 記録時間を変更（デフォルト: 30分）

habit-tracker sync                   # Claude Codeのセッションログから作業時間を自動記録
habit-tracker sync -tag <habit名>    # tagを指定して自動記録

habit-tracker serve                  # Web UIを起動してブラウザで管理

habit-tracker -h                     # コマンド一覧を表示
habit-tracker <command> -h           # 各コマンドの詳細なオプション・使用例を表示
```

## セットアップ

### 1. Google Calendar API の認証情報を作成

1. [GCP コンソール](https://console.cloud.google.com/) でプロジェクトを作成
2. 「APIとサービス」→「ライブラリ」で **Google Calendar API** を有効化
3. 「認証情報」→「認証情報を作成」→「OAuth クライアント ID」→ 種類は **デスクトップアプリ**
   （初回は同意画面の設定を求められる。User Type は「外部」＋テストユーザーに自分を追加）
4. JSON をダウンロードして `~/.config/habit/credentials.json` として配置
   （`HABIT_CREDENTIALS_DIR` 環境変数でディレクトリを変更可能）

> **注意**: `credentials.json` を公開リポジトリに push しないこと。

### 2. 認証とセットアップ

```sh
habit-tracker auth login   # ブラウザでOAuth認証 → トークン保存
habit-tracker setup        # カレンダーを新規作成 or 既存から選択し、追跡する習慣を登録
```

### 3. 動作確認

```sh
habit-tracker              # デフォルト習慣のheatmapが表示される
```

## ストレージ

```text
~/.config/habit/
├── token.json       # OAuthトークン
└── config.json      # 習慣・カレンダー設定
```

環境変数でパスを上書き可能:
- `HABIT_CONFIG_DIR` — 設定ディレクトリのパス（デフォルト: `~/.config/habit/`）
- `HABIT_CREDENTIALS_DIR` — `credentials.json` があるディレクトリのパス

### config.json

```json
{
  "calendar_id": "abc123@group.calendar.google.com",
  "calendar_name": "habit",
  "habits": ["筋トレ", "読書"]
}
```

> **注意**: config.json は `setup` 経由でのみ変更する。手動編集は非対応。

- `habits` — 追跡する習慣名のリスト。先頭の習慣が `view` のデフォルト表示になる
- `last_synced_at` — `sync` が最後に処理したセッションログの時刻（`sync` が自動更新）
- 習慣ごとにカレンダーを分ける運用なら `{"calendars": ["筋トレ", "読書"]}` だけでよい

## view の表示

GitHub contributions グラフと同じレイアウト。列 = 週（左が過去）、行 = 曜日（上が日曜）。

```text
筋トレ  🔥 12日連続


Sun  □□□□□■□□■□■□□□□■□□■□■□□□□■□□■□■□□□□■□□■□■□□□
Mon  ■□□■□■□■□□■□■□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■
Tue  □■□□■□■□■□□■□□■□■□□■□■□■□□■□□■□■□□■□■□■□□■□□
Wed  □□■□□■□■□□■□■□□□■□■□□■□■□■□□■□■□□□■□■□□■□■□□
Thu  ■□□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■□■□□■□■□□□□
Fri  □■□□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■□■□□■□■□□□
Sat  □□■□□■□□■□■□□■□□□■□□■□■□□■□□□■□□■□■□□■□□□□□□
```

- ■ = 実施日、□ = 未実施
- 🔥 連続日数: 今日から遡って連続実施している日数（今日未実施でも昨日まで続いていれば継続扱い）

## add の仕様

- 引数なし: 登録済み習慣を矢印キーで選択（一覧末尾の「New habit」で新規名を入力可）
- `add <habit名>`: 名前を直接指定して記録
- `-d <分>`: 記録する時間を直接指定（デフォルト: 30分、指定時は分数入力をスキップ）
- `-d` 未指定時は habit 選択後に記録時間（分）を数値入力（デフォルト: 30分）
- 同じ日に同じ習慣が記録済みなら何もしない（重複防止）

## sync の仕様

Claude Codeとのセッションログ（`~/.claude/projects/<カレントディレクトリのスラッグ>/*.jsonl`）を読み、作業時間をカレンダーへ自動登録する。

- メッセージの間隔が15分以上空いたら別イベントとして分割
- 各イベントの説明欄に、そのブロック中に送った実際のプロンプトを箇条書きで記録
- `-tag <habit名>`: 記録先のhabit名を指定（省略時は矢印キーで選択、新規入力も可）
- 最終同期時刻を `config.json` の `last_synced_at` に保存し、次回以降は差分のみ処理（二重登録防止）
