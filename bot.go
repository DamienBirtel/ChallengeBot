package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
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

// richnessMap[cell] stores the richness value of cell (0-3)
var richnessMap [37]int

var richnessValues = [4]int{0, 0, 2, 4}

// neighboursMap[cell][direction] stores the index of the adjacent cell
// from cell towards direction (-1 if no adjacent cell)
var neighboursMap [37][6]int

type worldState struct {
	day       int
	nutrients int

	sun   int
	score int

	oppSun   int
	oppScore int

	// treeMap[cellIndex] stores the value of size of the tree on this cell (1-4, 0 if none)
	treeMap [37]int

	// shadowMap[cell][shadowValue] 0 if no shadow,
	// 1 if there is a shadow cast by a tree of size shadowValue+1
	shadowMap [37][3]int

	// nbTrees[treeSize] stores the number of trees of size treeSize
	// already on the map and belonging to us
	nbTrees [4]int

	// nbOppTrees does the same as nbTrees for the opponent
	nbOppTrees [4]int

	// activeTrees and dormantTrees store the indexes of activeor dormant trees
	activeTrees  []int
	dormantTrees []int

	growCosts [3]int
}

func (ws *worldState) getInRangeFreeCells(startingCell int, radius int) []int {
	freeCells := [37]int{}
	toDoCells := []int{startingCell}
	for radius > 0 {
		nextToDoCells := []int{}
		for _, cell := range toDoCells {
			for _, neigh := range neighboursMap[cell] {
				if neigh != -1 && ws.treeMap[neigh] == 0 && richnessMap[neigh] > 0 {
					freeCells[neigh]++
					nextToDoCells = append(nextToDoCells, neigh)
				}
			}
		}
		toDoCells = nextToDoCells
		radius--
	}
	freeCells[startingCell] = 0
	doneCells := []int{}
	for i, cell := range freeCells {
		if cell > 0 {
			doneCells = append(doneCells, i)
		}
	}
	return doneCells
}

func (ws *worldState) getAllPossibleMoves() []move {
	moves := []move{}

	for _, treeIndex := range ws.activeTrees {
		treeSize := ws.treeMap[treeIndex]
		switch treeSize {
		case 1:
			if ws.sun >= ws.growCosts[0] {
				moves = append(moves, grow{index: treeIndex})
			}
		case 4:
			if ws.sun >= 4 {
				moves = append(moves, complete{index: treeIndex})
			}
			if ws.sun >= ws.nbTrees[0] {
				indexes := ws.getInRangeFreeCells(treeIndex, treeSize-1)
				for _, index := range indexes {
					moves = append(moves, seed{throwerIndex: treeIndex, receiverIndex: index})
				}
			}
		default:
			if ws.sun >= ws.growCosts[treeSize-1] {
				moves = append(moves, grow{index: treeIndex})
			}
			if ws.sun >= ws.nbTrees[0] {
				indexes := ws.getInRangeFreeCells(treeIndex, treeSize-1)
				for _, index := range indexes {
					moves = append(moves, seed{throwerIndex: treeIndex, receiverIndex: index})
				}
			}
		}
	}
	return moves
}

func (ws worldState) evaluate(m move) float64 {
	score := ws.score
	sun := ws.sun
	ws = m.simulate(ws)
	score = ws.score - score
	sun = ws.sun - sun
	score = score + sun
	return float64(score)
}

func (ws worldState) getScore() int {
	return ws.sun/3 + ws.score
}

type move interface {
	execute()
	simulate(ws worldState) worldState
}

type grow struct {
	index int
}

func (g grow) execute() {
	fmt.Println("GROW ", g.index)
}

func (g grow) simulate(ws worldState) worldState {
	size := ws.treeMap[g.index] - 1
	ws.treeMap[g.index]++
	ws.sun -= ws.growCosts[size]
	ws.growCosts[size]--
	ws.growCosts[size+1]++
	return ws
}

type seed struct {
	throwerIndex  int
	receiverIndex int
}

func (s seed) execute() {
	fmt.Println("SEED ", s.throwerIndex, " ", s.receiverIndex)
}

