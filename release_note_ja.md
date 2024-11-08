- go-readline-ny v1.6.1 の修正を反映: UNIX系デスクトップのターミナルの機能で、複数行を貼り付けた時、一部のテキストが消える不具合を修正

v0.16.1
=======
Nov 7, 2024

- プロンプトに `\n` が含まれていても描画が乱れないようにした
  ( プロンプトの最後の行以外は、編集行の1行目前でしか表示されないようにした )

v0.16.0
=======
Nov 4, 2024

- 一番下の行において go-readline-ny v1.6.0 の予想入力が効くようにした

v0.15.0
=======
Jun 10, 2024

- v0.14.0 ではヒストリ参照前の変更テキストが維持されるのが最新エントリだけだったが、全エントリに対して入力が確定されるまで保存するようにした

v0.14.0
========
May 27, 2024

- Ctrl-P/N を入力する前のテキストを、ヒストリの最新エントリ扱いにして、失なわれないようにした。

v0.13.0
=======
May 18, 2024

- Ctrl-P と Ctrl-N の機能を修正
    - **Ctrl-P** :
        カーソルが先頭行にあるとき、ヒストリ上バックして、前の行セットを取得する。
        先頭行ではない時は前の行へカーソルを移動する。
    - **Ctrl-N** :
        カーソルが最終行にあるとき、ヒストリ上前進して、次の行セットを取得する。
        最終行ではない時は次の行へカーソルを移動する。

v0.12.1
=======
Sep 30, 2023

- go-readline-ny v0.15.0 でビルドできなくなっていた問題を修正。
    - `(*Editor) LineFeed` のかわりに `(*Editor) LineFeedWriter` を使用するようにした。

v0.12.0
=======
Sep 4, 2023

- 空行の後に空行を挿入した時に、変更フラグが立たなかった不具合を修正
- `(*Editor) SubmitOnEnterWhen` を実装： Enter だけを入力した時にそれを入力終結と判断する条件を指定できます。

次の例では、最後の行の行末がセミコロンで終わっている場合、Enter キーで入力エンドとみなします。
（第一引数は全行、第二引数はカーソルの行位置です）

```go
var ed multiline.Editor

ed.SubmitOnEnterWhen(func(lines []string, _ int) bool {
		return strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), ";")
})
```

v0.11.2
=======
Aug 17, 2023

- 次のメソッドを改名して公開
   - CmdYank, CmdNextHistory, CmdPreviousHistory, CmdDeleteChar,  
    CmdDeleteChar, CmdBackwardDeleteChar, CmdForwardChar,  
    CmdBackwardChar, CmdPreviousChar, CmdNextLine  
- 行数が１未満の時、メソッド GotoEndLine が期待どおり動かない不具合を修正
- メソッド storeCurrentLine を Sync に改名（メソッドを公開）
- Ctrl-D で空行を削除した時に画面が乱れる不具合を修正


v0.11.1
=======
May 29, 2023

- キーのアサインの追加
    - ESC p : 過去方向へヒストリ参照
    - ESC n : 未来方向へヒストリ参照
- 新メソッドの実装
    - `(*Editor) Lines`
    - `(*Editor) GotoEndLine`
    - `(*Editor) CursorLine`

v0.11.0
=======
May 19, 2023

- タブ文字を `^I` の代わりに４桁ごとの位置までの空白で表現できるようになった。
- カレント行以外の行ヘッダが本来の長さより長くカウントされていて、本行の表示がより短い場所でカットされていた問題を修正

v0.10.0
=======
May 18, 2023

- 端末の最大行数よりも多い行でも画面を乱さずに編集できるようになった。

v0.9.1
=======
May 17, 2023

- １行で`SetDefault` が呼ばれ、`SetMoveEnd` が false で呼ばれるか呼ばれなかった時、行が二重に表示されていた不具合を修正

v0.9.0
=======
May 16, 2023

初期状態のテキストとカーソルの位置を指定するメソッド `(*Editor) SetDefault` と `(*Editor) SetMoveEnd` を実装した（ [example-default.go] 参照）

[example-default.go]: ./examples/example-default.go

v0.8.0
=======
May 9, 2023

- 指定したキーにコマンドを設定するメソッド `BindKey` を追加
- 非公開メソッド `submit` 、 `newline` を `Submit`、`NewLine` にリネームして公開
- メソッド `SwapEnter` を非推奨扱いに
- `BindKey` メソッドを使うサンプルとして、[examples/example-swap.go] を追加し、[README.md] でも引用

[examples/example-swap.go]: https://github.com/hymkor/go-multiline-ny/blob/master/examples/example-swap.go
[README.md]: https://github.com/hymkor/go-multiline-ny#readme

