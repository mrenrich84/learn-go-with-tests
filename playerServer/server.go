package poker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/websocket"
)

type PlayerStore interface {
	GetPlayerScore(name string) int
	RecordWin(name string)
	GetLeague() League
}

type PlayerServer struct {
	store PlayerStore
	http.Handler
	template *template.Template
	game     Game
}

const htmlTemplatePath = "game.html"

type Player struct {
	Name string
	Wins int
}


func NewPlayerServer(store PlayerStore, game Game) (*PlayerServer, error) {
	p := new(PlayerServer)

	tmpl, err := template.ParseFiles(htmlTemplatePath)

	if err != nil {
		return nil, fmt.Errorf("problem opening %s %v", htmlTemplatePath, err)
	}

	p.game = game

	p.template = tmpl
	p.store = store

	router := http.NewServeMux()
	router.Handle("/league", http.HandlerFunc(p.leagueHandler))
	router.Handle("/players/", http.HandlerFunc(p.playersHandler))
	router.Handle("/game", http.HandlerFunc(p.playGame))
	router.Handle("/ws", http.HandlerFunc(p.webSocket))

	p.Handler = router

	return p, nil
}
const jsonContentType = "application/json"

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(p.store.GetLeague())
}

func (p *PlayerServer) playersHandler(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	switch r.Method {
	case http.MethodPost:
		p.processWin(w, player)
	case http.MethodGet:
		p.showScore(w, player)
	}
}

func (p *PlayerServer) processWin(w http.ResponseWriter, player string) {
	p.store.RecordWin(player)
	w.WriteHeader(http.StatusAccepted)
}

func (p *PlayerServer) showScore(w http.ResponseWriter, player string) {
	score := p.store.GetPlayerScore(player)
	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}

func (p *PlayerServer) playGame(w http.ResponseWriter, r *http.Request) {
	p.template.Execute(w, nil)
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (p *PlayerServer) webSocket(w http.ResponseWriter, r *http.Request) {
	ws := newPlayerServerWS(w, r)

	numberOfPlayersMsg := ws.WaitForMsg()
	numberOfPlayers, _ := strconv.Atoi(numberOfPlayersMsg)
	p.game.Start(numberOfPlayers, ws)

	winner := ws.WaitForMsg()
	p.game.Finish(winner)
}

type playerServerWS struct {
	*websocket.Conn
}

func newPlayerServerWS(w http.ResponseWriter, r *http.Request) *playerServerWS {
	conn, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("problem upgrading connection to WebSockets %v\n", err)
	}

	return &playerServerWS{conn}
}

func (w *playerServerWS) WaitForMsg() string {
	_, msg, err := w.ReadMessage()
	if err != nil {
		log.Printf("error reading from websocket %v\n", err)
	}
	return string(msg)
}

func (w *playerServerWS) Write(p []byte) (n int, err error) {
	err = w.WriteMessage(websocket.TextMessage, p)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}