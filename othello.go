package main

import (
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"log"
	"runtime"
	"container/list"
	"fmt"
	"time"
	"math/rand"
)

type User struct {
	UserName string `json:"user_name"`
	TournamentId int `json:"tournament_id"`
	UserRole string `json:"user_role"`
}

type PlayMovement struct {
	TournamentId int `json:"tournament_id"`
	PlayerTurnId int `json:"player_turn_id"`
	GameId int64 `json:"game_id"`
	Movement int `json:"movement"`
}

type ReadyResponse struct {
	PlayerTurnId int `json:"player_turn_id"`
	GameId int64 `json:"game_id"`
	Board [64]int `json:"board"`
	MovementNumber int `json:"movementNumber"`
}

type FinishResponse struct {
	GameId int64 `json:"game_id"`
	PlayerTurnId int `json:"player_turn_id"`
	WinnerTurnId int `json:"winner_turn_id"`
	Board []int `json:"board"`
}

type PlayerReady struct {
	TournamentId int `json:"tournament_id"`
	GameId int64 `json:"game_id"`
	PlayerTurnId int `json:"player_turn_id"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	c, err := gosocketio.Dial(
		gosocketio.GetUrl("192.168.1.112", 3000, false),
		// gosocketio.GetUrl("localhost", 3000, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		log.Fatal(err)
	}
	for{

		err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
			log.Println("Connected")
			newUser:=User{UserName:"Hsing", TournamentId: 142857, UserRole:"player"}
			err = c.Emit("signin", newUser)
			if err != nil {
				log.Fatal(err)
			}
		})
		if err != nil {
			log.Fatal(err)
		}


		err = c.On("ok_signin", func(h *gosocketio.Channel) {
			log.Println("Successfully signed in")
		})
		if err != nil {
			log.Fatal(err)
		}

		err = c.On("finish", func(h *gosocketio.Channel, data FinishResponse) {
			log.Println("Terminó la partida")
			playerReady := PlayerReady{TournamentId: 142857, GameId: data.GameId, PlayerTurnId: data.PlayerTurnId}
			err = c.Emit("player_ready", playerReady)
		})
		if err != nil {
			log.Fatal(err)
		}

		err = c.On("ready", func (h *gosocketio.Channel, data ReadyResponse) {
			log.Println("mi turno")
			log.Println(data)
			point := CPUPlay(data.PlayerTurnId, data.Board)
			fmt.Println(point)
			jugada := PlayMovement{TournamentId: 142857, PlayerTurnId: data.PlayerTurnId, GameId: data.GameId, Movement: point}
			err = c.Emit("play", jugada)
		})
		if err != nil {
			log.Fatal(err)
		}

	}

	c.Close()

	log.Println(" [x] Complete")
}



type State struct {
	board [BOARD_LEN][BOARD_LEN]int
	score map[int]int
	moves *list.List
}

type Point struct {
	X int
	Y int
}

const (
	EMPTY = 0
	BLACK = 1
	WHITE = 2
	WALL = 3
	BOARD_LEN = 10 //include wall
	CAP_POSSIBILITY = (BOARD_LEN - 2) * (BOARD_LEN - 2) / 2
	MOVES_QUEUE_SIZE = 4
)



func CPUPlay(player int, theBoard [64]int) (point int) {
	actualBoard, player1Score, player2Score := boardFormat(theBoard, player)
	var st State
	var p Point
	//Declaring vars
	iterations := 5
	//Define who is who
	opponent := 2

	//setting the actual board on the state
	st.board = actualBoard
	st.score = make( map[int]int )
	st.score[player] = player1Score
	st.score[opponent] = player2Score

	if(player == 2) {
		opponent = 1
		st.score[player] = player2Score
		st.score[opponent] = player1Score
	}
	//calc plays
	if(iterations > 0) {
		p, _ = getTheBestMove(st, player, iterations, 1)

	}
	// fmt.Println(p)
	point = convert(p, player)
	return
}


func getTheBestMove(st State, player int, look int, look_idx int) (p Point, best_eval int) {
	var eval int
	var new_st State
	opponent := 2
	if(player == 2) {
		opponent = 1
	}
	me_or_opponent := look_idx % 2

	movables := make( []Point, CAP_POSSIBILITY )

	if exploreMovables(&st, player, movables) {
		for i := 0; movables[i].X != 0 || movables[i].Y != 0; i++ {

			new_st.board = st.board
			new_st.score = make( map[int]int )
			new_st.score[player] = st.score[player]
			new_st.score[opponent] = st.score[opponent]

			putStone_And_CalcScore(&new_st, player, movables[i])

			eval = miniMax(new_st, opponent, look - 1, look_idx + 1, movables[i])

			if i == 0 {
				best_eval = eval
				p = movables[i]
			}

			switch me_or_opponent {
			case 0:
				if eval > best_eval {
					best_eval = eval
					p = movables[i]
				}
			case 1:
				if eval < best_eval {
					best_eval = eval
					p = movables[i]
				}
			default:
				fmt.Print("This is not Me or Opponent. There is something wrong!\n")
			}
		}
	} else {
		// fmt.Print("----------!!Pass!!----------\n")
		p.X, p.Y  = 0, 0
	}
	return
}
func exploreMovables(st *State, player int, points []Point) (v_bool bool) {
	var p Point
	idx := 0
	v_bool = false

	for i := 1; i <= BOARD_LEN - 2; i++ {
		for j := 1; j <= BOARD_LEN - 2; j++ {
			p.X, p.Y = i, j
			if checkStone(st, player, p) {
				points[idx] = p
				v_bool = true
				idx++
			}
		}
	}
	return
}
func putStone_And_CalcScore(st *State, player int, p Point) {
	var number int
	opponent := 2
	if(player == 2) {
		opponent = 1
	}
	points := make( []Point, 8 )

	if checkStone_And_Direction(st, player, p, points) {
		if player == BLACK {
			st.board[p.X][p.Y] = BLACK
			number = turnStones(st, player, p, points)
			st.score[player] += number + 1 /*single piece of stone which player first put on the board*/
			st.score[opponent] -= number
		} else {
			st.board[p.X][p.Y] = WHITE
			number = turnStones(st, player, p, points)
			st.score[player] += number + 1
			st.score[opponent] -= number
		}
	}
}


/* return numbers of stones turned */
func turnStones(st *State, player int, p Point, points []Point) (number int) {
	var variation Point
	var i, j int
	number = 0
	color := player
	back_color := BLACK

	// fmt.Printf("Turn Stones %v\n", points)

	for idx := 0; idx < 8; idx++ {

		variation = points[idx]
		if variation.X == 0 && variation.Y == 0 {
			break
		}

		i, j = p.X + variation.X, p.Y + variation.Y

		for st.board[i][j] == back_color {
			st.board[i][j] = color
			number += 1
			i += variation.X
			j += variation.Y
		}
	}
	return
}


func checkStone(st *State, player int, p Point) (v_bool bool) {
	/*first : BLACK : 1, second : WHITE : -1*/
	v_bool = false
	color := WHITE
	back_color := BLACK

	if st.board[p.X][p.Y] != EMPTY {
		return
	}

	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			i, j := p.X + dx, p.Y + dy

			for st.board[i][j] == back_color {
				i += dx
				j += dy

				if st.board[i][j] == color {
					v_bool = true
					return
				}
			}
		}
	}
	return
}
func checkStone_And_Direction(st *State, player int, p Point, points []Point) (v_bool bool) {

	/*first : BLACK : 1, second : WHITE : -1*/
	var variation Point
	v_bool = false
	slice_idx := 0
	color := player
	back_color := BLACK

	if st.board[p.X][p.Y] != EMPTY {
//		fmt.Print("This is not EMPTY!\n")
//		v_bool = false
		return
	}

	// fmt.Printf("points : %v\n", points)

	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			variation.X = dx
			variation.Y = dy
			i, j := p.X + dx, p.Y + dy

			for st.board[i][j] == back_color {
				i += dx
				j += dy

				if st.board[i][j] == color {
					v_bool = true
					if(slice_idx < 8) {
						points[slice_idx] = variation
						slice_idx += 1
					}
				}
			}
//			fmt.Print("not\n")
		}
	}

//
	return
}

func miniMax(st State, player int, look int, look_idx int, b_move Point) (eval int) {
	// eval = perfectEval(st, player)
	if look == 0 {
		// eval = perfectEval(st, player)
		eval = simpleEval(st, player)
	} else {
		/*recursive*/
		_, eval = getTheBestMove(st, player, look, look_idx)
	}
	// fmt.Printf("miniMax return : %d\n", eval)
	return
}
func simpleEval(st State, player int) (eval int) {
	var sum int
	var p Point
	points := make( []Point, 4 )
	complete_edges := [4]bool{false, false, false, false} /*top, right, bottom, left*/

	/*weight parameters*/
	weight_corner := 99
	weight_edges := 3

	if chanceToGetCorner(&st, player) {
		eval = weight_corner
		return
	}

	if chanceToGetEdges(&st, player, points) {
		for idx := 0; idx < 4; idx++ {
			p = points[idx]
			switch {
			case p.X == 1 && p.Y == 1:
				if complete_edges[0] == false {
					for n := 0; n < 8; n++ {
						if st.board[p.X][p.Y + n] == player {
							sum++
							if n == 7 {
								complete_edges[0] = true
							}
						} else {
							break
						}
					}
				}
				if complete_edges[3] == false {
					for n := 1; n < 8; n++ {
						if st.board[p.X + n][p.Y] == player {
							sum++
							if n == 7 {
								complete_edges[3] = true
							}

						} else {
							break
						}
					}
				}

			case p.X == 1 && p.Y == 8:
				if complete_edges[0] == false {
					for n := 0; n < 8; n++ {
						if st.board[p.X][p.Y - n] == player {
							sum++
						} else {
							break
						}
					}
				}
				if complete_edges[1] == false {
					for n := 1; n < 8; n++ {
						if st.board[p.X + n][p.Y] == player {
							sum++
							if n == 7 {
								complete_edges[1] = true
							}
						} else {
							break
						}
					}
				}
			case p.X == 8 && p.Y == 1:
				if complete_edges[2] == false {
					for n := 0; n < 8; n++ {
						if st.board[p.X][p.Y + n] == player {
							sum++
							if n == 7 {
								complete_edges[2] = true
							}
						} else {
							break
						}
					}
				}
				if complete_edges[3]  == false {
					for n := 1; n < 8; n++ {
						if st.board[p.X - n][p.Y] == player {
							sum++
						} else {
							break
						}
					}
				}

			case p.X == 8 && p.Y == 8:
				if complete_edges[2] == false {
					for n := 0; n < 8; n++ {
						if st.board[p.X][p.Y - n] == player {
							sum++
						} else {
							break
						}
					}
				}

				if complete_edges[1] == false {
					for n := 1; n < 8; n++ {
						if st.board[p.X - n][p.Y] == player {
							sum++
						} else {
							break
						}
					}
				}
			}
		}
		eval = sum * weight_edges
		return
	}
//

	// eval = int((float64(mobilityEval(st, player))* 0.75)  + (float64(pointsEval(st, player))* 0.25))
	// eval = int((float64(mobilityEval(st, player))* 0.4)  + (float64(perfectEval(st, player))* 0.3) + float64(randomEval()) * 0.3)
	eval = randomEval()
	// eval = pointsEval(st, player)
	return
}

func randomEval() (e int) {
	t := time.Now()

	rand.Seed(t.UnixNano())

	if rand.Intn(1) == 0 {
		e = rand.Intn(15)
	} else {
		e = rand.Intn(15) * (-1)
	}
	return
}

// func pointsEval(st State, player int) (eval int) {
// 	opponent := player * (-1)
// 	eval = st.score[player] - st.score[opponent]
// 	return
// }
//
// func mobilityEval(st State, player int) (eval int) {
// 	eval = getMoves(&st, player)
// 	return
// }

func getMoves(st *State, player int) (moves int) {
	var p Point
	for i := 1; i <= BOARD_LEN - 2; i++ {
		for j := 1; j <= BOARD_LEN - 2; j++ {
			p.X, p.Y = i, j
			if checkStone(st, player, p) {
				moves++
			}
		}
	}
	return
}

func chanceToGetCorner(st *State, player int) bool {
	var p Point

	p.X, p.Y = 1, 1
	if checkStone(st, player, p) {
		return true
	}

	p.X, p.Y = 1, 8
	if checkStone(st, player, p) {
		return true
	}

	p.X, p.Y = 8, 1
	if checkStone(st, player, p) {
		return true
	}

	p.X, p.Y = 8, 8
	if checkStone(st, player, p) {
		return true
	}

	return false
}

func chanceToGetEdges(st *State, player int, points []Point) (v_bool bool) {
	var p Point
	v_bool = false
	idx := 0

	for i := 1; i <= 8; i += 7 {
		for j := 1; j <= 8; j += 7 {
			p.X, p.Y = i, j
			if st.board[p.X][p.Y] == player {
				v_bool = true
				points[idx] = p
				idx++
			}
		}
	}
	return
}

func initBoard() (board [BOARD_LEN][BOARD_LEN]int) {
	board = [BOARD_LEN][BOARD_LEN]int{}
	for i := 0; i < BOARD_LEN; i++ {
		board[0][i] = 3
	}

	for i := 1; i < BOARD_LEN - 1; i++ {
		board[i][0] = 3
		board[i][9] = 3
	}

	for i := 0; i < BOARD_LEN; i++ {
		board[BOARD_LEN - 1][i] = 3
	}

	return
}

func convert(p Point, turn int) (point int) {
	point = 0
	// if(turn == 2) {
		// fmt.Println(p.X)
		// fmt.Println(p.Y)
		point = ((p.Y - 1) * 8)
		// fmt.Println(point)
		point = point + p.X - 1
		// fmt.Println(point)
	// }else {
	// 	point = ((p.X - 1) * 8)
	// 	fmt.Println(point)
	// 	point = point + p.Y - 2
	// }

	return
}

func boardFormat(board [64]int, turn int) (newBoard [BOARD_LEN][BOARD_LEN]int, player1Score int, player2Score int){
	newBoard = initBoard()
	player1Score = 2
	player2Score = 2
	for i := 0; i < 64; i++ {
		y := (i / 8) + 1
		x := (i % 8) + 1
		if(board[i] == 1) {
			player1Score++
		}else if(board[i] == 2) {
			player2Score++
		}
		if(turn == 2) {
			newBoard[x][y] = board[i]
		}else {
			if(board[i] == 1) {
				newBoard[x][y] = 2
			}else if(board[i] == 2) {
				newBoard[x][y] = 1
			}
		}
	}
	return
}
