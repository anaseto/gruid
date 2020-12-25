This tcell package depends on the [tcell](https://github.com/gdamore/tcell/v2)
terminal library, which uses the permissive enough [Apache License
v2](https://github.com/gdamore/tcell/blob/master/LICENSE). The library is pure
Go, so no special steps are required to install it, and it allows for easy
cross-compilation.

You only need to look into the `tcell.Style` type documentation to define the
styling that has to be used by the Driver.

Note that the terminal grid is not a true grid: some characters are two-cell
wide (such as wide east-asian characters). The character width can be computed
thanks to the [go-runewidth](github.com/mattn/go-runewidth) package, so it
could be taken into account, but it is a bit cumbersome and ad hoc with respect
to the graphical drivers which can handle this problem more easily (the tile
width can be adjusted to fit any wanted character). Currently, runes with zero
value are ignored by this terminal driver, so they can be placed after a wide
character, but this is not portable to other drivers.
