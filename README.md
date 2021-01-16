# Gruid

[![Go Reference](https://pkg.go.dev/badge/github.com/anaseto/gruid.svg)](https://pkg.go.dev/github.com/anaseto/gruid)
[![Go Report Card](https://goreportcard.com/badge/github.com/anaseto/gruid)](https://goreportcard.com/report/github.com/anaseto/gruid)

The **gruid** *module* provides packages for easily building grid-based
applications in Go.  The library abstracts rendering and input for different
platforms. The module provides drivers for terminal apps (driver/tcell), native
graphical apps (driver/sdl) and browser apps (driver/js). The original
application for the library was creating grid-based games, but it's also
well-suited for any grid-based application.

The core **gruid** *package* uses a convenient and flexible architecture of
updating a model in response to messages strongly inspired from the
[bubbletea](https://github.com/charmbracelet/bubbletea) module for building
terminal apps, which in turn is based on the functional [Elm
Architecture](https://guide.elm-lang.org/architecture/). The architecture has
been adapted to be more idiomatic in Go in the context of grid-based
applications: less functional and more efficient.

You can find there [annotated examples](examples/) and the
[documentation](https://pkg.go.dev/github.com/anaseto/gruid).

*Note: the module is already usable and the core functionality and APIs should
be stable or close to it. It is expected that after a testing period to collect
user feedback, a stable version will be released.*

# Overview of packages

The **gruid** package defines the Model and Driver interfaces and allows to
start the “update on message then draw” main loop of an application. It also
defines a convenient and efficient slice grid structure to represent the
logical contents of the screen and manipulate them.

The **ui** package defines common UI widgets and utilities: menu/table widget,
pager, text input, label, styled text drawing facilities and replay
functionality.

The **tcell**, **js**, and **sdl** packages in the [drivers
sub-directory](drivers/) provide specific rendering and input implementations
satisfying gruid's package Driver interface. The provided terminal driver only
handles full-window applications. See the README.md files in their respective
folders for specific build and deployment instructions. *Note that until lazy
module loading comes (hopefully with Go 1.16), unless you manually remove the
sdl dependency, you will probably need to install SDL2 even if you only want to
use tcell for the terminal.*

The **tiles** package contains helpers for drawing fonts on images, which can
be used to manage character tiles using a Driver from either drivers/sdl or
drivers/js.

The **paths** package provides efficient implementations of some common
pathfinding algorithms that are often used in grid-based games, such as
roguelikes. You will find implementations of the A\* algorithm, as well as
Dijkstra, breadth first, and connected components maps computations. See the
[movement example](examples/movement/move.go) for an annotated example using
A\* and the mouse.

The **rl** package provides some additional utilities commonly needed in
grid-based games such as roguelikes. The package provides an event priority
queue, a field of view algorithm, map generation algorithms, as well as vault
parsing and manipulation utilities.

# Examples

In addition of the [annotated examples](examples/) that come within the gruid
module, you may want to look also into some real world examples of gruid
programs:

+ [gospeedr](https://github.com/anaseto/gospeedr) : a simple speed reading program.

# See also

If you need to handle wide-characters, that is characters that take two cells
in the terminal, you may want to look into
[go-runewidth](https://github.com/mattn/go-runewidth).

The [clipboard](https://github.com/atotto/clipboard) module may be of interest
for some applications too, as copying and pasting is not handled by gruid. Note
that, at this time, the clipboard module does not support the js platform, but
there's at least one fork that does.

As gruid only provides a few map generation algorithms, you may be interested
in the [dngn](https://github.com/SolarLune/dngn) module, which provides map
generation algorithms too, though its representations of maps is different.
