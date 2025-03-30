package main

import (
	"fmt"
	"github.com/tarantool/go-tarantool"
	_ "log"
	"math/rand"
	"strconv"
	"time"
)

// VoteRepository Репозиторий для работы с Tarantool
type VoteRepository struct {
	client *tarantool.Connection
}

// NewVoteRepository Создаём новое соединение с базой данных
func NewVoteRepository(host, port string, opts tarantool.Opts) (*VoteRepository, error) {
	client, err := tarantool.Connect(host+port, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Tarantool: %v", err)
	}

	return &VoteRepository{client: client}, nil
}

// SaveVote Метод для сохранения голосования
func (r *VoteRepository) SaveVote(vote *Vote) (string, error) {
	voteID := generateVoteID() // Генерируем уникальный ID голосования
	vote.ID = strconv.Itoa(voteID)
	vote.Votes = make(map[string]int)

	_, err := r.client.Insert("votes", []interface{}{vote.ID, vote.Question, vote.Options, vote.Votes, vote.CreatorID, vote.Closed})
	if err != nil {
		return "", fmt.Errorf("failed to insert vote: %v", err)
	}
	return vote.ID, nil
}

func generateVoteID() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(9000000) + 1000000 // Генерируем 7-значный ID
}

// GetVoteByID Метод для получения голосования по ID
func (r *VoteRepository) GetVoteByID(id string) (*Vote, error) {
	// Выполняем запрос
	resp, err := r.client.Select("votes", "primary", 0, 1, tarantool.IterEq, []interface{}{id})
	if err != nil {
		return nil, fmt.Errorf("Ошибка запроса: %v", err)
	}

	// Проверяем, есть ли данные
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("vote not found")
	}

	// Преобразуем результат в структуру Vote
	data := resp.Data[0].([]interface{}) // Берем первую строку из ответа
	vote := &Vote{
		ID:        data[0].(string),
		Question:  data[1].(string),
		Options:   convertToStringSlice(data[2]),
		Votes:     convertToStringIntMap(data[3]),
		CreatorID: data[4].(string),
		Closed:    data[5].(bool),
	}
	return vote, nil
}

// SaveVoteResult Метод для сохранения голоса пользователя
func (r *VoteRepository) SaveVoteResult(voteID string, userID string, optionIndex int) error {
	// Получаем текущее голосование
	vote, err := r.GetVoteByID(voteID)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("vote not found")
	}

	// Обновляем словарь голосов
	vote.Votes[userID] = optionIndex

	// Обновляем запись в базе Tarantool
	_, err = r.client.Update("votes", "primary", []interface{}{voteID}, []tarantool.Op{
		{Op: "=", Field: 3, Arg: vote.Votes},
	})
	if err != nil {
		return fmt.Errorf("failed to update vote: %v", err)
	}

	return nil
}

// SaveVoteStatus Метод для изменения статуса голосования
func (r *VoteRepository) SaveVoteStatus(voteID string, status bool) error {
	// Получаем текущее голосование
	vote, err := r.GetVoteByID(voteID)
	if err != nil {
		return err
	}
	if vote == nil {
		return fmt.Errorf("vote not found")
	}

	// Обновляем статус голосования
	vote.Closed = status

	// Обновляем запись в базе Tarantool
	resp, err := r.client.Update("votes", "primary", []interface{}{voteID}, []tarantool.Op{
		{Op: "=", Field: 5, Arg: vote.Closed},
	})
	fmt.Println(resp)
	if err != nil {
		return fmt.Errorf("failed to update vote: %v", err)
	}

	return nil
}

func (r *VoteRepository) DeleteVote(voteID string) error {
	_, err := r.client.Delete("votes", "primary", []interface{}{voteID})
	if err != nil {
		return fmt.Errorf("failed to delete vote: %v", err)
	}

	return nil
}

// Вспомогательная функция для преобразования интерфейса в []string
func convertToStringSlice(input interface{}) []string {
	if input == nil {
		return nil
	}
	if arr, ok := input.([]interface{}); ok {
		strArr := make([]string, len(arr))
		for i, v := range arr {
			strArr[i] = fmt.Sprintf("%v", v)
		}
		return strArr
	}
	return nil
}

// Вспомогательная функция для преобразования интерфейса в map[string]int
func convertToStringIntMap(input interface{}) map[string]int {
	if input == nil {
		return nil
	}

	if mp, ok := input.(map[interface{}]interface{}); ok {
		strIntMap := make(map[string]int, len(mp))
		for k, v := range mp {
			keyStr := fmt.Sprintf("%v", k) // Преобразуем ключ в строку
			if valInt, ok := v.(uint64); ok {
				strIntMap[keyStr] = int(valInt) // Если значение int — сохраняем
			} else {
				fmt.Printf("Unexpected value type: %T\n", v) // Логируем неожиданные типы
			}
		}
		return strIntMap
	}

	return nil
}
