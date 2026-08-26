# リリースとHomebrew配布の仕組み

開発者向け。エンドユーザー向けの利用方法は [README](../README.md#install) を参照。

## 全体フロー

1. `v*` 形式のタグをpushすると `.github/workflows/release.yml` が起動
2. `build` ジョブ: darwin/amd64, darwin/arm64, linux/amd64 の3バイナリをクロスビルドし、artifactとしてアップロード
3. `release` ジョブ: `build` の成果物をGitHub Releaseに添付して公開（タグ名に `-` を含む場合はprerelease扱い）
4. `update-formula` ジョブ: 公開済みReleaseのバイナリから `scripts/update-formula.sh` でsha256を計算し、`Formula/habit-tracker.rb` を再生成して `main` に直接push

`update-formula` は `release` の後続ジョブなので、Releaseに実際にアップロードされたバイナリのsha256を計算する（`build`ジョブの一時artifactではなく、公開済みの成果物を参照する）。push権限はワークフロー先頭の `permissions: contents: write` のみで足り、追加のPAT/secretは不要（同一リポジトリへのpushのため）。

## Formulaを別リポジトリ（tap）に切らない理由

Homebrewの `brew tap <user>/<repo>` という短縮記法は `https://github.com/<user>/homebrew-<repo>` に自動変換される規約になっている。これに従うなら `homebrew-habit-tracker` のような専用リポジトリが必要になる。

一方、`brew tap <user>/<repo> <url>` のように3引数目にURLを明示すれば、命名規約に関係なく任意のリポジトリをtapできる。このプロジェクトでは後者を採用し、`habit-tracker` リポジトリ自体に `Formula/habit-tracker.rb` を同梱することでリポジトリを増やさずに済ませている。

```sh
brew tap tomo-local/habit-tracker https://github.com/tomo-local/habit-tracker
```

トレードオフ: `gh` のような著名なCLIは `homebrew-core`（brew同梱の公式tap）にFormulaがあるため `brew install gh` だけで済むが、`homebrew-core` に載るには知名度・メンテ体制などの審査基準があり個人ツールは対象外。個人ツールの現実的な選択肢は「専用tapリポジトリを切る」か「本体リポジトリにFormulaを同梱する」の2択で、本プロジェクトは後者。

## tap trust について（Homebrew 6.0+）

Homebrew 6.0で導入されたtap trust機構により、未知のtapは初回 `brew install` 時に `Refusing to load formula ... from untrusted tap` で失敗する。ユーザー側で以下を一度実行する必要がある（README にも記載済み）。

```sh
brew trust tomo-local/habit-tracker
```

## Formulaのローカル検証手順

新しいFormulaや `update-formula.sh` の変更を試すときは、実際にリリースされたバイナリを使ってローカルでtap〜installまで通す。

```sh
# 構文チェック
ruby -c Formula/habit-tracker.rb

# ローカルの作業ツリーをそのままtapして動作確認
brew tap tomo-local/habit-tracker "file:///path/to/habit-tracker"
brew trust tomo-local/habit-tracker
brew install habit-tracker
habit-tracker -h

# 後片付け
brew uninstall habit-tracker
brew untap tomo-local/habit-tracker
```

注意: `file://` tapはgit管理下のcommit済み内容をcloneする。`Formula/habit-tracker.rb` を編集した直後は一旦コミットしないと反映されない。

## `scripts/update-formula.sh` を手動実行する場合

CIを待たずに手元でFormulaを再生成したいとき（例: リリース後に手動修正が必要になった場合）は、対象タグを指定して直接実行できる。`gh` コマンドの認証が必要。

```sh
scripts/update-formula.sh v0.0.1-beta.1
```

`Formula/habit-tracker.rb` が上書きされるので、差分を確認してからコミットする。
