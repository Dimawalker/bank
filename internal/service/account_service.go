package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dimawalker/bank/internal/models"
	"github.com/Dimawalker/bank/internal/repository"
	"github.com/sirupsen/logrus"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
	creditRepo  *repository.CreditRepository
	logger      *logrus.Logger
}

func NewAccountService(db *sql.DB, logger *logrus.Logger) *AccountService {
	return &AccountService{
		accountRepo: repository.NewAccountRepository(db),
		creditRepo:  repository.NewCreditRepository(db),
		logger:      logger,
	}
}

func (s *AccountService) CreateAccount(req *models.CreateAccountRequest) (*models.Account, error) {
	account := &models.Account{
		UserID:    req.UserID,
		Balance:   req.Balance,
		Currency:  req.Currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.accountRepo.Create(account); err != nil {
		s.logger.WithError(err).Error("Failed to create account")
		return nil, errors.New("internal server error")
	}

	return account, nil
}

func (s *AccountService) GetAccountByID(accountID int64) (*models.Account, error) {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account by ID")
		return nil, errors.New("account not found")
	}

	return account, nil
}

func (s *AccountService) GetUserAccounts(userID int64) ([]*models.Account, error) {
	accounts, err := s.accountRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user accounts")
		return nil, errors.New("internal server error")
	}

	return accounts, nil
}

func (s *AccountService) Transfer(req *models.TransferRequest) error {
	tx, err := s.accountRepo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	srcAccount, err := s.accountRepo.GetByID(req.FromAccountID)
	if err != nil {
		return fmt.Errorf("failed to get source account: %w", err)
	}

	dstAccount, err := s.accountRepo.GetByID(req.ToAccountID)
	if err != nil {
		return fmt.Errorf("failed to get destination account: %w", err)
	}

	if srcAccount.Currency != dstAccount.Currency {
		return errors.New("currency mismatch between accounts")
	}

	if srcAccount.Balance < req.Amount {
		return errors.New("insufficient funds")
	}

	srcAccount.Balance -= req.Amount
	dstAccount.Balance += req.Amount

	if err := s.accountRepo.UpdateBalance(srcAccount.ID, srcAccount.Balance); err != nil {
		return fmt.Errorf("failed to update source account balance: %w", err)
	}

	if err := s.accountRepo.UpdateBalance(dstAccount.ID, dstAccount.Balance); err != nil {
		return fmt.Errorf("failed to update destination account balance: %w", err)
	}

	transaction := &models.Transaction{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Type:          "transfer",
		CreatedAt:     time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AccountService) Deposit(accountID int64, amount float64) error {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account")
		return errors.New("account not found")
	}

	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		s.logger.WithError(err).Error("Failed to update account balance")
		return errors.New("internal server error")
	}

	transaction := &models.Transaction{
		ToAccountID: accountID,
		Amount:      amount,
		Type:        "deposit",
		CreatedAt:   time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		s.logger.WithError(err).Error("Failed to create transaction record")
		return errors.New("internal server error")
	}

	return nil
}

func (s *AccountService) Withdraw(accountID int64, amount float64) error {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account")
		return errors.New("account not found")
	}

	if account.Balance < amount {
		return errors.New("insufficient funds")
	}

	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		s.logger.WithError(err).Error("Failed to update account balance")
		return errors.New("internal server error")
	}

	transaction := &models.Transaction{
		FromAccountID: accountID,
		Amount:        amount,
		Type:          "withdrawal",
		CreatedAt:     time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		s.logger.WithError(err).Error("Failed to create transaction record")
		return errors.New("internal server error")
	}

	return nil
}

type TransactionAnalytics struct {
	TotalTransactions int            `json:"total_transactions"`
	TotalAmount       float64        `json:"total_amount"`
	AverageAmount     float64        `json:"average_amount"`
	MaxAmount         float64        `json:"max_amount"`
	MinAmount         float64        `json:"min_amount"`
	TransactionsByDay map[string]int `json:"transactions_by_day"`
}

func (s *AccountService) GetTransactionAnalytics(userID int64, startDate, endDate time.Time) (*TransactionAnalytics, error) {
	accounts, err := s.accountRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user accounts")
		return nil, err
	}

	var totalTransactions int
	var totalAmount float64
	var maxAmount float64
	var minAmount float64
	transactionsByDay := make(map[string]int)

	for _, account := range accounts {
		transactions, err := s.accountRepo.GetTransactions(account.ID, startDate, endDate)
		if err != nil {
			s.logger.WithError(err).Error("Failed to get account transactions")
			return nil, err
		}

		totalTransactions += len(transactions)
		for _, tx := range transactions {
			totalAmount += tx.Amount
			if tx.Amount > maxAmount {
				maxAmount = tx.Amount
			}
			if tx.Amount < minAmount || minAmount == 0 {
				minAmount = tx.Amount
			}

			day := tx.CreatedAt.Format("2006-01-02")
			transactionsByDay[day]++
		}
	}

	var averageAmount float64
	if totalTransactions > 0 {
		averageAmount = totalAmount / float64(totalTransactions)
	}

	return &TransactionAnalytics{
		TotalTransactions: totalTransactions,
		TotalAmount:       totalAmount,
		AverageAmount:     averageAmount,
		MaxAmount:         maxAmount,
		MinAmount:         minAmount,
		TransactionsByDay: transactionsByDay,
	}, nil
}
