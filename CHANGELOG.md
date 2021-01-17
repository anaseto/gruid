This changelog file only lists important changes.

## v0.10.0 - 2020-01-17

+ A few changes in Menu, Pager and TextInput configuration in order to make
  styling more flexible, and simplify configuration too (some incompatible
  changes).
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
+ Add configurable map generation algoritms in rl package: RandomWalkCave and
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
