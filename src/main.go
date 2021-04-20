package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"backgamoon-back/src/structs"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// PORT is for port
const PORT = ":3000"

// ORDERS Variable
var ORDERS structs.Orders
var mysql *sql.DB
var timers = make(map[int]*time.Timer)
var connectionTickers = make(map[int]*time.Ticker)
var ping, _ = json.Marshal(1)
var disconnectNotification, _ = json.Marshal(5)
var reconnectNotification, _ = json.Marshal(6)

var clients = make(map[int]*websocket.Conn)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func recreateGame(tableID int) structs.GameData {
	mysql = openConnection()
	var col structs.TestStruct
	var col1 string
	var data structs.TestStruct1
	var order structs.TestStruct2

	rows, _ := mysql.Query("SELECT `data`.`data` FROM data WHERE data.tableID = ?", tableID)

	for rows.Next() {
		err := rows.Scan(&col1)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal([]byte(col1), &col)
		if col.Order == ORDERS.BeginGame {
			json.Unmarshal([]byte(col1), &data)
		} else if col.Order == ORDERS.MoveCoin {
			json.Unmarshal([]byte(col1), &order)
			isAccessibleCoin(&data.Payload, order.Payload[0], order.Payload[1])
		} else if col.Order == ORDERS.RollDice {
			json.Unmarshal([]byte(col1), &order)
			recreateRollDice(&data.Payload, order.Payload)
		} else if col.Order == ORDERS.AwakeCoin {
			var order structs.TestStruct3
			json.Unmarshal([]byte(col1), &order)
			makeCoinAlive(&data.Payload, order.Payload)
		} else if col.Order == ORDERS.CoinOut {
			var order structs.TestStruct3
			json.Unmarshal([]byte(col1), &order)
			coinGoesOut(&data.Payload, order.Payload)
		} else if col.Order == ORDERS.RollOneDice {
			json.Unmarshal([]byte(col1), &order)
			rollOneDice(&data.Payload, order.Payload)
		}

	}
	mysql.Close()
	return data.Payload
}

func rollOneDice(data *structs.GameData, from []int) {
	var rollerIndex = from[0]
	var winnerUserIndex = rollerIndex

	if data.FirstDice.Rolled == nil {
		data.FirstDice.Rolled = make(map[int]int)
	}

	data.FirstDice.Rolled[rollerIndex] = from[1]

	if data.FirstDice.Rolled[1-rollerIndex] != 0 {
		if data.FirstDice.Rolled[1-rollerIndex] == data.FirstDice.Rolled[rollerIndex] {
			data.FirstDice.Rolled[1-rollerIndex] = 0
			data.FirstDice.Rolled[rollerIndex] = 0
		} else if data.FirstDice.Rolled[1-rollerIndex] > data.FirstDice.Rolled[rollerIndex] {
			winnerUserIndex = 1 - rollerIndex
		} else {
			winnerUserIndex = rollerIndex
		}

		data.Begginer = data.Players[winnerUserIndex].ID
	}
}

func coinGoesOut(data *structs.GameData, from int) {
	currentPlayer := data.CurrentPlayer()

	index1 := 0
	if currentPlayer.Direction == true {
		index1 = from
	} else {
		index1 = 23 - from + 1
	}

	smallerDice := 0

	if data.Rolled[1].Rolled < data.Rolled[smallerDice].Rolled {
		smallerDice = 1
	}

	// TODO use mutex to lock user's data when request comes. because if multiple requests come at the same time code BREAKS.

	if data.Rolled[smallerDice].Used == 0 && data.Rolled[smallerDice].Rolled >= index1 {
		if data.Quadro == false {
			data.Rolled[smallerDice].Used = 1
		} else {
			if data.RollQuantity == 1 {
				data.Rolled[0].Used = 1
				data.Rolled[1].Used = 1
			}
		}
	} else if data.Rolled[1-smallerDice].Used == 0 && data.Rolled[1-smallerDice].Rolled >= index1 {
		if data.Quadro == false {
			data.Rolled[1-smallerDice].Used = 1
		} else {
			if data.RollQuantity == 1 {
				data.Rolled[0].Used = 1
				data.Rolled[1].Used = 1
			}
		}
	}
	data.RollQuantity--
	data.ThrowCoin(from)
}

