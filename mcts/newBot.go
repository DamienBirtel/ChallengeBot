package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"time"
)

const (
	SEED     = 0
	GROW     = 1
	COMPLETE = 2
	WAIT     = 3

	EAST = 0
	NE   = 1
	NW   = 2
	WEST = 3
	SW   = 4
	SE   = 5

	FINAL_DAY = 24

	KNOWN_DEPTH = 2

	PLAYER   = 1
	OPPONENT = 0
)

/************************************************/
/*												*/
/*				GAME DATA AND LOGIC				*/
/*												*/
/************************************************/

// richnessMap stores to fertility value (0,2 or 4)of each cell (-1 if unfertile)
var richnessMap [37]int

// neighboursMap stores the indexes of every adjacent cells for each cell (-1 if out of bounds)
var neighboursMap [37][6]int

type State struct {
	day       int
	nutrients int

	// 1 = player info, 0 = opponent info
	sun   [2]int
	score [2]int

	// an array storing the size (0-3) of every tree on every cell (-1 = no tree)
	treeMap [37]int

	// an array storing the shadows of every tree by size for every cell (1 if shadowed, 0 if not)
	shadowMap [3][37]int

	// first array index indicates wich player has the trees
	// second array index indicates the size of the trees
	nbTrees [2][4]int

	// the index list of active and dormant trees for each player
	activeTreesIndex  [2][]int
	dormantTreesIndex [2][]int

	// this stores a value for each player to show if they are waiting (1), or not (0)
	isWaiting [2]int

	// the list of cost to grow a tree for each player
	growCost [2][3]int
}

type Move struct {
	code        int
	treeIndex   int
	targetIndex int
}

func newState() State {

	treeMap := [37]int{}
	for i := 0; i < len(treeMap); i++ {
		treeMap[i] = -1
	}

	s := State{
		day:               0,
		nutrients:         0,
		sun:               [2]int{},
		score:             [2]int{},
		treeMap:           treeMap,
		shadowMap:         [3][37]int{},
		nbTrees:           [2][4]int{},
		activeTreesIndex:  [2][]int{},
		dormantTreesIndex: [2][]int{},
		isWaiting:         [2]int{},
		growCost:          [2][3]int{{1, 3, 7}, {1, 3, 7}},
	}
	return s
}

func getData(scanner *bufio.Scanner) State {

	// numberOfTrees: the current amount of trees
	var numberOfTrees int

	// cellIndex: location of this tree
	// size: size of this tree: 0-3
	// isMine: 1 if this is your tree
	// isDormant: 1 if this tree is dormant
	var cellIndex, size int

	//var isMine, isDormant int
	var isMine, isDormant int

	s := newState()

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &s.day)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &s.nutrients)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &s.sun[PLAYER], &s.score[PLAYER])

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &s.sun[OPPONENT], &s.score[OPPONENT], &s.isWaiting[OPPONENT])

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfTrees)
	for i := 0; i < numberOfTrees; i++ {
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &cellIndex, &size, &isMine, &isDormant)

		s.treeMap[cellIndex] = size
		s.nbTrees[isMine][size]++
		switch isDormant {
		case 0:
			s.activeTreesIndex[isMine] = append(s.activeTreesIndex[isMine], cellIndex)
		default:
			s.dormantTreesIndex[isMine] = append(s.dormantTreesIndex[isMine], cellIndex)
		}
	}

	for i := 0; i < 3; i++ {
		s.growCost[PLAYER][i] += s.nbTrees[PLAYER][i+1]
		s.growCost[OPPONENT][i] += s.nbTrees[OPPONENT][i+1]
	}

	s = s.UpdateShadows()

	/////// 		don't know how to get rid of this
	///////
	var numberOfPossibleActions int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfPossibleActions)

	for i := 0; i < numberOfPossibleActions; i++ {
		scanner.Scan()
		possibleAction := scanner.Text()
		_ = possibleAction // to avoid unused error // try printing something from here to start with
	}
	///////
	/////// 		don't know how to get rid of this

	return s
}

func (m Move) Print() {
	switch m.code {
	case SEED:
		fmt.Println("SEED ", m.treeIndex, " ", m.targetIndex)
	case GROW:
		fmt.Println("GROW ", m.treeIndex)
	case COMPLETE:
		fmt.Println("GROW ", m.treeIndex)
	case WAIT:
		fmt.Println("WAIT")
	}
}

