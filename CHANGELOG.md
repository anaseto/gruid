This changelog file only lists important changes.

## v0.6.0 - 2020-12-04

+ New field of view algorithm in rl package.
+ Add configurable map generation algoritms in rl package: RandomWalkCave and
  CellularAutomataCave.

## v0.5.0 - 2020-12-02

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
