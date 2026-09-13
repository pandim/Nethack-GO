package game

import (
	"github.com/gdamore/tcell/v2"
)

// =============================================================================
// СТРУКТУРА ТОРГОВЦА
// =============================================================================
type Merchant struct {
	X, Y  int
	Items []*Item
}

// =============================================================================
// СОЗДАНИЕ ТОРГОВЦА (Value мечей и щитов масштабируется)
// =============================================================================
func NewMerchant(x, y, depth int) *Merchant {
	swordValue := 5 + depth*2
	shieldValue := 3 + depth*2

	basePrices := []struct {
		name   string
		itype  ItemType
		value  int
		price  int
		symbol rune
		color  tcell.Color
	}{
		{"Зелье здоровья", ItemTypePotion, 10, 30, '!', tcell.ColorRed},
		{"Еда", ItemTypePotion, 0, 15, '%', tcell.ColorPurple},
		{"Меч", ItemTypeWeapon, swordValue, 50, '/', tcell.ColorYellow},
		{"Щит", ItemTypeArmor, shieldValue, 40, '[', tcell.ColorBlue},
	}

	items := make([]*Item, 0)
	for _, bp := range basePrices {
		// Базовая цена тоже немного растет с глубиной
		price := bp.price + depth*5 
		item := NewItem(0, 0, bp.name, bp.itype, bp.value, bp.symbol, bp.color)
		item.Price = price
		items = append(items, item)
	}

	return &Merchant{
		X:     x,
		Y:     y,
		Items: items,
	}
}

// =============================================================================
// ОТРИСОВКА ТОРГОВЦА
// =============================================================================
func (m *Merchant) Render(screen tcell.Screen, offsetX, offsetY int) {
	if m == nil || screen == nil {
		return
	}
	style := tcell.StyleDefault.
		Foreground(tcell.ColorAqua).
		Background(tcell.ColorBlack)
	screen.SetContent(m.X+offsetX, m.Y+offsetY, 'M', nil, style)
}