func makeCoinAlive(data *structs.GameData, to int) {
	player := data.CurrentPlayer()
	playerIndex := data.CurrentPlayerIndex()
	i := 0
	if player.Direction {
		i = 23 - to + 1
	} else {
		i = to + 1
	}

	if data.Quadro == true {
		//aq memgoni ragac shecdomaa
		if data.Rolled[0].Rolled == i && data.Rolled[0].Used == 0 {
			data.AwakeCoin(playerIndex, to)
		} else if data.Rolled[1].Rolled == i && data.Rolled[1].Used == 0 {
			data.AwakeCoin(playerIndex, to)
		} else {
			return
		}
	} else {
		//TODO check if user can place coin there. if tunnel.coins == 0 || tunnel.O == userID
		if data.Rolled[0].Rolled == i && data.Rolled[0].Used == 0 {
			data.AwakeCoin(playerIndex, to)
			data.Rolled[0].Used = 1
		} else if data.Rolled[1].Rolled == i && data.Rolled[1].Used == 0 {
			data.AwakeCoin(playerIndex, to)
			data.Rolled[1].Used = 1
		} else {
			return
		}
	}

	data.RollQuantity--
}

func isAccessibleCoin(data *structs.GameData, from int, to int) {
	delta := int(math.Abs(float64(to) - float64(from)))
	currentPlayer := data.CurrentPlayer()
	if data.Quadro {
		if delta%data.Rolled[0].Rolled == 0 { // თუ ნამდვილად გაგორებულის ჯერადზე გადადის.
			if delta/data.Rolled[0].Rolled <= data.RollQuantity { // თუ ყოფნის სვლების რაოდენობა.
				for index := 0; index < (delta / data.Rolled[0].Rolled); index++ {
					if currentPlayer.Direction {
						if !data.IsMovable(from, from-data.Rolled[0].Rolled*(index+1)) {
							return
						}
					} else {
						if !data.IsMovable(from, from+data.Rolled[0].Rolled*(index+1)) {
							return
						}
					}
				}
				data.RollQuantity -= delta / data.Rolled[0].Rolled
				fmt.Println(data.RollQuantity)
				data.MoveCoin(from, to)
				return
			}
			fmt.Println(`You are not able to move on such big destination`)
			return

		}
		fmt.Println(`You are not able to move coin on this place because ${delta}%${data.Rolled[0].Rolled}=${delta % data.Rolled[0].Rolled}`)
		return
	}

	if delta != data.Rolled[0].Rolled+data.Rolled[1].Rolled &&
		delta != data.Rolled[0].Rolled &&
		delta != data.Rolled[1].Rolled {
		fmt.Println("recreate you are requesting unprocessable roll")
		return
	}
	correctDirection := from < to
	tempcvladi := data.Rolled[0].Rolled
	tempcvladi1 := data.Rolled[1].Rolled
	if currentPlayer.Direction {
		correctDirection = from > to
		tempcvladi *= -1
		tempcvladi1 *= -1
	} else {
		correctDirection = from < to
		tempcvladi = data.Rolled[0].Rolled
		tempcvladi1 = data.Rolled[1].Rolled
	}

	if correctDirection {
		if data.Rolled[0].Used == 0 &&
			data.Rolled[1].Used == 0 &&
			delta == data.Rolled[0].Rolled+data.Rolled[1].Rolled {
			if data.IsMovable(from, to) {
				if data.IsMovable(from, from+tempcvladi) ||
					data.IsMovable(from, from+tempcvladi1) {
					data.MoveCoin(from, to)
					data.Rolled[0].Used = 1
					data.Rolled[1].Used = 1
					data.RollQuantity = 0
					return
				}
				fmt.Println("YOU CAN NOT MOVE YOUR COIN HERE")
				return

			}
		} else if data.Rolled[0].Used == 0 &&
			data.Rolled[0].Rolled == delta &&
			data.IsMovable(from, to) {
			data.MoveCoin(from, to)
			data.Rolled[0].Used = 1
			data.RollQuantity--
			return
		} else if data.Rolled[1].Used == 0 &&
			data.Rolled[1].Rolled == delta &&
			data.IsMovable(from, to) {
			data.MoveCoin(from, to)
			data.Rolled[1].Used = 1
			data.RollQuantity--
			return
		}
	} else {
		fmt.Println("Incorrect Direction")
		return
	}

}

