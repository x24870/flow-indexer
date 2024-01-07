package service

import (
	"context"
	"flow-indexer/internal/domain/account"
	flowEvent "flow-indexer/internal/domain/event"
	"flow-indexer/internal/domain/inscription"
)

type Service interface {
	UpdateBalance(ctx context.Context, insName, address string, isDeposit bool) error
	CreateFlowEvent(ctx context.Context, nftID uint64, account, event string, block uint64) error
}

type service struct {
	accountRepo     account.Repository
	inscriptionRepo inscription.Repository
	eventRepo       flowEvent.Repository
}

func NewService(
	accountRepo account.Repository,
	inscriptionRepo inscription.Repository,
	eventRepo flowEvent.Repository,
) Service {
	return &service{
		accountRepo:     accountRepo,
		inscriptionRepo: inscriptionRepo,
		eventRepo:       eventRepo,
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
		return err
	}
	return nil
}

func (s *service) CreateFlowEvent(ctx context.Context, nftID uint64, account, event string, block uint64) error {
	_, err := s.accountRepo.FirstOrCreate(ctx, account)
	if err != nil {
		return err
	}

	fe := flowEvent.FlowEvent{
		Account: account,
		NFTID:   nftID,
		Event:   event,
		Block:   block,
	}

	return s.eventRepo.Create(ctx, &fe)
}
