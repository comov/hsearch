package settings

import "fmt"

// mainSettingsText - this is template for message that show all configs
const mainSettingsText = `*Все настройки твоего бота*

` + mainSearchText + `

`+ mainFiltersText

const mainSearchText = `*Настройки поиска*
Искать для тебя квартиры: %s`
// Искать на elcat.diesel.kg: %t
// Искать на lalafo.kg: %t

const mainFiltersText = `*Фильтры поиска*
Только с фото: %s
Цена в KGS: %s
Цена в USD: %s`


// filter price text
const (
	textKGS = `Укажите суммы в сомах, через дефис в пределах которых нужно искать.
(0 - любая цена / -1 не искать в KGS)
(бот ждет ответа около минуты, потом забывает изменить этот фильтр)
К примеру:
10000 - 20000`
	textUSD = `Укажите суммы в долларах, через дефис в пределах которых нужно искать. К примеру:
(бот ждет ответа около минуты, потом забывает изменить этот фильтр)
(0 - любая цена / -1 не искать в USD)
250 - 350`
)


func yesNo(search bool) string {
	if search {
		return"Да"
	}
	return "Нет"
}

func price(prices [2]int) string {
	return fmt.Sprintf("%d - %d", prices[0], prices[1])
}