func recreateRollDice(data *structs.GameData, rolled []int) {
	data.Rolled[0].Rolled = rolled[0]
	data.Rolled[1].Rolled = rolled[1]
	data.Rolled[0].Used = 0
	data.Rolled[1].Used = 0
	data.RollQuantity = 2
	data.Quadro = false

	if data.Roller == 0 {
		data.Roller = data.Begginer
	} else {
		if data.Roller == data.Players[0].ID {
			data.Roller = data.Players[1].ID
		} else {
			data.Roller = data.Players[0].ID
		}
	}

	if data.Rolled[1].Rolled == data.Rolled[0].Rolled {
		data.RollQuantity = 4
		data.Quadro = true
	}
}

func rollDice(data *structs.Desk) {
	data.GameData.Rolled[0].Rolled = generateRandom(1, 6)
	data.GameData.Rolled[1].Rolled = generateRandom(1, 6)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ORDERS = structs.Orders{
		BeginGame:   0,
		MoveCoin:    1,
		RollDice:    2,
		CheckBegin:  3,
		Connection:  4,
		CoinMoved:   5,
		AwakeCoin:   7,
		CoinOut:     8,
		RollOneDice: 9,
	}

	router := mux.NewRouter()

	router.HandleFunc("/ws", wsHandler)

	fmt.Println("Your backgammon is listening on", PORT)

	// go listeToInput()
	log.Fatal(http.ListenAndServe(PORT, router))

}

func listeToInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("rs", text) == 0 {
			fmt.Println("restarting...")
			restart()
			fmt.Println("Good to GO : )")
		}
	}
}

func restart() {
	mysql = openConnection()
	mysql.Exec("TRUNCATE `data`")
	mysql.Exec("TRUNCATE usersStatus")
	mysql.Exec("DELETE FROM tablesStatus WHERE status != 0")
	for userID, element := range clients {
		connectionTickers[userID].Stop()
		element.Close()
	}
}

func sendDisconnectNotification(TableID int) {
	mysql = openConnection()
	rows, err := mysql.Query("SELECT userID from usersStatus WHERE tableID = ?", TableID)

	defer rows.Close()

	fancyHandleError(err)
	var userID int
	for rows.Next() {
		if err := rows.Scan(&userID); err != nil {
			log.Fatal(err.Error())
		}
		sendMessage(userID, disconnectNotification)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err.Error())
	}
}

func pingConnection(tableID int, userID int) {
	if connectionTickers[userID] != nil {
		connectionTickers[userID].Stop()
	}

	connectionTickers[userID] = time.NewTicker(1 * time.Second)

	for range connectionTickers[userID].C {
		err := clients[userID].WriteMessage(websocket.TextMessage, []byte(ping))
		if err != nil {
			mysql = openConnection()
			mysql.Query("INSERT INTO usersStatus (tableID, userID, status) VALUES (?, ?, 2)", tableID, userID) // set user status to disconnected
			mysql.Query("INSERT INTO tablesStatus VALUES (NULL, ?, 2)", tableID)                               // set table status to paused
			sendDisconnectNotification(tableID)
			connectionTickers[userID].Stop()
			clients[userID].Close()
			return
		}
	}
}

func printtable(tableID int, userID1 int, userID2 int, delay time.Duration, command int) {
	timers[tableID] = time.NewTimer(delay * time.Second)
	<-timers[tableID].C
	returnData := structs.MessageMOVE_COIN{Order: ORDERS.AwakeCoin} // TODO write good order code.
	returnData.Success = false
	byteCode, _ := json.Marshal(returnData)
	sendToUsers(append([]int{}, userID1, userID2), byteCode)
}

func cancelTimer(tableID int) {
	if timers[tableID] != nil {
		fmt.Println("Timer calnceled")
		timers[tableID].Stop()
	}
}

func generateRandom(from int, to int) int {
	return from + rand.Intn(to-from+1)
}

func addCoinsToTunnel(quantity int, playerID int) structs.Tunnel {
	return structs.Tunnel{O: playerID, Coins: quantity}
}

