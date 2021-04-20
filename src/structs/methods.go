package structs

import (
	"fmt"
	"math"
)

// TODO MoveCoinToKilled method
func (g GameData) MoveCoinToKilled(from int) {
	if g.Tunnels[from].Coins != 0 && g.Tunnels[from].O != g.CurrentPlayer().ID {
		g.KilledCoins[1-g.CurrentPlayerIndex()].Coins++
		g.Tunnels[from].Coins--
	}
}

// TODO AwakeCoin method
func (g GameData) AwakeCoin(from int, to int) {
	g.MoveCoinToKilled(to) // TODO mistake
	g.Tunnels[to].Coins++
	g.KilledCoins[from].Coins--
	g.Tunnels[to].O = g.KilledCoins[from].O
}

// MakeCoinAlive Method
func (g GameData) MakeCoinAlive(to int) bool {
	if !g.TunnelAccessible(to) {
		return false // Good
	}
	player := g.CurrentPlayer()

	i := 0
	if player.Direction {
		i = 23 - to + 1 // GOOD
	} else {
		i = to + 1
	}

	return (g.Rolled[0].Rolled == i && g.Rolled[0].Used == 0) ||
		(g.Rolled[1].Rolled == i && g.Rolled[1].Used == 0)
}

// ChangeTunnels method
func (g GameData) ChangeTunnels(from int, to int) {
	g.Tunnels[to].Coins++               // adds coin to target tunnel
	g.Tunnels[from].Coins--             // removes coin from source tunnel
	g.Tunnels[to].O = g.Tunnels[from].O // changes owner of target tunnel
}

// IsMovable method
//TODO name this function better
func (g GameData) IsMovable(from int, to int) bool {
	return g.Tunnels[to].Coins <= 1 ||
		g.Tunnels[to].O == g.Tunnels[from].O
}

// TunnelAccessible method
func (g GameData) TunnelAccessible(tunnelIndex int) bool {
	currentPlayer := g.CurrentPlayer()
	return g.Tunnels[tunnelIndex].Coins == 0 ||
		g.Tunnels[tunnelIndex].Coins == 1 ||
		g.Tunnels[tunnelIndex].O == currentPlayer.ID
}

// KillCoin method
func (g GameData) KillCoin(from int, to int) {
	currentIndex := g.CurrentPlayerIndex()
	g.KilledCoins[1-currentIndex].Coins++ // adds coin to opponents died coins list.
	g.Tunnels[to].O = g.Tunnels[from].O   // changes tunnel owners
	g.Tunnels[from].Coins--               // removes coin from source tunnel
}

// IsEveryBodyHome method
func (g GameData) IsEveryBodyHome() bool {
	currentPlayer := g.CurrentPlayer()
	var count int
	currentPlayerIndex := g.CurrentPlayerIndex()
	count = g.OutCoins[currentPlayerIndex].Coins

	start := 0
	end := 5

	if currentPlayer.Direction == false {
		start = 18
		end = 23
	}

	for i := start; i <= end; i++ {
		if g.Tunnels[i].Coins != 0 && g.Tunnels[i].O == currentPlayer.ID {
			count += g.Tunnels[i].Coins
		}
	}

	return count == 15
}

// MoveCoin method
func (g GameData) MoveCoin(from int, to int) {
	//TODO from and to tunnels should be roller's
	if g.Tunnels[to].Coins != 0 {
		if g.Tunnels[to].O == g.Tunnels[from].O {
			g.ChangeTunnels(from, to) // GOOD
		} else {
			if g.Tunnels[to].Coins == 1 {
				g.KillCoin(from, to) // GOOD
			}
		}
	} else {
		g.ChangeTunnels(from, to) // GOOD
	}
}

// ThrowCoin method
func (g GameData) ThrowCoin(from int) {
	g.OutCoins[g.CurrentPlayerIndex()].Coins++
	g.Tunnels[from].Coins--
}

// GetPlayerIndexByID method
func (g GameData) GetPlayerIndexByID(ID int) int {
	if g.Players[0].ID == ID {
		return 0
	}
	return 1
}

// CurrentPlayerIndex method
func (g GameData) CurrentPlayerIndex() int {
	if g.Players[0].ID == g.Roller {
		return 0
	}
	return 1
}

// CurrentPlayer method
func (g GameData) CurrentPlayer() Player {
	if g.Players[0].ID == g.Roller {
		return g.Players[0]
	}
	return g.Players[1]
}

// GetPlayerByID method
func (g GameData) GetPlayerByID(ID int) *Player {
	if g.Players[0].ID == ID {
		return &g.Players[0]
	}
	return &g.Players[1]
}

