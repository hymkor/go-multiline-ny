- Fix: Dirty flag was not set on inserting an empty line after an empty line

v0.11.2
=======
Aug 17, 2023

- Rename and export Editor's methods:  
  - CmdYank, CmdNextHistory, CmdPreviousHistory, CmdDeleteChar,  
    CmdDeleteChar, CmdBackwardDeleteChar, CmdForwardChar,  
    CmdBackwardChar, CmdPreviousChar, CmdNextLine  
- Fix: GotoEndLine() did not work as expected when .lines[] is lower than 1
- Rename storeCurrentLine() to Sync() (new public method)
- Fix: display disorder when deleting a blank line with C-d

v0.11.1
=======
May 29, 2023

- Add key assigns
    - ESC p : previous history
    - ESC n : next history
- Implement new methods:
    - `(*Editor) Lines`
    - `(*Editor) GotoEndLine`
    - `(*Editor) CursorLine`

v0.11.0
=======
May 19, 2023

- Tab characters can now be represented by a few spaces up to every fourth position instead of ^I
- Fix: the width of the line header had not been counted incorrect

v0.10.0
=======
May 18, 2023

- More lines than the terminal's maximum ones are editable now without disturbing the screen

v0.9.1
======
May 17, 2023

- Fix: If `SetDefault` was called with only one line and `SetMoveEnd` was called with false or not called, the line was printed twice.

v0.9.0
======
May 16, 2023

- Implemented methods `(*Editor) SetDefault` and `(*Editor) SetMoveEnd` to specify initial text and cursor position. (See [example-default.go])

[example-default.go]: ./examples/example-default.go

v0.8.0
======
May 9, 2023

- Implement method `BindKey`
- Rename and exposed internal method `submit` and `newline` as `Submit` and `NewLine`
- Deprecate `SwapEnter`
- Add example: `examples/example-swap.go`

v0.7.2
======
May 6, 2023

- Update go-tty to v0.0.5 for https://github.com/hymkor/go-multiline-ny/issues/1

v0.7.1
======
May 2, 2023

- Use `PromptWriter` on [go-readline-ny v0.11.3]

[go-readline-ny v0.11.3]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.11.3

v0.7.0
=======
May 1, 2023

- Fix: Backspace did not work to delete backward character since v0.6.9
- Update for [go-readline-ny v0.11.2]
- Ctrl-B can move cursor to the end of the previous line
- Ctrl-F can move cursor to the beginning of the next line

[go-readline-ny v0.11.2]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.11.2

v0.6.9
======
Apr 28, 2023

- Fix for [go-readline-ny v0.11.1]
- Remove `Read()`. Use `(*Editor).Read()`
- Remove `New()`. Use `&Editor{}`
- Hide `(*Editor).Prompt`

[go-readline-ny v0.11.1]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.11.1

v0.6.8
======
Apr 26, 2023

- Fix for [go-readline-ny v0.11.0]

[go-readline-ny v0.11.0]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.11.0

v0.6.7
======
Apr 19, 2023

- Disable Ctrl-S and Ctrl-R (one line version incremental search)
- Add the method: `(*Editor) SwapEnter()`
- Add the method: `(*Editor) SetHistoryCycling()`

v0.6.6
======
Apr 17, 2023

- Bind PAGEDOWN and PAGEUP to refer history
- Add 4 setter methods to the type Editor

v0.6.5
======
Apr 14, 2023

- Fix for [go-readline-ny v0.10.1]

[go-readline-ny v0.10.1]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.10.1

v0.6.4
======
Apr 13, 2023

- Fix for [go-readline-ny v0.10.0]

[go-readline-ny v0.10.0]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v0.10.0

v0.6.3
======
Apr 10, 2023

- Fix: pasting emoji caused the cursor position incorrect
- Fix: Ctrl-Y: the cursor position was not moved at the expected position
- (Editor) Read(): Ctrl-C stops the current loop and returns readline.CtrlC
- Fix: Coloring did not work not on the current line

v0.6.2
======
Apr 9, 2023

- `New()` never raises panic now. Instead, `(Editor) Read()` returns an error on initializing.

v0.6.1
======
Apr 8, 2023

- Bug fix (Ctrl-Y on v0.6.0)
- Fix: screen was distorted when line longer than the screen width exists
- `New()` is marked as Deprecate. It panics when initilizing failed.

v0.6.0
======
Apr 8, 2023

- Support: Ctrl-Y: pasting multiple lines
- Fixed: the control code was displayed as it was
- Ctrl-P/N: Set cursor position the tail of the line

v0.5.0
======
Apr 6, 2023

- Implement clear (ESCAPE)
- Implement history (ALT-P/ALT-N/CTRL-DOWN/CTRL-UP)

v0.4.0
=======
Apr 2, 2023

- Rename type: `MultiLine` to `Editor`
- Made the private field `editor` (`readline.Editor`) public as `LineEditor`
- The method `Read` initializes the instance when it has not been  
  ( The instance does not have to be initalized with `New` function when it is declared with `var` )

v0.3.0
======
Apr 1, 2023

- Made a class which has a constructor New and a method Read. The function Read still available
- Enabled to change the prompt
- Repaint with Ctrl-L
- Refactored the whole code

v0.2.0
======
May 25, 2023

- Enter and Ctrl-m can split the current line
- Delete and Backspace can join lines.

v0.1.0
======
May 25, 2023

+ The first version.