func generateData(tableID int) structs.Desk {
	mysql = openConnection()
	var data structs.Desk
	data.Order = ORDERS.BeginGame
	var userIDs []int
	var userID int

	results, err := mysql.Query("SELECT `userID` FROM usersStatus WHERE tableID = ?", tableID)
	for results.Next() {
		err = results.Scan(&userID)
		fancyHandleError(err)
		userIDs = append(userIDs, userID)
	}

	data.GameData.KilledCoins = append(data.GameData.KilledCoins, structs.Tunnel{O: userIDs[0]}, structs.Tunnel{O: userIDs[1]})
	data.GameData.Players = append(data.GameData.Players, structs.Player{ID: userIDs[0], Direction: true}, structs.Player{ID: userIDs[1], Direction: false})
	data.GameData.Rolled = append(data.GameData.Rolled, structs.Dice{}, structs.Dice{})
	data.GameData.Tunnels = append(data.GameData.Tunnels, addCoinsToTunnel(2, userIDs[1]), structs.Tunnel{}, structs.Tunnel{}, structs.Tunnel{}, structs.Tunnel{}, addCoinsToTunnel(5, userIDs[0]), structs.Tunnel{}, addCoinsToTunnel(3, userIDs[0]), structs.Tunnel{}, structs.Tunnel{}, structs.Tunnel{}, addCoinsToTunnel(5, userIDs[1]), addCoinsToTunnel(5, userIDs[0]), structs.Tunnel{}, structs.Tunnel{}, structs.Tunnel{}, addCoinsToTunnel(3, userIDs[1]), structs.Tunnel{}, addCoinsToTunnel(5, userIDs[1]), structs.Tunnel{}, structs.Tunnel{}, structs.Tunnel{}, structs.Tunnel{}, addCoinsToTunnel(2, userIDs[0]))
	data.GameData.OutCoins = append(data.GameData.OutCoins, structs.Tunnel{Coins: 0, O: userIDs[0]}, structs.Tunnel{Coins: 0, O: userIDs[1]})
	data.GameData.Roller = 0
	mysql.Close()
	return data
}

func checkTableStatus(tableID int) structs.CheckTableStatus {
	mysql = openConnection()
	var status int
	var err = mysql.QueryRow("SELECT state FROM tables WHERE `tables`.id = ?", tableID).Scan(&status)
	if err != nil {
		log.Fatal(err)
	}
	mysql.Close()

	return structs.CheckTableStatus{Status: status}
}

func checkBegin(tableID int) structs.CheckBeginStruct {
	mysql = openConnection()
	results, err := mysql.Query("SELECT DISTINCT(activeTables.userID), playerStatus.status FROM activeTables INNER JOIN playerStatus ON userID = playerStatus.playerID WHERE tableID = ? LIMIT 2", tableID)
	var begin = true
	var userID int
	var status int
	var userIDs []int
	fancyHandleError(err)
	for results.Next() {
		err = results.Scan(&userID, &status)
		if err != nil {
			panic(err.Error())
		}
		userIDs = append(userIDs, userID)
		if status == 0 {
			begin = false
		}

	}
	if len(userIDs) != 2 {
		begin = false
	}
	mysql.Close()
	return structs.CheckBeginStruct{Bool: begin, PlayersID: userIDs}
}

func getGameData(tableID int) structs.Desk {
	returnData := structs.Desk{}
	returnData.GameData = recreateGame(tableID)
	return returnData
}

func saveGameData(gameData []byte, tableID int) bool {
	mysql = openConnection()
	insForm, err := mysql.Prepare("INSERT INTO data (tableID, data) VALUES(?,?)")

	fancyHandleError(err)
	insForm.Exec(tableID, gameData)
	mysql.Close()
	return true
}

func sendToUsers(userIDs []int, data []byte) {
	for _, ID := range userIDs {
		if clients[ID] != nil {
			sendMessage(ID, data)
		}
	}
}

func sendMessage(userID int, data []byte) {
	err := clients[userID].WriteMessage(websocket.TextMessage, []byte(data))
	if err != nil {
		clients[userID].Close()
	}
}

