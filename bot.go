package main

import (
	"bufio"
	"fmt"
	"os"
)

// Directions values
const (
	EAST = 0
	NE   = 1
	NW   = 2
	WEST = 3
	SW   = 4
	SE   = 5
)

type cell struct {
	index     int
	richness  int
	neighbors [6]*cell
}

func newCell(idx int) *cell {
	return &cell{index: idx}
}

type world struct {
	grid [37]*cell
}

func newWorld() *world {
	return &world{}
}

func (w *world) fillLayer(size int, lastLayerSize int, currentLayer int, anchorCell *cell) {
	index := lastLayerSize + 1
	w.grid[index] = newCell(index)

	currentCell := w.grid[index]
	anchorCell.neighbors[EAST] = currentCell
	currentCell.neighbors[WEST] = anchorCell

	direction := NW
	overAnchorCell := anchorCell.neighbors[direction]
	for i := index + 1; i <= size; i++ {
		lastCell := currentCell
		currentCell = newCell(i)
		lastCell.neighbors[direction] = currentCell
		currentCell.neighbors[(direction+3)%6] = lastCell
		currentCell.neighbors[(direction+2)%6] = anchorCell
		anchorCell.neighbors[(direction+5)%6] = currentCell
		switch overAnchorCell {
		case nil:
			direction = (direction + 1) % 6
		default:
			overAnchorCell.neighbors[(direction+4)%6] = currentCell
			currentCell.neighbors[(direction+1)%6] = overAnchorCell
			anchorCell = overAnchorCell
		}
		overAnchorCell = anchorCell.neighbors[direction]
		w.grid[i] = currentCell
	}
	if lastLayerSize > 0 {
		w.grid[size].neighbors[direction] = w.grid[index]
		w.grid[index].neighbors[(direction+3)%6] = w.grid[size]
	}
}

func (w *world) fillGrid() {

	w.grid[0] = newCell(0)

	lastLayerSize := 0
	anchorCell := w.grid[0]
	for currentLayer := 1; currentLayer < 4; currentLayer++ {
		currentLayerSize := lastLayerSize + currentLayer*6
		w.fillLayer(currentLayerSize, lastLayerSize, currentLayer, anchorCell)
		anchorCell = w.grid[lastLayerSize+1]
		lastLayerSize = currentLayerSize
	}
}

func main() {

	// numberOfCells: 37
	var numberOfCells int

	// index: 0 is the center cell, the next cells spiral outwards
	// richness: 0 if the cell is unusable, 1-3 for usable cells
	// neigh0: the index of the neighbouring cell for each direction
	var index, richness, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int

	// day: the game lasts 24 days: 0-23
	var day int

	// nutrients: the base score you gain from the next COMPLETE action
	var nutrients int

	// sun: your sun points
	// score: your current score
	var sun, score int

	// oppSun: opponent's sun points
	// oppScore: opponent's score
	// oppIsWaiting: whether your opponent is asleep until the next day
	var oppSun, oppScore int
	var oppIsWaiting bool
	var _oppIsWaiting int

	// numberOfTrees: the current amount of trees
	var numberOfTrees int

	// cellIndex: location of this tree
	// size: size of this tree: 0-3
	// isMine: 1 if this is your tree
	// isDormant: 1 if this tree is dormant
	var cellIndex, size int
	var isMine, isDormant bool
	var _isMine, _isDormant int

	var numberOfPossibleMoves int

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	// FILL HEXGRID WITH CELLS
	world := newWorld()
	world.fillGrid()

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfCells)
	for i := 0; i < numberOfCells; i++ {
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &index, &richness, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)
		world.grid[index].richness = richness
	}

	// ROUNDS LOOP
	for {
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &day)

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &nutrients)

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &sun, &score)

		_ = oppIsWaiting
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &oppSun, &oppScore, &_oppIsWaiting)
		oppIsWaiting = _oppIsWaiting != 0

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &numberOfTrees)

		for i := 0; i < numberOfTrees; i++ {

			_, _ = isDormant, isMine
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &cellIndex, &size, &_isMine, &_isDormant)
			isMine = _isMine != 0
			isDormant = _isDormant != 0
		}

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &numberOfPossibleMoves)

		for i := 0; i < numberOfPossibleMoves; i++ {
			scanner.Scan()
			possibleMove := scanner.Text()
			_ = possibleMove // to avoid unused error
		}

		// fmt.Fprintln(os.Stderr, "Debug messages...")

		// GROW cellIdx | SEED sourceIdx targetIdx | COMPLETE cellIdx | WAIT <message>
	}
}