func (s seed) simulate(ws worldState) worldState {
	ws.sun -= ws.nbTrees[0]
	ws.nbTrees[0]++
	return ws
}

type complete struct {
	index int
}

func (c complete) execute() {
	fmt.Println("COMPLETE ", c.index)
}

func (c complete) simulate(ws worldState) worldState {
	ws.sun -= 4
	ws.score += ws.nutrients + richnessValues[richnessMap[c.index]]
	ws.treeMap[c.index] = 0
	ws.growCosts[2]--
	return ws
}

func main() {

	// numberOfCells: 37
	var numberOfCells int

	// index: 0 is the center cell, the next cells spiral outwards
	// richness: 0 if the cell is unusable, 1-3 for usable cells
	// neigh0: the index of the neighbouring cell for each direction
	var index, richness, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int

	// day: the game lasts 24 days: 0-23
	// var day int

	// nutrients: the base score you gain from the next COMPLETE action
	//var nutrients int

	// sun: your sun points
	// score: your current score
	//var sun, score int

	// oppSun: opponent's sun points
	// oppScore: opponent's score
	// oppIsWaiting: whether your opponent is asleep until the next day
	//var oppSun, oppScore int
	var oppIsWaiting bool
	var _oppIsWaiting int

	// numberOfTrees: the current amount of trees
	var numberOfTrees int

	// cellIndex: location of this tree
	// size: size of this tree: 0-3
	// isMine: 1 if this is your tree
	// isDormant: 1 if this tree is dormant
	var cellIndex, size int
	//var isMine, isDormant int
	var _isMine, _isDormant int

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfCells)
	for i := 0; i < numberOfCells; i++ {
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &index, &richness, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)
		richnessMap[index] = richness
		neighboursMap[index][0] = neigh0
		neighboursMap[index][1] = neigh1
		neighboursMap[index][2] = neigh2
		neighboursMap[index][3] = neigh3
		neighboursMap[index][4] = neigh4
		neighboursMap[index][5] = neigh5
	}

	ws := &worldState{}
	ws.growCosts[0] = 1
	ws.growCosts[1] = 3
	ws.growCosts[2] = 7
	for {
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &ws.day)

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &ws.nutrients)

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &ws.sun, &ws.score)

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &ws.oppSun, &ws.oppScore, &_oppIsWaiting)
		oppIsWaiting = _oppIsWaiting != 0
		_ = oppIsWaiting

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &numberOfTrees)
		ws.treeMap = [37]int{}
		ws.activeTrees = []int{}
		ws.dormantTrees = []int{}
		ws.nbTrees = [4]int{}
		for i := 0; i < numberOfTrees; i++ {
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &cellIndex, &size, &_isMine, &_isDormant)
			ws.treeMap[cellIndex] = size + 1
			switch _isMine {
			case 1:
				ws.nbTrees[size]++
				if size > 0 {
					ws.growCosts[size-1]++
				}
				switch _isDormant {
				case 1:
					ws.dormantTrees = append(ws.dormantTrees, cellIndex)
				default:
					ws.activeTrees = append(ws.activeTrees, cellIndex)
				}
			default:

			}
		}
		var numberOfPossibleActions int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &numberOfPossibleActions)

		for i := 0; i < numberOfPossibleActions; i++ {
			scanner.Scan()
			possibleAction := scanner.Text()
			_ = possibleAction // to avoid unused error // try printing something from here to start with
		}
		t0 := time.Now()
		possibleActions := ws.getAllPossibleMoves()
		switch len(possibleActions) {
		case 0:
			fmt.Println("WAIT")
		default:
			movesScore := make([]float64, len(possibleActions))
			//c := make(chan struct{}, 1)
			for i, m := range possibleActions {
				//go func(ms *[]float64, index int, m move) {
				w := *ws
				score := w.evaluate(m)
				//<-c
				movesScore[i] = score
				//c <- struct{}{}
				//}(&movesScore, i, m)
			}
			bestMove := movesScore[0]
			bestMoveIndex := 0
			for i, m := range movesScore {
				if m > bestMove {
					bestMove = m
					bestMoveIndex = i
				}
			}
			possibleActions[bestMoveIndex].execute()
		}
		fmt.Println(time.Since(t0))
	}
}
