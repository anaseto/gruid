This changelog file only lists important changes.

## v0.18.0 - 2020-02-19

This release moves the drivers packages into their own modules. The rationale
behind that it that users may want to only use the terminal driver, so it is
inconvenient to force a module dependence on SDL (at least, until lazy module
loading comes). It also makes it easier to fork the provided drivers, or use
custom ones without depending on any default ones.

There are now new modules:

+ github.com/anaseto/gruid-tcell
+ github.com/anaseto/gruid-sdl
+ github.com/anaseto/gruid-js

Their packages are drop-in replacements for the old drivers/\* packages.

## v0.17.0 - 2020-02-16

+ NewVault shorthand method for creating new vaults.
+ New Bounds method for ui.Menu, that returns the range occupied currently by
  the menu.
+ New AtU method in rl.Grid package, for cases in which performance does really
  matter.
+ Improve handling of short texts in pager (texts that do not use the whole
  available space).
+ Send a MsgScreen initially in js driver, as the other drivers already did.
+ Add source too in returned slice of visibles in SSC FOV methods.

## v0.16.0 - 2020-02-09

+ Use image.Image instead of image.RGBA in sdl and js drivers for GetImage
  (minor incompatible change).

## v0.15.0 - 2020-02-09

+ New FOV algorithm: symmetric shadow casting.
+ Improvement in JPS algorithm neighbor prunning based on the 2014 paper.

## v0.14.0 - 2020-02-05

+ New JPSPath function in paths packages for really fast pathfinding on a grid
  using the JPS algorithm.
+ Some improvements in Astar and Dijkstra algorithms performance.
+ New PopR method for rl.EventQueue, that returns an
  event along with its rank
+ New PushFirst method for rl.EventQueue that pushes an event to the queue in
  LIFO position among events of same rank, instead of FIFO (Push).
+ Renamed ui.Label's StyledText field to more fitting Content: StyledText was
  too vague, as the label's title is also a StyledText (incompatible change).

## v0.13.0 - 2020-02-02

+ New From method for rl.FOV that allows to get the previous position in a
  light ray from the current source without having to compute the whole ray.
+ VisionMap and LightMap used a radius parameter: it is moved to the Lighter
  interface for a more flexible and intuitive API (incompatible change).
+ New ActiveBounds method for ui.Menu that returns the range occupied by the
  active entry.
+ Improve SetScale handling in sdl driver.

## v0.12.0 - 2020-01-25

+ Add String methods for Point, Range and Grid types in gruid package.
+ Added a couple of examples in gruid package.
+ New Footer and alignment options for ui.Box.
+ Fix display in ui.Menu last's page: now the menu only takes the necessary
  space for the items present on the page, which can be less than in previous
  pages.

## v0.11.0 - 2020-01-20

+ Grid.Iterator method now returns a value instead of a pointer (could be
  incompatible, though not in practice).
+ FrameDecoder.Decode now properly skips potential extra data that may be
  present in case encoding was done in separate writes to a same file.
+ FrameDecoder.Decode now takes a pointer to Frame (incompatible change),
  allowing control over allocations.

## v0.10.0 - 2020-01-17

+ A few changes in Menu, Pager and TextInput configuration in order to make
  styling more flexible, and simplify configuration too (some incompatible
  changes).
+ Add method versions with formatted text for StyledText.
+ Add more tests in ui package, and fix a couple of issues.

## v0.9.0 - 2020-01-12

This release brings the module quite closer to a stable release. Test coverage
is now around 70-90% for gruid, paths and rl packages.

+ Make Color and AttrMask types always be uint32, as relying on the exact size
  of uint would have been an error. This change also makes gruid.Grid use half
  memory and avoid 32 bits of padding in 64 bit systems and improves speed of
  some operations, like Copy. This change should not be incompatible in
  practice.
+ The Ray function in rl package took a “from” argument, but this had to
  necessary be the src argument in last VisionMap call, so FOV keeps now that
  information around and Ray has one less parameter now (incompatible change).
+ Field of view functions return now an iterable slice of light nodes, in the
  same spirit as functions do in the paths package.
+ Performance improvements in field of view computation (simple benchmarks show
  between 1.5 and 2 times improvement)
+ Remove useless return value in GridIterator.SetP (potentially incompatible
  change).

## v0.8.0 - 2020-01-10

+ New Iterator method for grids that allows for more flexible grid iterations. 

A few API changes in paths packages to make it more intuitive, consistent and
idiomatic.  Incompatible changes are mentioned as such below.

+ Make DijktraMap return the slice to be iterated and remove redundant MapIter
  (incompatible change).
+ Add DijktraMapAt method to get the cost at a specific point.
+ Rename CostAt into BreadthFirstMapAt for consistency (incompatible change).
+ Make BreadthFirstMap return a list of nodes to be iterated too, so that it's
  consistent with DijktraMap.
+ Rename ComputeCC and ComputeCCAll into CCMap and CCMapAll, and make CCMap
  return a slice of connected component points instead of requiring CCIter for
  iteration (incompatible change).
+ Fix maxCost checking in DijktraMap (nodes with cost = maxCost - 1 +
  dij.Cost(...) could be in the result map with Cost functions that return
  values > 1).

## v0.7.0 - 2020-01-08

+ New Vault parsing and manipulation utility in rl package.
+ Add a few new iteration facilities for grid slice types, and make some
  optimizations based on added benchmarks.
+ MapGen type in rl package has now value receiver methods (incompatible
  change).
+ Improve movement example with new features (such as non-explored cells).

## v0.6.0 - 2020-01-04

+ New field of view algorithm in rl package.
+ Add configurable map generation algorithms in rl package: RandomWalkCave and
  CellularAutomataCave.

## v0.5.0 - 2020-01-02

+ New Iter method for styled text.
+ Made pager's horizontal scrolling markup-aware.
+ Disable any markups in the text input widget (it doesn't make really sense there).
+ Fixed a potential crash with SetLines for pager.
+ Fixed issue with formatting of already multi-line styled text.
+ Added many tests.

## v0.4.0 - 2020-12-29

+ New EXPERIMENTAL rl package. It currently just offers an event priority queue.
+ Now model widgets in ui package only recompute drawing when the state changed
  since last Draw.  Previously, the optimization check was not robust, as it
  just skipped recomputing drawing if the last Update on the widget returned a
  \*Pass as action, which could skip necessary drawings in case the developper
  forgot to call Draw after a previous Update that actually changed the state.

## v0.3.0 - 2020-12-26

+ Renamed Menu.Style.Selected into Menu.Style.Active (incompatible change)
+ Added optional RuneManager in tcell driver.

## v0.2.0 - 2020-12-25

+ Improvements in ui widgets and examples.
+ Astar now returns the path in normal order instead of reversed (incompatible
  change), which made sense from an algorithm point of view but was
  unintuitive.

## v0.1.0 - 2020-12-24

First release of Gruid.