v0.7.2
=======
May 6, 2023

- (#1) [Windows以外で桁数と行数が取り違えられる問題][issue-one] を修正するため、[go-tty] を [v0.0.5] へ更新しました。

[issue-one]: https://github.com/hymkor/go-multiline-ny/issues/1
[go-tty]: https://github.com/mattn/go-tty
[v0.0.5]: https://github.com/mattn/go-tty/releases/tag/v0.0.5

v0.7.1
=======
May 5, 2023

+  [go-readline-ny v0.11.3] で追加された `PromptWriter` を使用するようにした

[go-readline-ny v0.11.3]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.11.3


v0.7.0
=======
May 2, 2023

### v0.7.0 (May 1, 2023)

+ 現在の行頭から、LEFT・Ctrl-B で前の行の最後にカーソルを移動できるようになった
+ 現在の行末から、RIGHT・Ctrl-F で次の行の最初へカーソルを移動できるようになった 
+ v0.6.9 で Backspace でカーソル前の文字が削除できなくなっていた不具合を修正

### v0.6.9 (Apr 28, 2023)

+ 関数 `Read()` を削除。かわりに `(*Editor) Read()` をご利用ください
+ 関数 `New()` を削除。かわりに `&Editor{}` をご利用ください
+ フィールドを `(*Editor) Prompt` を隠蔽。かわりに `(*Editor) SetPrompt` をご利用ください

### v0.6.8 ( Apr 26, 2023)

- go-readline-ny を v0.11.0 へ上げたのみ

### v0.6.7 (Apr 19, 2023)

+ １行向けインクリメンタルサーチの Ctrl-S と Ctrl-R を無効化
+ `(*Editor) SwapEnter()` を追加
+ `(*Editor) SetHistoryCycle()` を追加

### v0.6.6 (Apr 17, 2023)

+ PAGEDOWN と PAGEUP をヒストリ参照に設定
+ Editor型に4つの Setter 系メソッドを追加

v0.6.5
=======
Apr 14, 2023

- v0.6.4
    - Fix for [go-readline-ny v0.10.0](https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.10.0)
- v0.6.5
    - Fix for [go-readline-ny v0.10.1](https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.10.1)

v0.6.3
=======
Apr 10, 2023

- 複数行を Ctrl-Y でペーストすると、カーソルが期待した位置に移動しない問題を修正（絵文字も対応）
- Ctrl-C を押下したとき、Readメソッドは戻り値：`readline.CtrlC` で終了するようにした
- 現在の行以外で、色付けが機能しない不具合を修正

v0.6.2
=======
Apr 9, 2023

- v0.6.0
    - **Ctrl-Y: 複数行貼り付けに対応**
    - **制御コードがそのまま表示される不具合を修正**
    - Ctrl-P/N: カーソル位置を行末にするようにした
- v0.6.1
    - Ctrl-Y: v0.6.0 の不具合修正
    -  **画面幅を越える行が存在したとき、表示が乱れる問題に対応した（オーバーを表示しない）**
    - `New` 関数は非推奨とした。初期化でエラーになったとき、panic させるようにした
- v0.6.2
    - `New` 関数では panic させないようにした。初期化のエラーは `(Editor) Read()` にて返すようにした.

v0.5.1
=======
Apr 6, 2023

Fix: the bug that the property `Prompt` did not work

v0.5.0
=======
Apr 6, 2023

- 入力内容の破棄（ESCAPE）を実装
- ヒストリ参照（ALT-P/ALT-N/CTRL-DOWN/CTRL-UP）を実装

v0.4.0
=======
Apr 2, 2023

- 型 `MultiLine` を `Editor` へ改名
- Private だった readline.Editor のフィールドを `LineEditor` として公開
- インスタンスが初期化されていなければ、メソッド `Read` で初期化を行うようにした。
   ( New でインスタンスを作成しなくても、var で変数を宣言するだけでよくした）

-----

The old style : 
```.go
m := multiline.New()
lines,err := m.Read(context.TODO())
```

The new style:
```.go
var m multiline.Editor
lines,err := m.Read(context.TODO())
```

The new Style can decreases heap allocations.

v0.3.0
=======
Apr 1, 2023

- クラス化。コンストラクター New とメソッド Read を作成。関数 Read も引き続き利用可能
- プロンプトを変更できるようにした。
- Ctrl-L で再表示できるようにした。
- 全体をリファクタリング

v0.2.0
=======
Mar 25, 2023

- Enter and Ctrl-m can split the current line
- Delete and Backspace can join lines.

v0.1.0
=======
Mar 25, 2023

The first version
