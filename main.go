package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const queryURL = "http://wowappprd.rio2016.com/json/medals/OG2016_medalsList.json"

type Country struct {
	NOC         string `json:"noc_code"`
	SortRank    string `json:"sort_rank"`
	Silver      string `json:"me_silver"`
	Rank        string `json:"rank"`
	SortRankTot string `json:"sort_rank_tot"`
	Gold        string `json:"me_gold"`
	RankTot     string `json:"rank_tot"`
	Bronze      string `json:"me_bronze"`
	MeTot       string `json:"me_tot"`
}

type Header struct {
	ResultMsg     string `json:"result_msg"`
	ServerTime    string `json:"server_time"`
	ResultCode    string `json:"result_code"`
	BusyDelayTime string `json:"busy_delay_time"`
	TranslationID string `json:"translation_id"`
}

type MedalList struct {
	MedalsList []Country `json:"medalsList"`
}

func (ms *MedalList) Len() int {
	return len(ms.MedalsList)
}

func (ms *MedalList) Swap(i, j int) {
	ms.MedalsList[i], ms.MedalsList[j] = ms.MedalsList[j], ms.MedalsList[i]
}

func (ms *MedalList) Less(i, j int) bool {
	a, _ := strconv.Atoi(ms.MedalsList[i].MeTot)
	b, _ := strconv.Atoi(ms.MedalsList[j].MeTot)
	return a > b
}

type BodyContent struct {
	MedalRank MedalList `json:"medalRank"`
}

type Record struct {
	Head Header      `json:"header"`
	Body BodyContent `json:"body"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var CurList Record

func GetJSON(h *Hub) {
	init := true
	duration := time.Duration(30) * time.Second
	for {
		if len(h.clients) == 0 && !init {
			time.Sleep(duration)
			continue
		}
		init = false
		resp, err := http.Get(queryURL)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		var rec Record
		err = json.NewDecoder(resp.Body).Decode(&rec)
		if err != nil {
			return
		}
		if reflect.DeepEqual(rec.Head, CurList.Head) {
			log.Println("the same...")
		} else {
			log.Println("fetching...")
			CurList = rec
			sort.Sort(&CurList.Body.MedalRank)
			h.broadcast <- CurList.Body.MedalRank.MedalsList
		}

		time.Sleep(duration)
	}
}

func main() {
	hub := newHub()
	go hub.run()
	go GetJSON(hub)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chatHandler(hub, w, r)
	})
	log.Fatal(http.ListenAndServe(":8181", nil))
}