func openConnection() *sql.DB {
	mysql, err := sql.Open("mysql", "swim:Password1!@tcp(51.158.173.204:3306)/backgammon")
	fancyHandleError(err)

	return mysql
}

func fancyHandleError(err error) (b bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log the where
		// the error happened, 0 = this function, we don't want that.
		pc, fn, line, _ := runtime.Caller(1)

		log.Printf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
		b = true
	}
	return
}

func rollOneDiceHandler(request *structs.Request) {
	var gameData = getGameData(request.TableID)
	if gameData.GameData.FirstDice.Rolled == nil {
		gameData.GameData.FirstDice.Rolled = make(map[int]int)
	}

	var rollerIndex = gameData.GameData.GetPlayerIndexByID(request.UserID)
	var winnerUserIndex = rollerIndex
	var status = 0

	if gameData.GameData.FirstDice.Rolled[rollerIndex] == 0 { // if user has not rolled yet.
		toSave := structs.SaveDataPayloadArray{
			Order: ORDERS.RollOneDice,
			Data:  append([]int{}, rollerIndex, generateRandom(1, 6)), Time: time.Now().Unix(),
		}
		byteCode, _ := json.Marshal(toSave)
		saveGameData(byteCode, request.TableID)
		sendToUsers(append([]int{}, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID), byteCode)

		gameData.GameData.FirstDice.Rolled[rollerIndex] = toSave.Data[1]

		if gameData.GameData.FirstDice.Rolled[1-rollerIndex] != 0 {
			if gameData.GameData.FirstDice.Rolled[1-rollerIndex] == gameData.GameData.FirstDice.Rolled[rollerIndex] {
				gameData.GameData.FirstDice.Rolled[1-rollerIndex] = 0
				gameData.GameData.FirstDice.Rolled[rollerIndex] = 0
				status = 1 // they are equal
			} else if gameData.GameData.FirstDice.Rolled[1-rollerIndex] > gameData.GameData.FirstDice.Rolled[rollerIndex] {
				winnerUserIndex = 1 - rollerIndex
			} else {
				winnerUserIndex = rollerIndex
			}
		} else {
			// second player has not rolled yet.
			status = 2
		}

		if status != 2 {
			byteCode, _ := json.Marshal(structs.RollOneDiceResponse{
				Order: ORDERS.RollOneDice,
				Payload: structs.RollOneDiceResponsePayload{
					Status: status,
					UserID: gameData.GameData.Players[winnerUserIndex].ID,
					Dice:   gameData.GameData.FirstDice.Rolled[winnerUserIndex]},
			})
			sendToUsers(append([]int{}, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID), byteCode)
		}
	} else { // user rolled one dice before.
		status = 3
		byteCode, _ := json.Marshal(structs.RollOneDiceResponse{
			Order: ORDERS.RollOneDice,
			Payload: structs.RollOneDiceResponsePayload{
				Status: status,
				UserID: gameData.GameData.Players[winnerUserIndex].ID,
				Dice:   gameData.GameData.FirstDice.Rolled[winnerUserIndex]},
		})

		sendMessage(request.UserID, byteCode)
	}
	// go printtable(req.TableID, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID, 30, 3)
}

