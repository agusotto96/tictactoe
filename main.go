package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	connCh := make(chan *websocket.Conn)
	gameCh := make(chan Game)
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		connCh <- conn
	})
	go http.ListenAndServe("0.0.0.0:8080", nil)
	go BuildGames(connCh, gameCh)
	for game := range gameCh {
		go StartGame(game)
	}
}

func BuildGames(connCh <-chan *websocket.Conn, gameCh chan<- Game) {
	game := Game{}
	for conn := range connCh {
		if game.P1 == nil {
			game.P1 = conn
		} else if game.P2 == nil {
			game.P2 = conn
			gameCh <- game
			game = Game{}
		}
	}
}

func StartGame(game Game) {
	p1 := game.P1
	p2 := game.P2
	ticTacToe := TicTacToe{}
	p1Turn := true
	winner := 0
	t1Err := WriteMsg(p1, &ticTacToe)
	t2Err := WriteMsg(p2, &ticTacToe)
	if t1Err == nil && t2Err == nil {
		for {
			if p1Turn {
				nErr := WriteMsg(p1, "TURN")
				if nErr != nil {
					break
				}
				move, rErr := ReadMsg(p1)
				if rErr != nil {
					break
				}
				pErr := ProcessMove(&ticTacToe, move, 1)
				if pErr != nil {
					break
				}
			} else {
				nErr := WriteMsg(p2, "TURN")
				if nErr != nil {
					break
				}
				move, rErr := ReadMsg(p2)
				if rErr != nil {
					break
				}
				pErr := ProcessMove(&ticTacToe, move, 2)
				if pErr != nil {
					break
				}
			}
			t1Err2 := WriteMsg(p1, &ticTacToe)
			t2Err2 := WriteMsg(p2, &ticTacToe)
			if t1Err2 != nil || t2Err2 != nil {
				break
			}
			winner = CheckWinner(&ticTacToe)
			if winner != 0 {
				if winner == 1 {
					WriteMsg(p1, "WIN")
					WriteMsg(p2, "LOSS")
				}
				if winner == 2 {
					WriteMsg(p1, "LOSS")
					WriteMsg(p2, "WIN")
				}
				if winner == 3 {
					WriteMsg(p1, "TIE")
					WriteMsg(p2, "TIE")
				}
				p1.Close()
				p2.Close()
				return
			}
			p1Turn = !p1Turn
		}
	}
	WriteMsg(p1, "DISCONNECTED")
	WriteMsg(p2, "DISCONNECTED")
	p1.Close()
	p2.Close()
}

func ReadMsg(conn *websocket.Conn) (string, error) {
	var msg string
	conn.SetReadDeadline(time.Now().Add(45 * time.Second))
	err := conn.ReadJSON(&msg)
	return msg, err
}

func WriteMsg(conn *websocket.Conn, msg interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return conn.WriteJSON(msg)
}

func ProcessMove(t *TicTacToe, move string, p int) error {
	if move == "NW" && t.NW == 0 {
		t.NW = p
		return nil
	}
	if move == "N" && t.N == 0 {
		t.N = p
		return nil
	}
	if move == "NE" && t.NE == 0 {
		t.NE = p
		return nil
	}
	if move == "W" && t.W == 0 {
		t.W = p
		return nil
	}
	if move == "C" && t.C == 0 {
		t.C = p
		return nil
	}
	if move == "E" && t.E == 0 {
		t.E = p
		return nil
	}
	if move == "SW" && t.SW == 0 {
		t.SW = p
		return nil
	}
	if move == "S" && t.S == 0 {
		t.S = p
		return nil
	}
	if move == "SE" && t.SE == 0 {
		t.SE = p
		return nil
	}
	return errors.New("invalid move")
}

func CheckWinner(t *TicTacToe) int {
	if (t.NW == 1 && t.N == 1 && t.NE == 1) ||
		(t.W == 1 && t.C == 1 && t.E == 1) ||
		(t.SW == 1 && t.S == 1 && t.SE == 1) ||
		(t.NW == 1 && t.W == 1 && t.SW == 1) ||
		(t.N == 1 && t.C == 1 && t.S == 1) ||
		(t.NE == 1 && t.E == 1 && t.SE == 1) ||
		(t.NW == 1 && t.C == 1 && t.SE == 1) ||
		(t.NE == 1 && t.C == 1 && t.SW == 1) {
		return 1
	}
	if (t.NW == 2 && t.N == 2 && t.NE == 2) ||
		(t.W == 2 && t.C == 2 && t.E == 2) ||
		(t.SW == 2 && t.S == 2 && t.SE == 2) ||
		(t.NW == 2 && t.W == 2 && t.SW == 2) ||
		(t.N == 2 && t.C == 2 && t.S == 2) ||
		(t.NE == 2 && t.E == 2 && t.SE == 2) ||
		(t.NW == 2 && t.C == 2 && t.SE == 2) ||
		(t.NE == 2 && t.C == 2 && t.SW == 2) {
		return 2
	}
	if t.NW == 0 || t.N == 0 || t.NE == 0 ||
		t.W == 0 || t.C == 0 || t.E == 0 ||
		t.SW == 0 || t.S == 0 || t.SE == 0 {
		return 0
	}
	return 3
}

type Game struct {
	P1, P2 *websocket.Conn
}

type TicTacToe struct {
	NW, N, NE, W, C, E, SW, S, SE int
}