// IsAccessibleCoin method
func (g GameData) IsAccessibleCoin(from int, to int, userID int) bool {
	if g.KilledCoins[g.CurrentPlayerIndex()].Coins != 0 { // user has dead coin
		return false
	}

	delta := int(math.Abs(float64(to) - float64(from)))
	currentPlayer := g.CurrentPlayer()

	if userID != currentPlayer.ID { // check if player can move
		return false
	}

	if g.Tunnels[from].Coins == 0 { // check if source tunnel is empty or not
		return false
	}

	if currentPlayer.ID != g.Tunnels[from].O { // check if this is your coin
		return false
	}

	if currentPlayer.Direction == true && from < to { // check if user moves coin with correct destination
		return false
	} else if currentPlayer.Direction == false && to < from {
		return false
	}

	if g.Quadro { // checks if user rolled equal dices
		if delta%g.Rolled[0].Rolled == 0 { // თუ ნამდვილად გაგორებულის ჯერადზე გადადის.
			if delta/g.Rolled[0].Rolled <= g.RollQuantity { // თუ ყოფნის სვლების რაოდენობა.
				for index := 0; index < (delta / g.Rolled[0].Rolled); index++ {
					if currentPlayer.Direction {
						if !g.IsMovable(from, from-g.Rolled[0].Rolled*(index+1)) {
							return false
						}
					} else {
						if !g.IsMovable(from, from+g.Rolled[0].Rolled*(index+1)) {
							return false
						}
					}
				}

				g.RollQuantity -= delta / g.Rolled[0].Rolled
				g.MoveCoin(from, to)
				return true
			}
			return false

		}
		return false
	}

	if delta != g.Rolled[0].Rolled+g.Rolled[1].Rolled &&
		delta != g.Rolled[0].Rolled &&
		delta != g.Rolled[1].Rolled { // user wants to move on uncorrect destination
		return false
	}

	correctDirection := from < to
	rolledDice := g.Rolled[0].Rolled
	rolledDice1 := g.Rolled[1].Rolled
	if currentPlayer.Direction {
		correctDirection = from > to
		rolledDice *= -1
		rolledDice1 *= -1
	}

	if correctDirection {
		if g.Rolled[0].Used == 0 &&
			g.Rolled[1].Used == 0 &&
			delta == g.Rolled[0].Rolled+g.Rolled[1].Rolled {
			if g.IsMovable(from, to) {
				if g.IsMovable(from, from+rolledDice) ||
					g.IsMovable(from, from+rolledDice1) {
					// g.MoveCoin(from, to)
					// g.Rolled[0].Used = 1
					// g.Rolled[1].Used = 1
					// g.RollQuantity = 0
					return true
				}
				return false
			}
		} else if g.Rolled[0].Used == 0 &&
			g.Rolled[0].Rolled == delta &&
			g.IsMovable(from, to) {
			// g.MoveCoin(from, to)
			// g.Rolled[0].Used = 1
			g.RollQuantity--
			return true
		} else if g.Rolled[1].Used == 0 &&
			g.Rolled[1].Rolled == delta &&
			g.IsMovable(from, to) {
			// g.MoveCoin(from, to)
			// g.Rolled[1].Used = 1
			g.RollQuantity--
			return true
		}
	} else {
		return false
	}

	return false
}

// CoinGoesOut method
func (g GameData) CoinGoesOut(from int) bool {
	if g.IsEveryBodyHome() == false {
		fmt.Println("not every coin home")
		return false
	}

	currentPlayer := g.CurrentPlayer()

	if g.Tunnels[from].O != currentPlayer.ID || g.Tunnels[from].Coins == 0 {
		fmt.Println("not it's coin")
		return false
	}

	sourceTunnel := 0 // RANGE [1-6]
	ff := 1
	testVariable := 0
	if currentPlayer.Direction == true {
		sourceTunnel = from + 1
		testVariable = 5
		ff = -1
	} else {
		sourceTunnel = 23 - from + 1
		testVariable = 18
		ff = 1
	}

	if g.Rolled[0].Used == 0 && g.Rolled[0].Rolled >= sourceTunnel {
		if g.Rolled[0].Rolled == sourceTunnel {
			return true
		}
		for index := testVariable; index != from; index += ff {
			if g.Tunnels[index].Coins != 0 && // if there is coin on tunnel
				g.Tunnels[index].O == currentPlayer.ID { // and if it'his coin.
				fmt.Println("Here comes coin out message1")
				return false
			}
		}
	} else if g.Rolled[1].Used == 0 && g.Rolled[1].Rolled >= sourceTunnel { // GOOD
		if g.Rolled[1].Rolled == sourceTunnel {
			return true
		}
		for index := testVariable; index != from; index += ff {
			if g.Tunnels[index].Coins != 0 &&
				g.Tunnels[index].O == currentPlayer.ID {
				fmt.Println("Here comes coin out message2")
				return false
			}
		}
	} else {
		return false
	}

	return true
}

func (dice Dice) checkDices() bool {
	return true
}
