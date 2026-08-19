# habit-tracker

Google Calendar を使って日々の習慣を記録・可視化する CLI ツール。

- GitHub の contributions グラフ風のヒートマップを CLI に表示
- 1コマンドで習慣をカレンダーに記録
- 複数の Google アカウントの切り替えに対応
- Web UI から習慣・記録を管理（`serve`、優先度低）

## コマンド

```sh
habit-tracker                        # view と同じ（全習慣を合算したヒートマップを表示）

habit-tracker view                   # 習慣を選択してヒートマップを表示
habit-tracker view <habit>           # 指定した習慣を表示

habit-tracker auth login             # Google アカウントを認証してトークンを保存
habit-tracker auth list              # 認証済みアカウントの一覧を表示
habit-tracker auth switch            # 使用するアカウントを切り替え
habit-tracker auth remove            # アカウントを削除

habit-tracker setup                  # カレンダーを選択し、習慣を登録（対話形式）

habit-tracker add                    # 登録済みの習慣を対話形式で選んで記録
habit-tracker add <habit>            # 指定した習慣を直接記録
habit-tracker add -d 60 <habit>      # 記録時間を変更（デフォルト30分。-d は習慣名より前に指定する）

habit-tracker serve                  # Web UI をブラウザで起動

habit-tracker -h                     # コマンド一覧を表示
habit-tracker <command> -h           # 各コマンドの詳細なオプションと例を表示
```

## セットアップ

### 1. Google Calendar API の認証情報を作成する

1. [GCP Console](https://console.cloud.google.com/) でプロジェクトを作成する
2. 「APIs & Services」→「Library」から **Google Calendar API** を有効化する
3. 「Credentials」→「Create Credentials」→「OAuth client ID」で **Desktop app** を選択する
   （初回利用時に同意画面の設定を求められる。User Type は「External」にし、自分自身をテストユーザーとして追加する）
4. ダウンロードした JSON を `~/.config/habit/credentials.json` に配置する
   （配置先ディレクトリは環境変数 `HABIT_CREDENTIALS_DIR` で変更可能）

> **注意**: `credentials.json` を公開リポジトリに push しないこと。

### 2. 認証してセットアップする

```sh
habit-tracker auth login   # ブラウザ経由で OAuth 認証 -> トークンを保存
habit-tracker setup        # 新規カレンダーの作成または既存カレンダーの選択、追跡する習慣の登録
```

### 3. 動作確認

```sh
habit-tracker              # 全習慣を合算したヒートマップを表示
```

## ストレージ

```text
~/.config/habit/
├── token.json       # OAuth トークン
└── config.json      # 習慣・カレンダーの設定
```

パスは環境変数で上書きできる:
- `HABIT_CONFIG_DIR` — 設定ディレクトリのパス（デフォルト: `~/.config/habit/`）
- `HABIT_CREDENTIALS_DIR` — `credentials.json` を配置するディレクトリのパス

### config.json

```json
{
  "calendar_id": "abc123@group.calendar.google.com",
  "calendar_name": "habit",
  "habits": ["workout", "reading"]
}
```

> **注意**: config.json は `setup` 経由でのみ変更すること。手動編集はサポート対象外。

- `habits` — 追跡する習慣名のリスト。引数なしで `view` を実行した場合（または `habit-tracker` を単体で実行した場合）、全習慣を合算した「All habits」のヒートマップが表示される
- 習慣ごとにカレンダーを分けたい場合は `{"calendars": ["workout", "reading"]}` のみで十分

## view の出力

GitHub の contributions グラフと同じレイアウト。列 = 週（左が過去）、行 = 曜日（日曜が上）。

```text
workout  🔥 12 day streak


Sun  □□□□□■□□■□■□□□□■□□■□■□□□□■□□■□■□□□□■□□■□■□□□
Mon  ■□□■□■□■□□■□■□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■
Tue  □■□□■□■□■□□■□□■□■□□■□■□■□□■□□■□■□□■□■□■□□■□□
Wed  □□■□□■□■□□■□■□□□■□■□□■□■□■□□■□■□□□■□■□□■□■□□
Thu  ■□□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■□■□□■□■□□□□
Fri  □■□□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■□■□□■□■□□□
Sat  □□■□□■□□■□■□□■□□□■□□■□■□□■□□□■□□■□■□□■□□□□□□
```

- ■ = 実施済み、□ = 未実施
- 対話形式の一覧で「All habits」を選ぶと、追跡している全習慣を合算したヒートマップとストリークが表示される（いずれかの習慣を記録した日はカウントされる）
- 🔥 ストリーク: 今日から遡って連続で達成した日数（今日がまだ未達成でも、昨日が達成済みなら継続中としてカウントされる）

## add の挙動

- 引数なし: 矢印キーで登録済みの習慣を対話形式で選択（一覧末尾の「New habit」を選ぶと新しい習慣名を入力できる）
- `add <habit>`: 指定した習慣名を直接記録
- `-d <minutes>`: 記録時間を直接指定（デフォルト30分。指定した場合は対話プロンプトをスキップする）。`-d` は必ず習慣名より前に指定する（`add -d 60 <habit>`）
- `-d` を省略した場合、習慣を選択した後に分数の入力を求められる（デフォルト30分。1〜1440の範囲外の値は拒否される）
- その習慣が当日すでに記録済みの場合、何も起こらない（重複防止）

## ライセンス

[MIT](../../LICENSE)
