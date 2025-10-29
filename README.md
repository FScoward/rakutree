# Rakutree

Git worktree管理を楽にするTUIツール

## 機能

- **Worktree一覧表示**: すべてのworktreeを見やすく表示
- **スマートパス提案**: 既存worktreeから学習したパターンで自動提案
- **Worktree追加**: ブランチを選択してworktreeを簡単に作成
- **Worktree削除**: 不要なworktreeを選択して削除
- **対話的なUI**: 矢印キーで操作できる直感的なインターフェース

## インストール

```bash
go install github.com/FScoward/rakutree/cmd/rtr@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/FScoward/rakutree.git
cd rakutree
go build -o rtr ./cmd/rtr
```

## 使い方

gitリポジトリ内で実行:

```bash
rtr
```

### 操作方法

- `↑/↓` または `j/k`: カーソル移動
- `Enter`: 選択
- `ESC`: 戻る
- `q`: 終了（メインメニューから）

### 機能詳細

#### Worktree一覧表示
現在のリポジトリのすべてのworktreeを表示します。各worktreeのパス、ブランチ名、コミットハッシュが確認できます。

#### Worktree追加
1. ブランチを選択
2. パス候補から選択（既存worktreeのパターンから学習した提案が表示されます）
   - 学習したパターンに基づく提案
   - デフォルトパターン（`../ブランチ名`、`../worktrees/ブランチ名` など）
   - カスタムパス入力
3. Enterで作成

**スマートパス提案の仕組み**:
- 既存のworktreeのパスを分析してパターンを検出
- 例: `../feature-foo`、`../feature-bar` → 新しいブランチに対して `../feature-baz` を提案
- 使用頻度の高いパターンを優先的に表示
- 初回利用時はデフォルトパターンを提案

#### Worktree削除
1. 削除したいworktreeを選択
2. Enterで削除実行

## 技術スタック

- [Go](https://golang.org/)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUIフレームワーク
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUIコンポーネント
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - スタイリング

## プロジェクト構造

```
rakutree/
├── cmd/rtr/           # メインエントリーポイント
├── internal/
│   ├── git/           # git worktree操作
│   └── tui/           # TUI実装
├── go.mod
└── README.md
```

## ライセンス

MIT

## 貢献

プルリクエストを歓迎します！
