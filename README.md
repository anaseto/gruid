**Migrated to https://codeberg.org/anaseto/gruid because of new 2FA requirement (you'll need to update your imports)**

# Gruid

[![pkg.go.dev](https://pkg.go.dev/badge/github.com/anaseto/gruid.svg)](https://pkg.go.dev/github.com/anaseto/gruid)
[![godocs.io](https://godocs.io/github.com/anaseto/gruid?status.svg)](https://godocs.io/github.com/anaseto/gruid)

The **gruid** *module* provides packages for easily building grid-based
applications in Go.  The library abstracts rendering and input for different
platforms. There are drivers available for terminal apps
([gruid-tcell](https://github.com/anaseto/gruid-tcell)), native graphical apps
([gruid-sdl](https://github.com/anaseto/gruid-sdl)) and browser apps
([gruid-js](https://github.com/anaseto/gruid-js)). The original application for
the library was creating grid-based games, but it's also well-suited for any
grid-based application.

The core **gruid** *package* uses a convenient and flexible architecture of
updating a model in response to messages strongly inspired from the
[bubbletea](https://github.com/charmbracelet/bubbletea) module for building
terminal apps, which in turn is based on the functional [Elm
Architecture](https://guide.elm-lang.org/architecture/). The architecture has
been adapted to be more idiomatic in Go in the context of grid-based
applications: less functional and more efficient.

You can find examples below in the [Examples](#examples) section.

# Overview of packages

The full documentation is linked at the top of this README. We provide here a
quick overview.

The **gruid** package defines the Model and Driver interfaces and allows to
start the “update on message then draw” main loop of an application. It also
defines a convenient and efficient slice grid structure to represent the
logical contents of the screen and manipulate them.

The **ui** package defines common UI widgets and utilities: menu/table widget,
pager, text input, label, styled text drawing facilities and replay
functionality.

The **tiles** package contains helpers for drawing fonts on images, which can
be used to manage character tiles using a Driver from either gruid-sdl or
gruid-js.

The **paths** package provides efficient implementations of some common
pathfinding algorithms that are often used in grid-based games, such as
roguelikes. You will find implementations of the A\* and
[JPS](https://en.wikipedia.org/wiki/Jump_point_search) algorithms, as well as
Dijkstra, breadth first, and connected components maps computations. See
`move.go` in the movement example in
[gruid-examples](https://github.com/anaseto/gruid-examples) for an annotated
example using JPS and the mouse.

The **rl** package provides some additional utilities commonly needed in
grid-based games such as roguelikes. The package provides an event priority
queue, two complementary field of view algorithms, map generation algorithms,
as well as vault parsing and manipulation utilities.

# Drivers

The **tcell**, **sdl**, and **js** packages in the
([gruid-tcell](https://github.com/anaseto/gruid-tcell)),
([gruid-sdl](https://github.com/anaseto/gruid-sdl)) and
([gruid-js](https://github.com/anaseto/gruid-js)) modules provide specific
rendering and input implementations satisfying gruid's package Driver
interface. The provided terminal driver only handles full-window applications.
See the README.md files in the respective repositories for specific build and
deployment instructions (gruid-sdl will require SDL2, and gruid-js will require
a bit of HTML and js).

# Examples

The [gruid-examples](https://github.com/anaseto/gruid-examples) module offers
some simple annotated examples of gruid usage.

You may want to look also into some real world examples of gruid programs:

+ [harmonist](https://github.com/anaseto/harmonist) : a stealth roguelike game.
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
generation algorithms too, though its representation of maps is different.