func newConnectionHandler(request *structs.Request, ws *websocket.Conn) {
	mysql = openConnection()

	go pingConnection(request.TableID, request.UserID)
	var byteCode []byte
	var status int
	var RUQ int
	var available bool

	tableRow := mysql.QueryRow(
		`SELECT
			tablesStatus.status,
			RUQ
		FROM
			tables
			INNER JOIN tablesStatus ON tables.id = tablesStatus.tableID
		WHERE
			tables.id = ?
		ORDER BY
			tablesStatus.id DESC
		LIMIT 1`, request.TableID)
	err := tableRow.Scan(&status, &RUQ)
	fancyHandleError(err)

	if status == 0 {
		mysql.Query("INSERT INTO usersStatus (tableID, userID, status) VALUES (?, ?, 1)", request.TableID, request.UserID)
		row := mysql.QueryRow("SELECT IF((SELECT SUM(usersStatus.status)FROM usersStatus WHERE usersStatus.id IN(SELECT MAX(usersStatus.id) AS id FROM usersStatus WHERE usersStatus.tableID = ? GROUP BY usersStatus.userID))=?,'true','false') AS available", request.TableID, RUQ)
		err := row.Scan(&available)
		fancyHandleError(err)

		if available {
			mysql.Query("INSERT INTO tablesStatus  VALUES (NULL, ?, 1)", request.TableID)
			rowa, _ := mysql.Query("SELECT DISTINCT usersStatus.userID FROM usersStatus WHERE tableID = ?", request.TableID)
			var userID int
			byteCode, _ = json.Marshal(structs.CheckBeginResponse{Order: request.Order, Answer: true})
			for rowa.Next() {
				err := rowa.Scan(&userID)
				fancyHandleError(err)

				sendMessage(userID, byteCode)
			}
		} else {
			byteCode, _ = json.Marshal(structs.CheckBeginResponse{Order: request.Order, Answer: false})
			sendMessage(request.UserID, byteCode)
		}
	} else if status == 1 {
		var gameData = getGameData(request.TableID)
		gameData.Order = 11 // TODO this is for a little time.
		bytecode, _ := json.Marshal(gameData)
		sendMessage(request.UserID, bytecode)
	} else if status == 2 {
		row := mysql.QueryRow("SELECT IF((SELECT status FROM usersStatus WHERE usersStatus.tableID= ? AND usersStatus.userID= ? ORDER BY id DESC LIMIT 1)=2,'true','false') AS available", request.TableID, request.UserID)
		err := row.Scan(&available)
		fancyHandleError(err)
		if available {
			mysql.Query("INSERT INTO usersStatus (tableID, userID, status) VALUES (?, ?, 1)", request.TableID, request.UserID)
			row := mysql.QueryRow("SELECT IF((SELECT SUM(usersStatus.status)FROM usersStatus WHERE usersStatus.id IN(SELECT MAX(usersStatus.id)AS id FROM usersStatus WHERE usersStatus.tableID=? GROUP BY usersStatus.userID))=?,'true','false')AS available", request.TableID, RUQ)
			err := row.Scan(&available)
			fancyHandleError(err)

			if available {
				mysql.Query("INSERT INTO tablesStatus VALUES (NULL, ?, 1)", request.TableID)

				data := getGameData(request.TableID)
				data.Time = time.Now().Unix()
				data.Order = 11
				byteCode, _ = json.Marshal(data)
				sendMessage(request.UserID, byteCode)
				sendToUsers(append([]int{}, data.GameData.Players[0].ID, data.GameData.Players[1].ID), reconnectNotification) // notify users that game resumes
			} else {
				// Still not enough players.
			}
		} else {
			// Person is not available to join this table because table is paused and it's ID is not in table's user's list.
		}
	}
}

