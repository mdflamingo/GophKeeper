package service

import (
	"github.com/mdflamingo/GophKeeper/internal/model"
	"github.com/mdflamingo/GophKeeper/internal/server/repository/postgres"
)

type GopheKeeperService struct {
	repo *postgres.DBStorage
}

func NewGopheKeeperService(repo *postgres.DBStorage) *GopheKeeperService {
	return &GopheKeeperService{repo: repo}
}

// GetUserItems возвращает список секретов в виде response моделей
func (s *GopheKeeperService) GetUserItems(userID int) (*model.UserDataListResponse, error) {
	if userID == 0 {
		return &model.UserDataListResponse{
			Items: []model.UserDataResponse{},
			Count: 0,
		}, nil
	}

	items, err := s.repo.GetList(userID)
	if err != nil {
		return nil, err
	}

	result := make([]model.UserDataResponse, len(items))
	for i, item := range items {
		result[i] = model.UserDataResponse{
			DataType:  postgres.DataType(item.DataType),
			MetaData:  item.MetaData,
			CreatedAt: item.CreatedAt,
		}
	}

	response := &model.UserDataListResponse{
		Items: result,
		Count: len(items),
	}

	return response, nil
}
