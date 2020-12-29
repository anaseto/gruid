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
