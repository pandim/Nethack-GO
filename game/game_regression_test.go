package game

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestGainXPSupportsMultipleLevelUps(t *testing.T) {
	player := NewPlayer(0, 0)

	if !player.GainXP(100) {
		t.Fatal("GainXP should report a level up")
	}
	if player.Level != 3 {
		t.Fatalf("expected level 3, got %d", player.Level)
	}
	if player.XP != 40 {
		t.Fatalf("expected 40 remaining XP, got %d", player.XP)
	}
}

func TestRespawnMonstersKeepsAliveBoss(t *testing.T) {
	level := &Level{
		Width:      5,
		Height:     5,
		Depth:      1,
		Tiles:      make([][]Tile, 5),
		Rooms:      make([]Room, 0),
		Items:      make([]*Item, 0),
		VisitCount: 1,
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]Tile, 5)
		for x := range level.Tiles[y] {
			level.Tiles[y][x].Type = TileFloor
		}
	}

	boss := NewBoss(1, 1, "Босс", 10, 1, 1, 1, 'B', 0, BossAbilityNone)
	level.Monsters = []*Monster{
		boss,
		NewMonster(2, 2, "Монстр", 1, 1, 1, 1, 'm', 0),
	}

	level.respawnMonsters()

	foundBoss := false
	for _, monster := range level.Monsters {
		if monster == boss {
			foundBoss = true
			break
		}
	}
	if !foundBoss {
		t.Fatal("alive boss was removed during respawn")
	}
}

func TestMessagesKeepHistoryAndScroll(t *testing.T) {
	game := &Game{}
	for i := 0; i < messageHeight+2; i++ {
		game.addMessage(string(rune('A' + i)))
	}

	game.messageScroll = messageHeight
	if !game.handleMessageScroll(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)) {
		t.Fatal("page down should be handled")
	}
	if game.messageScroll != 0 {
		t.Fatalf("expected scroll to return to bottom, got %d", game.messageScroll)
	}

	if !game.handleMessageScroll(tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModNone)) {
		t.Fatal("left bracket should scroll up")
	}
	expectedScroll := len(game.messages) - messageHeight
	if game.messageScroll != expectedScroll {
		t.Fatalf("expected fallback scroll to move up, got %d", game.messageScroll)
	}
}