func UpdateOneShadow(shadowMap *[3][37]int, m Move, s State) {
	sunDirection := s.day % 6
	nextCell := neighboursMap[m.treeIndex][sunDirection]
	size := s.treeMap[m.treeIndex]
	if size > 0 {
		for i := 0; i < size; i++ {
			if nextCell < 0 {
				break
			}
			shadowMap[size-1][nextCell]--
			nextCell = neighboursMap[nextCell][sunDirection]
		}
	}
	if m.code == GROW {
		newSize := size + 1
		nextCell := neighboursMap[m.treeIndex][sunDirection]
		for i := 0; i < newSize; i++ {
			if nextCell < 0 {
				break
			}
			shadowMap[newSize-1][nextCell]++
			nextCell = neighboursMap[nextCell][sunDirection]
		}
	}
}

func (s State) UpdateShadows() State {
	sunDirection := s.day % 6
	s.shadowMap = [3][37]int{}
	for i, size := range s.treeMap {
		if size > 0 {
			nextCell := neighboursMap[i][sunDirection]
			for j := 0; j < size; j++ {
				if nextCell < 0 {
					break
				}
				s.shadowMap[size-1][nextCell]++
				nextCell = neighboursMap[nextCell][sunDirection]
			}
		}
	}
	return s
}

func (s State) GetSunPoints() [2]int {
	sunPoints := [2]int{}
	for i := 0; i < 2; i++ {
		for _, treeIndex := range s.activeTreesIndex[i] {
			size := s.treeMap[treeIndex]
			if size > 0 {
				for shadowIndex := size - 1; shadowIndex < 3; shadowIndex++ {
					if s.shadowMap[shadowIndex][treeIndex] > 0 {
						size = 0
						break
					}
				}
				sunPoints[i] += size
			}
		}
		sunPoints[i] += s.sun[i]
	}
	return sunPoints
}

