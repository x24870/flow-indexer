package service

import (
	"context"
	"flow-indexer/internal/domain/account"
	"flow-indexer/internal/domain/inscription"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type Service interface {
	UpdateBalance(ctx context.Context, insName, address string, isDeposit bool) error
}

type service struct {
	accountRepo     account.Repository
	inscriptionRepo inscription.Repository
}

func NewService(
	accountRepo account.Repository,
	inscriptionRepo inscription.Repository,
) Service {
	return &service{
		accountRepo:     accountRepo,
		inscriptionRepo: inscriptionRepo,
	}
}

func (s *service) UpdateBalance(ctx context.Context, insName, address string, isDeposit bool) error {
	acc, err := s.accountRepo.FirstOrCreate(ctx, address)
	if err != nil {
		return err
	}

	balance, err := s.inscriptionRepo.GetorCreateByInscriptionAndAddress(ctx, insName, acc.Address)
	if err != nil {
		return err
	}

	if isDeposit {
		balance.Amount += 1
	} else {
		balance.Amount -= 1
	}

	err = s.inscriptionRepo.Update(ctx, balance)
	if err != nil {
		logger.Error("UpdateBalance", zap.Error(err))
		return err
	}
	return nil
}
