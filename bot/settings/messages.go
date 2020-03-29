package settings

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
Цена в KGS: %s (0 - любая цена / -1 не искать в KGS)
Цена в USD: %s (0 - любая цена / -1 не искать в USD)`


// filter price text
const (
	textKGS = `Укажите суммы в сомах, через дефис в пределах которых нужно искать. К примеру:
10000 - 20000`
	textUSD = `Укажите суммы в долларах, через дефис в пределах которых нужно искать. К примеру:
250 - 350`
)


func yesNo(search bool) string {
	if search {
		return"Да"
	}
	return "Нет"
}