func (s State) GetFreeCellsInRange(startingCell int, radius int) []int {
	freeCells := [37]int{}
	toDoCells := []int{startingCell}
	for radius > 0 {
		nextToDoCells := []int{}
		for _, cell := range toDoCells {
			for _, neigh := range neighboursMap[cell] {
				if neigh != -1 && s.treeMap[neigh] == -1 && richnessMap[neigh] > -1 {
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

func (s State) GetLegalMoves(playerCode int) []Move {

	moves := []Move{Move{code: WAIT}}
	if s.isWaiting[playerCode] == 1 {
		return moves
	}

	for _, treeIndex := range s.activeTreesIndex[playerCode] {
		size := s.treeMap[treeIndex]
		switch size {
		case 0:
			if s.sun[playerCode] >= s.growCost[playerCode][0] {
				moves = append(moves, Move{code: GROW, treeIndex: treeIndex})
			}
		case 3:
			if s.sun[playerCode] >= 4 {
				moves = append(moves, Move{code: COMPLETE, treeIndex: treeIndex})
			}
			if s.sun[playerCode] >= s.nbTrees[playerCode][0] {
				indexes := s.GetFreeCellsInRange(treeIndex, size)
				for _, index := range indexes {
					moves = append(moves, Move{code: SEED, treeIndex: treeIndex, targetIndex: index})
				}
			}
		default:
			if s.sun[playerCode] >= s.growCost[playerCode][size] {
				moves = append(moves, Move{code: GROW, treeIndex: treeIndex})
			}
			if s.sun[playerCode] >= s.nbTrees[playerCode][0] {
				indexes := s.GetFreeCellsInRange(treeIndex, size)
				for _, index := range indexes {
					moves = append(moves, Move{code: SEED, treeIndex: treeIndex, targetIndex: index})
				}
			}
		}
	}
	return moves
}

func (s State) IsEqual(st State) bool {
	isEqual := (s.sun == st.sun) && (s.score == st.score) && (s.treeMap == st.treeMap)
	if !isEqual {
		return false
	}
	for i := 0; i < len(s.activeTreesIndex[PLAYER]); i++ {
		if s.activeTreesIndex[PLAYER][i] != st.activeTreesIndex[PLAYER][i] {
			return false
		}
	}
	for i := 0; i < len(s.activeTreesIndex[OPPONENT]); i++ {
		if s.activeTreesIndex[OPPONENT][i] != st.activeTreesIndex[OPPONENT][i] {
			return false
		}
	}
	return true
}

func RemoveFromSlice(s []int, element int) []int {
	for i, e := range s {
		if e == element {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
			break
		}
	}
	return s
}

func (s State) Seed(m Move, playerCode int, isSuccessful bool) State {
	s.activeTreesIndex[playerCode] = RemoveFromSlice(s.activeTreesIndex[playerCode], m.treeIndex)
	s.dormantTreesIndex[playerCode] = append(s.dormantTreesIndex[playerCode], m.treeIndex)
	s.sun[playerCode] -= s.nbTrees[playerCode][0]
	if isSuccessful {
		s.dormantTreesIndex[playerCode] = append(s.dormantTreesIndex[playerCode], m.targetIndex)
		s.treeMap[m.targetIndex] = 0
		s.nbTrees[playerCode][0]++
	}
	sort.Ints(s.activeTreesIndex[playerCode])
	sort.Ints(s.dormantTreesIndex[playerCode])
	return s
}

func (s State) Grow(m Move, playerCode int) State {
	s.activeTreesIndex[playerCode] = RemoveFromSlice(s.activeTreesIndex[playerCode], m.treeIndex)
	s.dormantTreesIndex[playerCode] = append(s.dormantTreesIndex[playerCode], m.treeIndex)
	size := s.treeMap[m.treeIndex]
	s.sun[playerCode] -= s.growCost[playerCode][size]
	s.growCost[playerCode][size]++
	s.nbTrees[playerCode][size+1]++
	s.nbTrees[playerCode][size]--
	if size > 0 {
		s.growCost[playerCode][size-1]--
	}
	UpdateOneShadow(&s.shadowMap, m, s)
	s.treeMap[m.treeIndex]++
	sort.Ints(s.activeTreesIndex[playerCode])
	sort.Ints(s.dormantTreesIndex[playerCode])
	return s
}

func (s State) Complete(m Move, playerCode int) State {
	s.activeTreesIndex[playerCode] = RemoveFromSlice(s.activeTreesIndex[playerCode], m.treeIndex)
	UpdateOneShadow(&s.shadowMap, m, s)
	s.treeMap[m.treeIndex] = 0
	s.score[playerCode] += s.nutrients
	s.nutrients--
	s.nbTrees[playerCode][3]--
	sort.Ints(s.activeTreesIndex[playerCode])
	return s
}

func (s State) PlaySpecialCases(playerMove Move, opponentMove Move) State {

	switch playerMove.code {
	case WAIT:
		s.day++
		s.activeTreesIndex[PLAYER] = append(s.activeTreesIndex[PLAYER], s.dormantTreesIndex[PLAYER]...)
		s.dormantTreesIndex[PLAYER] = []int{}
		s.activeTreesIndex[OPPONENT] = append(s.activeTreesIndex[OPPONENT], s.dormantTreesIndex[OPPONENT]...)
		s.dormantTreesIndex[OPPONENT] = []int{}
		s.isWaiting[OPPONENT], s.isWaiting[PLAYER] = 0, 0
		s = s.UpdateShadows()
		s.sun = s.GetSunPoints()
		sort.Ints(s.activeTreesIndex[PLAYER])
		sort.Ints(s.activeTreesIndex[OPPONENT])
	case SEED:
		isSuccessful := true
		if playerMove.targetIndex == opponentMove.targetIndex {
			isSuccessful = false
		}
		s = s.Seed(playerMove, PLAYER, isSuccessful)
		s = s.Seed(opponentMove, OPPONENT, isSuccessful)
	case GROW:
		s = s.Grow(playerMove, PLAYER)
		s = s.Grow(opponentMove, OPPONENT)
	case COMPLETE:
		s = s.Complete(playerMove, PLAYER)
		s = s.Complete(opponentMove, OPPONENT)
		s.score[OPPONENT]++
	}
	return s
}

func (s State) Play(playerMove Move, opponentMove Move) State {

	if playerMove.code == opponentMove.code {
		return s.PlaySpecialCases(playerMove, opponentMove)
	}

	switch playerMove.code {
	case WAIT:
		s.isWaiting[PLAYER] = 1
	case SEED:
		s = s.Seed(playerMove, PLAYER, true)
	case GROW:
		s = s.Grow(playerMove, PLAYER)
	case COMPLETE:
		s = s.Complete(playerMove, PLAYER)
	}

	switch opponentMove.code {
	case WAIT:
		s.isWaiting[OPPONENT] = 1
	case SEED:
		s = s.Seed(opponentMove, OPPONENT, true)
	case GROW:
		s = s.Grow(opponentMove, OPPONENT)
	case COMPLETE:
		s = s.Complete(opponentMove, OPPONENT)
	}
	return s
}

/************************************************/
/*												*/
/*				TREE DATA AND LOGIC				*/
/*												*/
/************************************************/

type GameTree struct {
	root *Node
}

type Node struct {
	nbVisit           int
	state             State
	playerMoveList    []Move
	opponentMoveList  []Move
	playerMoveScore   []float64
	opponentMoveScore []float64
	parent            *Node
	children          []*Node
}

func newNode(s State, parentNode *Node) *Node {
	return &Node{
		nbVisit:           0,
		state:             s,
		playerMoveList:    s.GetLegalMoves(PLAYER),
		opponentMoveList:  s.GetLegalMoves(OPPONENT),
		playerMoveScore:   []float64{},
		opponentMoveScore: []float64{},
		parent:            parentNode,
		children:          []*Node{},
	}
}

func (n *Node) GetAllChildrenNodes(depth int) {

	if depth == 0 {
		return
	}

	if n.state.day == FINAL_DAY {
		return
	}

	if len(n.children) == 0 {
		for _, playerMove := range n.playerMoveList {
			for _, opponentMove := range n.opponentMoveList {
				nextState := n.state.Play(playerMove, opponentMove)
				n.children = append(n.children, newNode(nextState, n))
			}
		}
	}

	for _, child := range n.children {
		child.GetAllChildrenNodes(depth - 1)
	}
}

func (gt *GameTree) Update(s State) {

	if gt.root == nil {
		gt.root = newNode(s, nil)
	} else {

		for i, child := range gt.root.children {
			if child.state.IsEqual(s) {
				gt.root = gt.root.children[i]
				gt.root.parent = nil
				break
			}
		}
	}
	gt.root.GetAllChildrenNodes(KNOWN_DEPTH)
}

func (gt *GameTree) Compute(t time.Duration) Move {
	t0 := time.Now()
	for time.Since(t0) < t {

	}
	bestMove := Move{code: WAIT}
	return bestMove
}

func (gt *GameTree) Print() {
	fmt.Fprintln(os.Stderr, gt.root.children[0])
}

/************************************************/
/*												*/
/*					MAIN LOGIC					*/
/*												*/
/************************************************/

func main() {

	// numberOfCells: 37
	var numberOfCells int

	// index: 0 is the center cell, the next cells spiral outwards
	// richness: 0 if the cell is unusable, 1-3 for usable cells
	// neigh0: the index of the neighbouring cell for each direction
	var index, richness, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfCells)
	for i := 0; i < numberOfCells; i++ {
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &index, &richness, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)
		switch richness {
		case 0:
			richnessMap[index] = -1
		case 1:
			richnessMap[index] = 0
		case 2:
			richnessMap[index] = 2
		case 3:
			richnessMap[index] = 4
		}
		neighboursMap[index][0] = neigh0
		neighboursMap[index][1] = neigh1
		neighboursMap[index][2] = neigh2
		neighboursMap[index][3] = neigh3
		neighboursMap[index][4] = neigh4
		neighboursMap[index][5] = neigh5
	}

	gameTree := &GameTree{}
	firstRound := true
	firstToWait := false
	for {
		if firstToWait {
			firstToWait = false
			gameTree.root = nil
		}
		state := getData(scanner)

		t0 := time.Now()
		gameTree.Update(state)

		// compute stuff while there is time
		t := 1 * time.Millisecond
		if firstRound {
			firstRound = false
			t = 1 * time.Millisecond
		}
		move := gameTree.Compute(t)
		if move.code == WAIT && state.isWaiting[OPPONENT] == 0 {
			firstToWait = true
		}
		move.Print()
		fmt.Fprintln(os.Stderr, time.Since(t0))
	}
}
