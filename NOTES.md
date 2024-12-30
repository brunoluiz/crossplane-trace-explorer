# Notes

These are general notes about development around bubbletea and things that I still need to do / think about:

1. It seems `bubbletea` apps do not work well with go routines. When using error groups together with it to
implement the watcher, the `signal.NotifyContext` got affected by it since the app hijacks the input and `ctrl+c`
can't be handled correctly. Since it does not capture the keys, it never emits the SIGINT.
The hack around it was to call the `cancel` when the `tea.Quit` happens and handle `ctrl+c` within the tea app.
  - Fix was released in more recent versions with the introduction of `tea.Interrupt`

2. base16 colors can be used as a way to keep the app colours the same in any machine. See `tui` package.

3. Crossplane does not sadly expose its internals. Everything is in `internal/`.
