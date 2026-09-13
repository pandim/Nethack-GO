package game

import (
	"math/rand/v2"

	"github.com/gdamore/tcell/v2"
)

const (
	MaxSpawnAttempts = 100
)

// =============================================================================
// СПАВН МОНСТРОВ
// =============================================================================
func (l *Level) spawnMonsters(count int) {
	if l == nil || count <= 0 || l.Width < 3 || l.Height < 3 {
		return
	}
	monsterTypes := []struct {
		name                 string
		hp, attack, gold, xp int
		symbol               rune
		color                tcell.Color
	}{
		{"Гоблин", 8, 2, 5, 8, 'g', tcell.ColorGreen},
		{"Орк", 12, 3, 10, 15, 'o', tcell.ColorDarkRed},
		{"Скелет", 10, 2, 8, 12, 's', tcell.ColorWhite},
		{"Крыса", 4, 1, 2, 3, 'r', tcell.ColorBrown},
	}
	depth := l.Depth
	for i := 0; i < count; i++ {
		var x, y int
		attempts := 0
		for {
			x = 1 + rand.IntN(l.Width-2)
			y = 1 + rand.IntN(l.Height-2)
			if l.Tiles[y][x].Type == TileFloor && !l.hasMonsterAt(x, y) && !l.hasItemAt(x, y) && !l.isStairsAt(x, y) {
				break
			}
			attempts++
			if attempts > MaxSpawnAttempts {
				return
			}
		}
		mt := monsterTypes[rand.IntN(len(monsterTypes))]

		// 🆕 СБАЛАНСИРОВАННОЕ МАСШТАБИРОВАНИЕ
		hp := mt.hp * (1 + depth/2)
		// ⚡ УСКОРЕННЫЙ РОСТ АТАКИ: было depth/3, стало depth/2
		attack := mt.attack * (1 + depth/2)
		// Золото растет плавнее: на 1 этаже x1.5, на 5 этаже x3.5, на 10 этаже x6
		// Это лучше соотносится с линейным ростом цен у торговцев
		gold := mt.gold * (1 + depth/2)
		// Опыт растет линейно, гарантируя стабильный прогресс
		xp := mt.xp + (depth * 2)
		m := NewMonster(x, y, mt.name, hp, attack, gold, xp, mt.symbol, mt.color)
		m.SetLogger(l.logger)
		l.Monsters = append(l.Monsters, m)
	}
}

// =============================================================================
// СПАВН ПРЕДМЕТОВ (Value мечей и щитов масштабируется)
// =============================================================================
func (l *Level) spawnItems(count int) {
	if l == nil || count <= 0 || l.Width < 3 || l.Height < 3 {
		return
	}

	depth := l.Depth
	// НОВОЕ: Value зависит от глубины
	swordValue := 5 + depth*2
	shieldValue := 3 + depth*2

	itemTypes := []struct {
		name   string
		itype  ItemType
		value  int
		symbol rune
		color  tcell.Color
	}{
		{"Зелье здоровья", ItemTypePotion, 35, '!', tcell.ColorRed},
		{"Меч", ItemTypeWeapon, swordValue, '/', tcell.ColorYellow},
		{"Щит", ItemTypeArmor, shieldValue, '[', tcell.ColorBlue},
		{"Мешок золота", ItemTypeGold, 20, '$', tcell.ColorYellow},
		{"Еда", ItemTypePotion, 0, '%', tcell.ColorPurple},
	}
	for i := 0; i < count; i++ {
		var x, y int
		attempts := 0
		for {
			x = 1 + rand.IntN(l.Width-2)
			y = 1 + rand.IntN(l.Height-2)
			if l.Tiles[y][x].Type == TileFloor && !l.hasItemAt(x, y) && !l.hasMonsterAt(x, y) && !l.isStairsAt(x, y) {
				break
			}
			attempts++
			if attempts > MaxSpawnAttempts {
				return
			}
		}
		it := itemTypes[rand.IntN(len(itemTypes))]
		l.Items = append(l.Items, NewItem(x, y, it.name, it.itype, it.value, it.symbol, it.color))
	}
}

// =============================================================================
// СПАВН ТОРГОВЦЕВ
// =============================================================================
func (l *Level) spawnMerchants(depth int) {
	if l == nil {
		return
	}
	if depth != 2 && depth != 5 && depth != 8 && depth != 11 && depth != 14 {
		return
	}
	var x, y int
	attempts := 0
	for {
		x = 1 + rand.IntN(l.Width-2)
		y = 1 + rand.IntN(l.Height-2)
		if l.Tiles[y][x].Type == TileFloor && !l.hasMonsterAt(x, y) && !l.hasItemAt(x, y) && !l.isStairsAt(x, y) {
			break
		}
		attempts++
		if attempts > MaxSpawnAttempts {
			return
		}
	}
	merchant := NewMerchant(x, y, depth)
	l.Merchants = append(l.Merchants, merchant)
}