func rollDiceHandler(request *structs.Request) {
	var gameData = getGameData(request.TableID)
	if gameData.GameData.Rolled[0].Rolled == 0 {
		fmt.Println("there was not dice rolled yet")
		if gameData.GameData.Begginer == request.UserID { // there is no dice rolled, and now first time
			gameData.GameData.Roller = request.UserID
		} else {
			fmt.Println("You are not allowed to throw first dice")
			return
		}
	} else {
		fmt.Println("there was dice rolled before this")
		if gameData.GameData.RollQuantity == 0 {
			if gameData.GameData.Roller == request.UserID {
				fmt.Println("Not you turn to roll")
				return
			}
			// should be able to roll;

		} else {
			if gameData.GameData.Roller == request.UserID {
				fmt.Println("you should play what you rolled : )")
				return
			}
			index := gameData.GameData.CurrentPlayerIndex()
			if gameData.GameData.KilledCoins[index].Coins == 0 {
				fmt.Println("you are not allowed to roll because opponent has to roll")
				return
			}
			if gameData.GameData.CurrentPlayer().Direction {
				if (gameData.GameData.Tunnels[gameData.GameData.Rolled[0].Rolled-1].Coins > 1 && gameData.GameData.Rolled[0].Used == 0) ||
					(gameData.GameData.Tunnels[gameData.GameData.Rolled[1].Rolled-1].Coins > 1 && gameData.GameData.Rolled[1].Used == 0) {
					fmt.Println("user can not wake his coin, so you can roll")
					// allow roll
					// if gameData.GameData.Tunnels[gameData.GameData.Rolled[0].Rolled-1].O == gameData.GameData.CurrentPlayer().ID {
					//
					// } else {
					// }
				} else {
					fmt.Println("user can wake his coin so you should wait for it.")
					return
				}
			} else {
				if (gameData.GameData.Tunnels[23-gameData.GameData.Rolled[0].Rolled+1].Coins > 1 && gameData.GameData.Rolled[0].Used == 0) ||
					(gameData.GameData.Tunnels[23-gameData.GameData.Rolled[1].Rolled+1].Coins > 1 && gameData.GameData.Rolled[1].Used == 0) {
					fmt.Println("user can not wake his coin, so you can roll")
					// allow roll
					// if gameData.GameData.Tunnels[gameData.GameData.Rolled[0].Rolled-1].O == gameData.GameData.CurrentPlayer().ID {
					//
					// } else {
					// }
				} else {
					fmt.Println("user can wake his coin so you should wait for it.")
					return
				}
			}

			// here should be check if player can move or not,
			// if not opponent should be able to roll dice. or maybe this should be automatically
		}
	}

	rollDice(&gameData)
	returnData := structs.SaveDataPayloadArray{Order: ORDERS.RollDice, Time: time.Now().Unix()}
	returnData.Data = append(returnData.Data, gameData.GameData.Rolled[0].Rolled, gameData.GameData.Rolled[1].Rolled)
	byteCode, _ := json.Marshal(returnData)
	saveGameData(byteCode, request.TableID)
	cancelTimer(request.TableID)
	// gameData.sendMessageToPlayers();
	// go printtable(request.TableID, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID, 30, 3)
	sendToUsers(append([]int{}, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID), byteCode)
}

func coinOutHandler(request *structs.Request) {
	var gameData = getGameData(request.TableID)
	allowedMove := gameData.GameData.CoinGoesOut(request.Payload.From)
	if allowedMove {
		cancelTimer(request.TableID)
		// gameData.GameData = allowedMove.Data // TODO check later
		toSave := structs.SaveDataPayloadNumber{Order: ORDERS.CoinOut, Data: request.Payload.From}
		byteCode, _ := json.Marshal(toSave)
		saveGameData(byteCode, request.TableID)

		returnDataOther := structs.CoinMoved{
			Order: ORDERS.CoinOut,
			Payload: structs.FromTo{
				From: request.Payload.From,
				To:   0,
			},
		}
		//TODO create good sctuct instead of coinMoved which will have only From value.
		byteCode, _ = json.Marshal(returnDataOther)
		// go printtable(request.TableID, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID, 30, 3)
		sendToUsers(append([]int{}, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID), byteCode)
	} else {
		returnData := structs.MessageMOVE_COIN{Order: ORDERS.CoinOut, Success: allowedMove, Time: 0}
		byteCode, _ := json.Marshal(returnData)
		sendMessage(request.UserID, byteCode)
	}
}

func moveCoinHandler(request *structs.Request) {
	var gameData = getGameData(request.TableID)
	allowedMove := gameData.GameData.IsAccessibleCoin(request.Payload.From, request.Payload.To, request.UserID)
	if allowedMove {
		cancelTimer(request.TableID)
		// gameData.Order = request.Order
		// gameData.GameData = allowedMove.Data
		saveData := structs.SaveDataPayloadArray{
			Order: ORDERS.MoveCoin,
			Time:  time.Now().Unix(),
		}
		saveData.Data = append(saveData.Data, request.Payload.From, request.Payload.To)
		byteCode, _ := json.Marshal(saveData)
		saveGameData(byteCode, request.TableID)

		returnDataOther := structs.CoinMoved{
			Order: ORDERS.CoinMoved,
			Payload: structs.FromTo{
				From: request.Payload.From,
				To:   request.Payload.To},
			Time: time.Now().Unix(),
		}
		byteCode, _ = json.Marshal(returnDataOther)
		// go printtable(request.TableID, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID, 30, 3)
		sendMessage(gameData.GameData.Players[0].ID, byteCode)
		sendMessage(gameData.GameData.Players[1].ID, byteCode)
		// go printtable(request.TableID, gameData.GameData.Players[0].ID, gameData.GameData.Players[0].ID, 30, 7)
	} else {
		// აქ უნდა ჩაემატოს რომ თუ რაღაცას ატრაკებს იუზერი ფრონტის ჩეკი თუ გაიარა რა ანუ, მარტო ამ იუზერს უნდა დავუბრუნოთ პასუხი. ამწუთას ასე არაა.
		returnData := structs.MessageMOVE_COIN{Order: ORDERS.MoveCoin, Success: allowedMove, Time: 0}
		byteCode, _ := json.Marshal(returnData)
		sendMessage(request.UserID, byteCode)
	}
}

