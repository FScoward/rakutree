# Rakutree

Git worktree管理を楽にするTUIツール

## 機能

- **Worktree一覧表示**: すべてのworktreeを見やすく表示
- **Worktree追加**: ブランチを選択してworktreeを簡単に作成
- **Worktree削除**: 不要なworktreeを選択して削除
- **対話的なUI**: 矢印キーで操作できる直感的なインターフェース

## インストール

```bash
go install github.com/FScoward/rakutree/cmd/rakutree@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/FScoward/rakutree.git
cd rakutree
go build -o rakutree ./cmd/rakutree
```

## 使い方

gitリポジトリ内で実行:

```bash
rakutree
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
2. worktreeのパスを入力（省略すると `../<ブランチ名>` が使用されます）
3. Enterで作成

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
├── cmd/rakutree/       # メインエントリーポイント
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
