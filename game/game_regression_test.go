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

func TestPickupEquipmentAutoEquipsAndUpgrades(t *testing.T) {
	player := NewPlayer(0, 0)
	level := &Level{Items: make([]*Item, 0)}
	game := &Game{player: player, level: level}

	firstSword := NewItem(0, 0, "Меч", ItemTypeWeapon, 7, '/', tcell.ColorYellow)
	level.Items = append(level.Items, firstSword)
	game.pickupItem(firstSword)
	if player.EquippedWeapon == nil || player.EquippedWeapon.Value != 7 {
		t.Fatalf("expected first sword to be equipped with value 7")
	}
	if len(player.Inventory) != 0 || len(level.Items) != 0 {
		t.Fatalf("picked-up sword should not remain in inventory or on level")
	}

	secondSword := NewItem(0, 0, "Меч", ItemTypeWeapon, 17, '/', tcell.ColorYellow)
	level.Items = append(level.Items, secondSword)
	game.pickupItem(secondSword)
	if player.EquippedWeapon.Value != 10 {
		t.Fatalf("expected second sword to add 3 upgrade, got %d", player.EquippedWeapon.Value)
	}
	if len(player.Inventory) != 0 || len(level.Items) != 0 {
		t.Fatalf("upgrading sword should not create an inventory item")
	}

	shield := NewItem(0, 0, "Щит", ItemTypeArmor, 5, '[', tcell.ColorBlue)
	level.Items = append(level.Items, shield)
	game.pickupItem(shield)
	if player.EquippedArmor == nil || player.EquippedArmor.Value != 5 {
		t.Fatalf("expected shield to be equipped with value 5")
	}
	if len(player.Inventory) != 0 || len(level.Items) != 0 {
		t.Fatalf("picked-up shield should not remain in inventory or on level")
	}
}

func TestBuyEquipmentAutoEquipsAndUpgrades(t *testing.T) {
	player := NewPlayer(0, 0)
	player.Gold = 300
	game := &Game{player: player}

	game.currentMerchant = &Merchant{Items: []*Item{
		NewItem(0, 0, "Меч", ItemTypeWeapon, 7, '/', tcell.ColorYellow),
	}}
	game.currentMerchant.Items[0].Price = 50
	game.buyItem(0)
	if player.EquippedWeapon == nil || player.EquippedWeapon.Value != 7 || len(player.Inventory) != 0 {
		t.Fatalf("expected purchased sword to be equipped, not stored in inventory")
	}

	game.currentMerchant = &Merchant{Items: []*Item{
		NewItem(0, 0, "Меч", ItemTypeWeapon, 17, '/', tcell.ColorYellow),
	}}
	game.currentMerchant.Items[0].Price = 50
	game.buyItem(0)
	if player.EquippedWeapon.Value != 10 || len(player.Inventory) != 0 {
		t.Fatalf("expected purchased sword to upgrade equipped weapon")
	}

	game.currentMerchant = &Merchant{Items: []*Item{
		NewItem(0, 0, "Щит", ItemTypeArmor, 5, '[', tcell.ColorBlue),
	}}
	game.currentMerchant.Items[0].Price = 40
	game.buyItem(0)
	if player.EquippedArmor == nil || player.EquippedArmor.Value != 5 || len(player.Inventory) != 0 {
		t.Fatalf("expected purchased shield to be equipped, not stored in inventory")
	}
}

func TestIdenticalRelicsStackAndSellIndividually(t *testing.T) {
	player := NewPlayer(0, 0)
	game := &Game{player: player}

	first := NewRelic(0, 0, RelicDiamond, "Алмаз", 100, '*', tcell.ColorWhite)
	second := NewRelic(0, 0, RelicDiamond, "Алмаз", 100, '*', tcell.ColorWhite)
	other := NewRelic(0, 0, RelicCrown, "Корона гоблинов", 200, '*', tcell.ColorWhite)
	game.addToInventoryWithStack(first)
	game.addToInventoryWithStack(second)
	game.addToInventoryWithStack(other)

	if len(player.Inventory) != 2 || player.Inventory[0].Count != 2 {
		t.Fatalf("expected identical relics to form one stack")
	}
	if player.CountRelics() != 2 {
		t.Fatalf("expected two unique relics, got %d", player.CountRelics())
	}

	game.sellRelic(0)
	if player.Gold != 100 || player.Inventory[0].Count != 1 {
		t.Fatalf("expected one relic sold from stack, gold=%d count=%d", player.Gold, player.Inventory[0].Count)
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

	game.messageScroll = 2
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
