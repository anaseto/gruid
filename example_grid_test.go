// This example demontrates how to create a new grid and do some simple
// manipulations.
package gruid_test

import (
	"fmt"

	"github.com/anaseto/gruid"
)

func ExampleGrid() {
	// Create a new 20x20 grid.
	gd := gruid.NewGrid(20, 20)
	// Fill the whole grid with dots.
	gd.Fill(gruid.Cell{Rune: '.'})
	// Define a range (5,5)-(15,15).
	rg := gruid.NewRange(5, 5, 15, 15)
	// Define a slice of the grid using the range.
	rectangle := gd.Slice(rg)
	// Fill the rectangle with #.
	rectangle.Fill(gruid.Cell{Rune: '#'})
	// Print the grid using a non-styled string representation.
	fmt.Print(gd)
	// Output:
	// ....................
	// ....................
	// ....................
	// ....................
	// ....................
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// .....##########.....
	// ....................
	// ....................
	// ....................
	// ....................
	// ....................
}

func ExampleGridIterator() {
	// Create a new 26x2 grid.
	gd := gruid.NewGrid(26, 2)
	// Get an iterator.
	it := gd.Iterator()
	// Iterate on the grid and fill it with successive alphabetic
	// characters.
	r := 'a'
	max := gd.Size()
	for it.Next() {
		it.SetCell(gruid.Cell{Rune: r})
		r++
		if it.P().X == max.X-1 {
			r = 'A'
		}
	}
	// Print the grid using a non-styled string representation.
	fmt.Print(gd)
	// Output:
	// abcdefghijklmnopqrstuvwxyz
	// ABCDEFGHIJKLMNOPQRSTUVWXYZ
}
