- Fixed an issue where parts of the previous lines were sometimes not erased during inc
remental search.(#7)

v0.20.1
=======
Feb 9, 2025

- Use the Clipboard of `go-readline-ny.Editor` instead of the operating system’s clipboard.

v0.20.0
=======
Feb 9, 2025

- Avoid using `ESC[s` (which saves the cursor position), `ESC[u` (which restores the cursor position), and `ESC[J` (which clears lines after the cursor) when concatenating lines or referencing the history to support the macOS version of JetBrains IDE Terminal. [#7],[IJPL-60199]

Thanks to [@apstndb]

[#7]: https://github.com/hymkor/go-multiline-ny/issues/7
[IJPL-60199]: https://youtrack.jetbrains.com/issue/IJPL-60199/Console-doesnt-support-ANSI-escape-code-for-clearing

v0.19.2
=======
Jan 25, 2025

- Slightly reduced the number of calls to the pattern matching function for syntax highlighting

v0.19.1
=======
Jan 23, 2025

- Updated the dependency of go-readline-ny to v1.7.3 that improves processing time when the syntax highlighting evaluation function is slow by reducing the number of its calls

v0.19.0
=======
Jan 22, 2025

- Enhanced multi-line syntax highlighting to allow color changes across line boundaries. [#6]
- Discontinued support for the field `.LineEditor.Coloring`, which was used for the old syntax highlighting. For compatibility, `(*Editor)` SetColoring remains available
- To configure highlighting, use `Editor.{Highlight, DefaultColor, ResetColor}` instead of `Editor.LineEditor.{Highlight, DefaultColor, ResetColor}`.

Please update the code to assign values to the fields as shown below:

- `var ed multiline`
- `ed.LineEditor.Highlight = ...` to `ed.Highlight = ...`
- `ed.LineEditor.ResetColor = ...` to `ed.ResetColor = ...`
- `ed.LineEditor.DefaultColor = ...` to `ed.DefaultColor = ...`

Thanks to [@apstndb]

[#6]: https://github.com/hymkor/go-multiline-ny/issues/6

v0.18.4
=======
Jan 19, 2025

- Fix a panic caused by incomplete support for the new syntax highlighting in go-readline-ny v1.7

v0.18.3
=======
Jan 18, 2025

- Update go-readline-ny dependency to [v1.7.0](https://github.com/nyaosorg/go-readline-ny/releases/tag/v1.7.0). It introduced a new interface for syntax highlighting.
- Deprecated `(*Editor) SetColoring`

v0.18.2
=======
Jan 9, 2025

- Updated go-readline-ny dependency to [v1.6.3](https://github.com/nyaosorg/go-readline-ny/releases/tag/v1.6.3), which fixes an issue where the default color setting was incorrectly set to bold.

v0.18.1
=======
Nov 28, 2024

- Add `NewPrefixCommand` as the constructor of the structure `PrefixCommand`

v0.18.0
=======
Nov 24, 2024

- When prefix key(Escape) is pressed, echo it
- Assign Escape → Enter to submit

v0.17.0
=======
Nov 20, 2024

- Implement the incremental search (Ctrl-R) [#4]

Thanks to [@apstndb]

[@apstndb]: https://github.com/apstndb
[#4]: https://github.com/hymkor/go-multiline-ny/issues/4

v0.16.4
=======
Nov 19, 2024

- Fix: on the legacy terminal of Windows, cursor does not move to the upper line
- Fix: on the terminal of Linux/macOS desktop, backspace-key could not remove the line feed (#5,Thanks to @apstndb)

v0.16.3
=======
Nov 9, 2024

- Fix: when editing the longer lines than screen height, the number of the lines scrolling was one line short
  ( It seemed to assume the height of status line which does not exist on default, therefore make the field `.StatusLineHeight` and use it as the height of the status line. )
- Fix: with no next history entry, when trying to move to the next line at the bottom line, cursor had moved at the top line.
  ( It was the behaviour to move the top line of the next entry of the history. But, because no next entry exists, the cursor moved to the top line without changing the current entry of history )

v0.16.2
=======
Nov 9, 2024

- With modifying go-readline-ny on v1.6.1: Fix: some text was missing when pasting multi-lines using the terminal feature of Linux Desktop (#3,Thanks to @apstndb)

v0.16.1
=======
Nov 7, 2024

- Prevent from incorrect rendering when the prompt includes `\n` (All lines except the last one of the prompt are displayed only on the first line of the edit line,Thanks to @apstndb)

v0.16.0
=======
Nov 4, 2024

- The prediction of go-readline-ny v1.6.0 is available now on the bottom line

v0.15.0
=======
Jun 10, 2024

- Not only the last entry of history, but all modified entries are kept the last value until the current input is completed

v0.14.0
========
May 27, 2024

- The text before the first Ctrl-P/N is treated as if it were the latest entry in the history not to lose them

v0.13.0
=======
May 18, 2024

- Modify the features of Ctrl-P and Ctrl-N
    - **Ctrl-P** :
        When the cursor exists the top of lines,
        move BACK through the history list, fetching the previous lines.
        Otherwise move the cursor to the previous line.
    - **Ctrl-N** :
        When the cursor exists the bottom of lines,
        move FORWARD through the history list, fetching the next lines.
        Otherwise move the cursor to the next line.

v0.12.1
=======
Sep 30, 2023

- Fix: Could not build with go-readline-ny v0.15.0
    - Use `(*Editor) LineFeedWriter` instead of `LineFeed`

v0.12.0
=======
Sep 2, 2023

- Fix: Dirty flag was not set on inserting an empty line after an empty line
- Implement `(*Editor) SubmitOnEnterWhen`

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

- Update [go-tty] to [v0.0.5][go-tty_v0.0.5] for [#1]

Thanks to [@spiegel-im-spiegel]

[@spiegel-im-spiegel]: https://github.com/spiegel-im-spiegel
[#1]: https://github.com/hymkor/go-multiline-ny/issues/1
[go-tty]: https://github.com/mattn/go-tty
[go-tty_v0.0.5]: https://github.com/mattn/go-tty/releases/tag/v0.0.5

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
