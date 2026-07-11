# habit-tracker

Googleカレンダーの「習慣専用カレンダー」を元に、GitHub contributions 風の格子ビューで
習慣の継続状況(連続日数・実施回数)を表示するローカルツール。

## 仕組み

- 習慣ごとに Google カレンダーを 1 つ用意する(例: 「筋トレ」「読書」)
- そのカレンダーにイベントがある日 = その習慣を実施した日としてカウント
- `habit-tracker` を起動するとブラウザで格子ビューが開く(リロードで再取得)

## セットアップ

### 1. Google Calendar API の認証情報を作成

1. [GCP コンソール](https://console.cloud.google.com/) でプロジェクトを作成
2. 「APIとサービス」→「ライブラリ」で **Google Calendar API** を有効化
3. 「認証情報」→「認証情報を作成」→「OAuth クライアント ID」→ 種類は **デスクトップアプリ**
   (初回は同意画面の設定を求められる。User Type は「外部」+ テストユーザーに自分を追加)
4. JSON をダウンロードしてリポジトリ直下に `credentials.json` として配置
   (`~/.config/habit-tracker/credentials.json` でも可。実行ファイルの隣 → カレントディレクトリ → `~/.config` の順で探す)

> **注意**: credentials.json を含むこのリポジトリを公開リポジトリに push しないこと。

### 2. ビルドとカレンダー確認

```sh
go build -o habit-tracker .
./habit-tracker list   # 初回はブラウザで OAuth 認証 → カレンダー名一覧が出る
```

### 3. 習慣カレンダーを設定

`~/.config/habit-tracker/config.json`:

```json
{
  "calendars": ["habit"],
  "group_by_title": true,
  "habits": ["筋トレ", "読書"]
}
```

- `group_by_title: true` — 1つのカレンダー内でイベントタイトルごとに習慣を分ける
- `habits` — 正式な習慣名のリスト(誤字対策)。タイトルは正規化(空白・全角半角・大小文字)の上、
  習慣名を含めば集約される(「筋トレ30分」→「筋トレ」)。どれにも一致しないタイトルは
  独立した行として表示されるので、誤字に気づいたらカレンダー側を修正する。
  登録済みの習慣はイベントが無くても行が表示される。
- 習慣ごとにカレンダーを分ける運用なら `{"calendars": ["筋トレ", "読書"]}` だけでよい

## 使い方

```sh
./habit-tracker              # サーバー起動 + ブラウザが開く
./habit-tracker -weeks 52    # 表示期間を52週に
./habit-tracker -port 9000   # ポート変更

./habit-tracker add                            # 登録済み習慣から番号で選んで記録(複数可)
./habit-tracker add 筋トレ                     # 名前を直接指定して記録
./habit-tracker -date 2026-07-10 add 筋トレ    # 日付を指定して記録
```

`add` は同じ日に同じ習慣が記録済みなら何もしない(重複防止)。
`group_by_title` が true の場合は config の先頭カレンダーに、false の場合は習慣名と同名のカレンダーに書き込む。

- 🔥 連続日数: 今日から遡って連続実施している日数(今日未実施でも昨日まで続いていれば継続扱い)
- 格子: 列 = 週、行 = 曜日(日曜始まり)。緑 = 実施日、白枠 = 今日
