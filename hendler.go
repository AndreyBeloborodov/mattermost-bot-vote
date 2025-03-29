package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"strconv"
	"strings"
)

func handleVoteCreate(app *application, post *model.Post) {
	// Убираем префикс "/vote create" и разбиваем сообщение на части
	message := strings.TrimPrefix(post.Message, "/vote create ")
	parts := strings.Split(message, ", ")

	// Должен быть хотя бы один вопрос и два варианта ответа
	if len(parts) < 3 {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: формат команды - /vote create вопрос, вариант1, вариант2, ...", post.Id)
		return
	}

	// Отправляем голосование на обработку в бизнес-логику
	processVoteCreate(app, parts[0], parts[1:], post)
}

func handleVote(app *application, post *model.Post) {
	// Разбираем команду
	parts := strings.Split(post.Message, " ")
	if len(parts) != 4 {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: неверный формат команды. Используйте /vote send <id> <номер варианта>", post.Id)
		return
	}

	voteID := parts[2]
	optionIndex, err := strconv.Atoi(parts[3])
	if err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: неверный формат команды. Используйте /vote <id> <номер варианта>", post.Id)
		return
	}

	processVote(app, voteID, optionIndex, post)
}

func handleVoteResult(app *application, post *model.Post) {
	// Разбираем команду
	parts := strings.Split(post.Message, " ")
	if len(parts) != 3 {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: неверный формат команды. Используйте /vote result <id>", post.Id)
		return
	}

	voteID := parts[2]

	processVoteResult(app, voteID, post)
}

func handleVoteClose(app *application, post *model.Post) {
	// Разбираем команду
	parts := strings.Split(post.Message, " ")
	if len(parts) != 3 {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: неверный формат команды. Используйте /vote close <id>", post.Id)
		return
	}

	voteID := parts[2]

	processVoteClose(app, voteID, post)
}

func handleVoteDelete(app *application, post *model.Post) {
	// Разбираем команду
	parts := strings.Split(post.Message, " ")
	if len(parts) != 3 {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: неверный формат команды. Используйте /vote delete <id>", post.Id)
		return
	}

	voteID := parts[2]

	processVoteDelete(app, voteID, post)
}

func handleVoteHelp(app *application, post *model.Post) {
	helpMessage := `
Доступные команды для голосования:

1. /vote create <вопрос>, <вариант1>, <вариант2>, ... - создаёт новое голосование с вопросом и вариантами.
2. /vote send <id голосования> <номер варианта> - голосовать за один из вариантов.
3. /vote result <id голосования> - просматривать результаты голосования.
4. /vote close <id голосования> - завершить голосование (доступно только создателю голосования).
5. /vote delete <id голосования> - удалить голосование (доступно только создателю голосования).

Пример использования:
- /vote create "Какой ваш любимый цвет?", "Красный", "Синий", "Зелёный"
- /vote send 12345 2
- /vote result 12345
`

	// Отправляем сообщение с помощью функции sendMsgToChannel
	sendMsgToChannel(app, post.ChannelId, helpMessage, post.Id)
}