// =============================================================================
// СПАВН АЛТАРЕЙ
// =============================================================================
func (l *Level) spawnAltars() {
	if l == nil {
		return
	}
	var x, y int
	attempts := 0
	for {
		x = 1 + rand.IntN(l.Width-2)
		y = 1 + rand.IntN(l.Height-2)
		if l.Tiles[y][x].Type == TileFloor && !l.hasMonsterAt(x, y) && !l.hasItemAt(x, y) && !l.isStairsAt(x, y) {
			break
		}
		attempts++
		if attempts > MaxSpawnAttempts {
			return
		}
	}
	altar := NewAltar(x, y)
	l.Altars = append(l.Altars, altar)
}

// =============================================================================
// СПАВН СУНДУКОВ
// =============================================================================
func (l *Level) spawnChests() {
	if l == nil {
		return
	}
	var x, y int
	attempts := 0
	for {
		x = 1 + rand.IntN(l.Width-2)
		y = 1 + rand.IntN(l.Height-2)
		if l.Tiles[y][x].Type == TileFloor && !l.hasMonsterAt(x, y) && !l.hasItemAt(x, y) && !l.isStairsAt(x, y) && !l.hasChestAt(x, y) {
			break
		}
		attempts++
		if attempts > MaxSpawnAttempts {
			return
		}
	}
	isGolden := rand.IntN(100) < 20
	var chest *Chest
	if isGolden {
		chest = NewGoldenChest(x, y)
	} else {
		chest = NewChest(x, y)
	}
	l.Chests = append(l.Chests, chest)
}

// =============================================================================
// ВОЗРОЖДЕНИЕ МОНСТРОВ (ПРИ ПОВТОРНОМ ПОСЕЩЕНИИ)
// =============================================================================
func (l *Level) respawnMonsters() {
	if l == nil {
		return
	}
	monsters := make([]*Monster, 0, len(l.Monsters))
	for _, monster := range l.Monsters {
		if monster != nil && monster.IsBoss && monster.HP > 0 {
			monsters = append(monsters, monster)
		}
	}
	l.Monsters = monsters
	count := (5 + l.Depth) / 2
	if count < 2 {
		count = 2
	}
	monsterTypes := []struct {
		name                 string
		hp, attack, gold, xp int
		symbol               rune
		color                tcell.Color
	}{
		{"Гоблин", 8, 2, 5, 8, 'g', tcell.ColorGreen},
		{"Орк", 12, 3, 10, 15, 'o', tcell.ColorDarkRed},
		{"Скелет", 10, 2, 8, 12, 's', tcell.ColorWhite},
		{"Крыса", 4, 1, 2, 3, 'r', tcell.ColorBrown},
	}
	strengthMultiplier := 1 + (l.VisitCount / 2)
	for i := 0; i < count; i++ {
		var x, y int
		attempts := 0
		for {
			x = 1 + rand.IntN(l.Width-2)
			y = 1 + rand.IntN(l.Height-2)
			if l.Tiles[y][x].Type == TileFloor && !l.hasMonsterAt(x, y) && !l.hasItemAt(x, y) && !l.isStairsAt(x, y) {
				break
			}
			attempts++
			if attempts > MaxSpawnAttempts {
				return
			}
		}
		mt := monsterTypes[rand.IntN(len(monsterTypes))]

		// 🆕 СБАЛАНСИРОВАННОЕ МАСШТАБИРОВАНИЕ С УЧЕТОМ VisitCount
		hp := mt.hp * (1 + l.Depth/2) * strengthMultiplier
		// ⚡ УСКОРЕННЫЙ РОСТ АТАКИ: было depth/3, стало depth/2
		attack := mt.attack * (1 + l.Depth/2) * strengthMultiplier
		// Золото также масштабируется от множителя посещений
		gold := mt.gold * (1 + l.Depth/2) * strengthMultiplier
		xp := mt.xp + (l.Depth * 2)
		m := NewMonster(x, y, mt.name, hp, attack, gold, xp, mt.symbol, mt.color)
		m.SetLogger(l.logger)
		l.Monsters = append(l.Monsters, m)
	}
}

// =============================================================================
// ГЕНЕРАЦИЯ ТОВАРОВ ТОРГОВЦА (Value мечей и щитов масштабируется)
// =============================================================================
func newMerchantItems(depth, visitCount int) []*Item {
	swordValue := 5 + depth*2
	shieldValue := 3 + depth*2

	basePrices := []struct {
		name         string
		itype        ItemType
		value, price int
		symbol       rune
		color        tcell.Color
	}{
		{"Зелье здоровья", ItemTypePotion, 35, 30, '!', tcell.ColorRed},
		{"Еда", ItemTypePotion, 0, 15, '%', tcell.ColorPurple},
		{"Меч", ItemTypeWeapon, swordValue, 50, '/', tcell.ColorYellow},
		{"Щит", ItemTypeArmor, shieldValue, 40, '[', tcell.ColorBlue},
	}

	items := make([]*Item, 0)
	priceMultiplier := visitCount
	for _, bp := range basePrices {
		if rand.IntN(100) < 70 {
			price := (bp.price + depth*5) * priceMultiplier
			item := NewItem(0, 0, bp.name, bp.itype, bp.value, bp.symbol, bp.color)
			item.Price = price
			items = append(items, item)
		}
	}
	return items
}