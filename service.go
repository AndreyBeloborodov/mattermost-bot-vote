package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
)

func processVoteCreate(app *application, question string, options []string, post *model.Post) {
	// Формируем сущность голосования
	vote := Vote{
		Question:  question,
		Options:   options,
		CreatorID: post.UserId,
		Closed:    false,
	}

	// Сохраняем голосование в репозитории и получаем ID
	voteID, err := app.voteRepo.SaveVote(&vote)
	if err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: не удалось создать голосование", post.Id)
		app.logger.Error().Err(err)
		return
	}
	vote.ID = voteID

	// Формируем сообщение
	message := fmt.Sprintf("📊 *Голосование #%s*\n*%s*\n\n", vote.ID, vote.Question)
	for i, option := range vote.Options {
		message += fmt.Sprintf("🔹 %d. %s\n", i+1, option)
	}

	// Отправляем сообщение в канал
	sendMsgToChannel(app, post.ChannelId, message, "")
}

func processVote(app *application, voteID string, optionIndex int, post *model.Post) {
	// Получаем голосование из репозитория
	vote, err := app.voteRepo.GetVoteByID(voteID)
	if err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: голосование не найдено", post.Id)
		return
	}

	// Проверяем, что голосование открыто
	if vote.Closed {
		sendMsgToChannel(app, post.ChannelId, "Голосование закрыто", post.Id)
		return
	}

	// Проверяем, что пользователь уже голосовал
	if _, voted := vote.Votes[post.UserId]; voted {
		sendMsgToChannel(app, post.ChannelId, "Вы уже голосовали!", post.Id)
		return
	}

	// Проверяем корректность варианта
	if optionIndex < 1 || optionIndex > len(vote.Options) {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: некорректный номер варианта", post.Id)
		return
	}

	// Записываем голос
	if err = app.voteRepo.SaveVoteResult(voteID, post.UserId, optionIndex); err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: не удалось сохранить голос", post.Id)
		app.logger.Error().Err(err)
		return
	}

	// Подтверждаем голос
	confirmationMsg := fmt.Sprintf("✅ Ваш голос за вариант *%s* засчитан!", vote.Options[optionIndex-1])
	sendMsgToChannel(app, post.ChannelId, confirmationMsg, post.Id)
}

func processVoteResult(app *application, voteID string, post *model.Post) {
	// Получаем голосование из репозитория
	vote, err := app.voteRepo.GetVoteByID(voteID)
	if err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: голосование не найдено", post.Id)
		app.logger.Error().Err(err)
		return
	}

	// Формируем результаты голосования
	results := make(map[string]int)
	for _, choice := range vote.Votes {
		results[vote.Options[choice-1]]++
	}

	// Создаём сообщение с результатами
	message := fmt.Sprintf("📊 *Результаты голосования #%s*\n*%s*\n\n", vote.ID, vote.Question)
	for i, option := range vote.Options {
		count := results[option]
		message += fmt.Sprintf("🔹 %d. %s - %d голос(ов)\n", i+1, option, count)
	}

	// Отправляем сообщение с результатами в канал
	sendMsgToChannel(app, post.ChannelId, message, "")
}

func processVoteClose(app *application, voteID string, post *model.Post) {
	vote, err := app.voteRepo.GetVoteByID(voteID)
	if err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: голосование не найдено", post.Id)
		app.logger.Error().Err(err)
		return
	}

	// Проверяем, что голосование закрывает его создатель
	if vote.CreatorID != post.UserId {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: только создатель голосования может его завершить", post.Id)
		return
	}

	// Закрываем голосование
	if err = app.voteRepo.SaveVoteStatus(vote.ID, true); err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: не удалось закрыть голосование", post.Id)
		app.logger.Error().Err(err)
		return
	}

	sendMsgToChannel(app, post.ChannelId, fmt.Sprintf("✅ Голосование #%s завершено!", vote.ID), "")
}

func processVoteDelete(app *application, voteID string, post *model.Post) {
	vote, err := app.voteRepo.GetVoteByID(voteID)
	if err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: голосование не найдено", post.Id)
		app.logger.Error().Err(err)
		return
	}

	// Проверяем, что голосование удаляет его создатель
	if vote.CreatorID != post.UserId {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: только создатель голосования может его удалить", post.Id)
		return
	}

	// Удаляем голосование
	if err = app.voteRepo.DeleteVote(voteID); err != nil {
		sendMsgToChannel(app, post.ChannelId, "Ошибка: не удалось удалить голосование", post.Id)
		app.logger.Error().Err(err)
		return
	}

	sendMsgToChannel(app, post.ChannelId, "🗑 Голосование удалено!", post.Id)
}