func awakeCoinHandler(request *structs.Request) {
	var gameData = getGameData(request.TableID)
	allowedMove := gameData.GameData.MakeCoinAlive(request.Payload.To)
	if allowedMove {
		cancelTimer(request.TableID)
		// gameData.Order = request.Order
		// gameData.GameData = allowedMove.Data
		toSave := structs.SaveDataPayloadNumber{
			Order: ORDERS.AwakeCoin,
			Data:  request.Payload.To,
		}

		byteCode, _ := json.Marshal(toSave)
		saveGameData(byteCode, request.TableID)

		returnDataOther := structs.CoinMoved{
			Order: ORDERS.AwakeCoin,
			Payload: structs.FromTo{
				From: request.UserID,
				To:   request.Payload.To,
			},
			Time: time.Now().Unix(),
		}
		//TODO create good struct
		byteCode, _ = json.Marshal(returnDataOther)
		// go printtable(request.TableID, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID, 30, 3)

		sendMessage(gameData.GameData.Players[0].ID, byteCode)
		sendMessage(gameData.GameData.Players[1].ID, byteCode)
	} else {
		// აქ უნდა ჩაემატოს რომ თუ რაღაცას ატრაკებს იუზერი ფრონტის ჩეკი თუ გაიარა რა ანუ, მარტო ამ იუზერს უნდა დავუბრუნოთ პასუხი. ამწუთას ასე არაა.
		returnData := structs.MessageMOVE_COIN{Order: ORDERS.AwakeCoin, Success: allowedMove, Time: 0}
		byteCode, _ := json.Marshal(returnData)
		sendMessage(request.UserID, byteCode)
	}
}

func beginGameHandler(req *structs.Request) {
	//TODO aq albat table id unda miigo da naxos bazashi ra userebzea table.
	var gameData = generateData(req.TableID)
	gameData.Time = time.Now().Unix()
	byteCode, _ := json.Marshal(gameData)
	saveGameData(byteCode, req.TableID)
	sendToUsers(append([]int{}, gameData.GameData.Players[0].ID, gameData.GameData.Players[1].ID), byteCode)
}

func checkBeginHandler(request *structs.Request) {
	var res = checkBegin(request.TableID)
	var tableStatusResponse = checkTableStatus(request.TableID)
	var byteCode []byte
	if tableStatusResponse.Status == 0 {
		byteCode, _ = json.Marshal(structs.CheckBeginResponse{Order: 3, Answer: res.Bool})
	} else if tableStatusResponse.Status == 1 {
		data := getGameData(request.TableID)
		byteCode, _ = json.Marshal(data)
	}
	sendToUsers(res.PlayersID, byteCode)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// fmt.Println(err)
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		req := structs.Request{}

		json.Unmarshal([]byte(msg), &req)

		if req.Order == ORDERS.Connection {
			clients[req.UserID] = ws // save websocket connection in client's list to access it with userid.
			newConnectionHandler(&req, ws)
		} else if req.Order == ORDERS.CheckBegin {
			checkBeginHandler(&req)
		} else if req.Order == ORDERS.BeginGame {
			beginGameHandler(&req)
		} else if req.Order == ORDERS.RollDice {
			rollDiceHandler(&req)
		} else if req.Order == ORDERS.AwakeCoin {
			awakeCoinHandler(&req)
		} else if req.Order == ORDERS.MoveCoin {
			moveCoinHandler(&req)
		} else if req.Order == ORDERS.CoinOut {
			coinOutHandler(&req)
		} else if req.Order == ORDERS.RollOneDice {
			rollOneDiceHandler(&req)
		}
	}
}
