package structs

type Orders struct {
	//TODO must rewatch those codes.
	Connection, CheckBegin, BeginGame, RollDice, RollOneDice, MoveCoin, CoinMoved, AwakeCoin, CoinOut int
}

type StatusCodes struct {
	//TODO there is no any status code available now. but is required for future!
	NotYourMove, NotYourCoin, NotYourRoll, NotYourTunnel, BadDestination, YouHaveDiedCoin, CoinMoved, AwakeCoin int
}

type Coin struct {
}

type Player struct {
	ID        int
	Direction bool
}

type Dice struct {
	Rolled int
	Used   uint
}

type Tunnel struct {
	Coins int `json:"Coins"`
	O     int
}

type TestStruct struct {
	Order   int    `json:"o"`
	Payload string `json:"p"`
}

type RollOneDice struct {
	Order   int                `json:"o"`
	Payload RollOneDicePayload `json:"p"`
}

type RollOneDicePayload struct {
	RollerIndex int
	Dice        int
}

type RollOneDiceResponse struct {
	Order   int                        `json:"o"`
	Payload RollOneDiceResponsePayload `json:"p"`
}

type RollOneDiceResponsePayload struct {
	Status int `json:"status"`
	UserID int `json:"userid"`
	Dice   int `json:"dice"`
}

type TestStruct1 struct {
	Order   int      `json:"o"`
	Payload GameData `json:"p"`
}

type TestStruct2 struct {
	Order   int   `json:"o"`
	Payload []int `json:"p"`
}
type TestStruct3 struct {
	Order   int `json:"o"`
	Payload int `json:"p"`
}

type Request struct {
	Order   int    `json:"o"`
	Payload FromTo `json:"p"`
	UserID  int    `json:"uID"`
	TableID int    `json:"tID"`
}

type FromTo struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type Desk struct {
	Order    int      `json:"o"`
	GameData GameData `json:"p"`
	Time     int64    `json:"time"`
}

type Timer struct {
	TimeOut *Timer
}

type SaveDataPayloadNumber struct {
	Order int   `json:"o"`
	Data  int   `json:"p"`
	Time  int64 `json:"time"`
}

type SaveDataPayloadArray struct {
	Order int   `json:"o"`
	Data  []int `json:"p"`
	Time  int64 `json:"time"`
}

type MessageAwake struct {
	Order int   `json:"o"`
	Data  int   `json:"p"`
	Time  int64 `json:"time"`
}

type MessageMOVE_COIN struct {
	Order   int   `json:"o"`
	Success bool  `json:"payload"`
	Time    int64 `json:"time"`
}

type GameData struct {
	Tunnels      []Tunnel  `json:"t"`
	KilledCoins  []Tunnel  `json:"kc"`
	OutCoins     []Tunnel  `json:"oc"`
	Players      []Player  `json:"p"`
	Rolled       []Dice    `json:"rd"`
	Quadro       bool      `json:"q"`
	RollQuantity int       `json:"rq"`
	Roller       int       `json:"ro"`
	FirstDice    FirstDice `json:"-"`
	Begginer     int       `json:"begginer"`
}

type FirstDice struct {
	Rolled map[int]int
}

type CheckBeginStruct struct {
	Bool      bool  `json:"bgn"`
	PlayersID []int `json:"pID"`
}

type CheckTableStatus struct {
	Status int `json:"status"`
}

type CheckBeginResponse struct {
	Order  int  `json:"o"`
	Answer bool `json:"answer"`
}

type CoinMoved struct {
	Order   int    `json:"o"`
	Payload FromTo `json:"payload"`
	Time    int64  `json:"time"`
}
