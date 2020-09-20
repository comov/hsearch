package settings

import (
	"fmt"

	"github.com/comov/hsearch/structs"
)

// mainSettingsText - this is template for message that show all configs
const mainSettingsText = `
` + mainSearchText + `

` + mainFiltersText

const mainSearchText = `*Основные настройки поиска*
Искать для тебя квартиры: %s
Искать на elcat.diesel.kg: %s
Искать на house.kg: %s
Искать на lalafo.kg: %s`

const mainFiltersText = `*Фильтры поиска*
Только с фото: %s
Цена в KGS: %s
Цена в USD: %s`

// filter price text
const (
	textKGS = `Укажите суммы в сомах, через дефис в пределах которых нужно искать.

(0 - любая цена / -1 не искать в KGS)
(бот ждет ответа около минуты, потом забывает изменить этот фильтр)

Пример:
10000 - 20000`
	textUSD = `Укажите суммы в долларах, через дефис в пределах которых нужно искать.

(бот ждет ответа около минуты, потом забывает изменить этот фильтр)
(0 - любая цена / -1 не искать в USD)

Пример:
250 - 350`
)

func yesNo(v bool) string {
	if v {
		return "Да"
	}
	return "Нет"
}

func price(prices structs.Price) string {
	return fmt.Sprintf("%d - %d", prices[0], prices[1])
}
