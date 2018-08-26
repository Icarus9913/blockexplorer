package Reorg

import (
	"bytes"
	"encoding/gob"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/go-errors/errors"
	"github.com/romanornr/cyberchain/database"
	"fmt"
	"container/list"
)

var db = database.GetDatabaseInstance()

type Observer interface {
	Update()
}

type Subject interface {
	Attach (observer Observer)
	Detach (observer Observer)
	Notify()
}

type DefaultSubject struct {
	observer *list.List
}

func NewDefaultSubject() *DefaultSubject {
	return &DefaultSubject{observer:new(list.List)}
}

func (this *DefaultSubject) Attach(observer Observer) {
	this.observer.PushBack(observer)
}

func (this *DefaultSubject) Detach(observer Observer) {
	for obs := this.observer.Front(); obs != nil; obs = obs.Next() {
		if obs.Value.(Observer) == observer {
			this.observer.Remove(obs)
		}
	}
}

func (this *DefaultSubject) Notify() {
	for obs := this.observer.Front(); obs != nil; obs = obs.Next() {
		observer := obs.Value.(Observer)
		observer.Update()
	}
}

//
type GameState string

func NewGameState(state string) *GameState {
	gs := GameState(state)
	return &gs
}

type Game struct {
	*DefaultSubject
	state *GameState
}

func NewGame() *Game {
	return &Game{DefaultSubject:NewDefaultSubject()}
}

func (this *Game) GetState() *GameState {
	return this.state
}

func (this *Game) SetState(state *GameState) {
	this.state = state
	this.Notify()
}

type Player struct {
	name         string
	lastState *GameState
	game         *Game
}

func NewPlayer(name string, game *Game) *Player {
	this := new(Player)
	this.name = name
	this.game = game
	return this
}

func (this *Player) Update() {
	this.lastState = this.game.GetState()
	fmt.Println(this.name, "noticed that game state has changed to: ", *this.lastState)
}
//


func Check(newBlock *btcjson.GetBlockVerboseResult) error {

	var lastBlock *btcjson.GetBlockVerboseResult
	_, lastBlockInDatabase := database.GetLastBlock(database.GetDatabaseInstance())
	decoder := gob.NewDecoder(bytes.NewReader(lastBlockInDatabase))
	decoder.Decode(&lastBlock)

	if lastBlock.Height == newBlock.Height {
		if lastBlock.Hash != newBlock.Hash {
			return errors.Errorf("reorg detected ! last block in DB %d %s\n new : block %d %s", lastBlock.Height, lastBlock.Hash, newBlock.Height, newBlock.Hash)
		}
		return nil
	}

	duplicateBlockHeight := database.FetchBlockHashByBlockHeight(newBlock.Height)

	if duplicateBlockHeight != nil {
		fmt.Println(newBlock.Hash)

		var oldBlock *btcjson.GetBlockVerboseResult
		decoder = gob.NewDecoder(bytes.NewReader(database.ViewBlock(string(duplicateBlockHeight))))
		decoder.Decode(&oldBlock)

		if oldBlock.Hash != newBlock.Hash{
			return errors.Errorf("reorg detected ! Block in DB %d %s\n new : block %d %s", oldBlock.Height, oldBlock.Hash, newBlock.Height, newBlock.Hash)
		}
	}


	return nil

	//h1 := "ded7508b6b6452bfc99961366e3206a7a258cf897d3148b46e590bbf6f23f3d9"
	//h2 := "e8957dac3477849c431dce6929e45ca829598bf45f05f776742f04f06c246ae7"
	//a, _ := chainhash.NewHashFromStr(h1)
	//b, _ := chainhash.NewHashFromStr(h2)
	//
	//fmt.Print(GetCommonBlockAncestorHeight(b, a))

}

//// GetCommonBlockAncestorHeight takes in:
//// (1) the hash of a block that has been reorged out of the main chain
//// (2) the hash of the block of the same height from the main chain
//// It returns the height of the nearest common ancestor between the two hashes,
//// or an error
//func GetCommonBlockAncestorHeight(reorgHash, chainHash *chainhash.Hash) (int32, error) {
//	fmt.Print(chainHash)
//
//	for reorgHash != chainHash {
//		reorgHeader, err := blockdata.GetBlockHeader(reorgHash)
//		if err != nil {
//			return 0, fmt.Errorf("unable to get reorg header for hash=%v: %v\n",
//				reorgHash, err)
//		}
//		chainHeader, err := blockdata.GetBlockHeader(chainHash)
//		if err != nil {
//			return 0, fmt.Errorf("unable to get header for hash=%v: %v\n",
//				chainHash, err)
//		}
//		reorgHash,_ = chainhash.NewHashFromStr(reorgHeader.PreviousHash)
//		chainHash, _ = chainhash.NewHashFromStr(chainHeader.PreviousHash)
//		//reorgHash = reorgHeader.PrevBlock
//		//chainHash = chainHeader.PrevBlock
//	}
//
//	verboseHeader, err := blockdata.GetBlockHeaderVerbose(chainHash)
//	if err != nil {
//		return 0, fmt.Errorf("unable to get verbose header for hash=%v: %v",
//			chainHash, err)
//	}
//
//	return verboseHeader.Height, nil
//}