# hsearch
По сути парсер ресурсов для создания объявлений о сдаче квартир в аренду. Ресурсы для парсинга:

- [diesel.elcat.kg](http://diesel.elcat.kg/)
- [house.kg](http://house.kg/)
- [lalafo.kg](https://lalafo.kg/)

- [Telegram ссылка на бота](https://t.me/house_search_assistant_bot)
- [Docker образ бота](https://hub.docker.com/r/comov/hsearch)

## Какую проблему решает бот?
Ни один из ресурсов, не предоставляет инструментов для отсеивания уже просмотренных тем, так
 же у некоторых нет фильтра по цене, количеству комнат, типу (квартира/офис) и
 это стало не удобным для меня. Я часто меняю квартиры и найти хорошую квартиру
 раньше других, это хорошо для меня. Можно создать группу с вашей подругой/другом, включить бота в группе, получать
 и сразу обсуждать предложенные варианты.

## Как он работает для пользователя
Бот раз в N времени заходит на все ресурсы, получает новые объявления, проверяет, подходит ли эта
 квартира тебе и отправляет ее. Все просто.

## Чего-то в боте не хватает?
И ты скорее всего прав! Можешь зайти [сюда](https://github.com/comov/hsearch/issues), нажать "New Issue"
 и создать задачу в которой мы обсудим то, что тебе не хватает.

## Developer documentation
Content manager ходит за объявлениями раз в N минут и тащит все объявления на
 первой странице, затем пишет в локальную бд на N недель. Дальше пользователь
 открывает бота, говорит что хочет получать объявления и начинает получать
 максимально свежие объявления.
 
The project supports go modules

```shell script
git clone https://github.com/comov/hsearch.git
cd hsearch
make mod
make migrate
make run
```

Build/Run with Docker 
```shell script
make dockerbuild
make dockerrun
```

For more information, take a look at Makefile

## we use sentry

[Sentry](https://sentry.io) is a cool bug tracker! But in GoLang I don't know how it is used. So I decided,
 that errors will be caught in the top-level of the trimmings. Hsearch has the following top-level components :
 
 - grabber - the need to fill the database with new data
 - parser - we got new data from HTML, so we need the html parser 
 - matcher - agent to find new data for each user and send him a message    
 - garbage - all data can be older, so we need to clean up him
 - bot - telegram interface for communication with hsearch
 - api (beta) - HTTP Api for the WEB and Mobile


## rsyslog setting
```shell script
root@docker-host:~# cat /etc/rsyslog.d/30-docker.conf
$FileCreateMode 0644
template(name="DockerLogFileName" type="list") {
 constant(value="/var/log/docker/")
 property(name="syslogtag" securepath="replace" \
 regex.expression="docker/\\(.*\\)\\[" regex.submatch="1")
 constant(value=".log")
}

if $programname == 'docker' then \
 if $syslogtag contains 'docker/' then \
 ?DockerLogFileName
 & stop
$FileCreateMode 0640
```

## Новые возможности:
 - [ ] Привести в порядок данные "By room numbers" ("1 м2"); "By floor" ("из 1"); "By area" ("м2", "float", "< 10", "> 399")
 - [ ] "Агенство" более 2-х объявлений (beta) + кнопка "сообшить об ошибки"
 - [ ] Фильтр по этажам
 - [ ] Фильтр по количевсту комнат
 - [ ] Не удаляются старые сообщения при клике "Точно нет"
 - [ ] Нет нотификации в desktop приложении "Больше не покажу"
 - [ ] Follow - следить за изменениями этого предложения Up/Change (кнопка в предложении)
 - [ ] Добавить настройки, которые позволят скрывать ненужные поля объявления

## Тех. долг:
 - [ ] Поправить документацию и описание. Указать страну. Выпилить шутейки.
 - [ ] Добавить sentry
 - [ ] Приемочные тесты
 - [ ] Пройтись по всем TODO в коде
 - [ ] Добавить в build тег и дату последней сборки
 - [ ] Переделать кнопки ответов и вшить в них offerId (description:123)
 - [ ] Порефакторить background manager
 - [ ] Порефакторить весь код. Распараллелить некоторые моменты
 - [ ] Мониторинг приложения
 - [ ] Админка и статистика
