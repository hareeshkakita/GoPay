package httptransport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
	"github.com/hareeshkakita/gopay/internal/service"
)

type Handler struct {
	service *service.WalletService
}

func NewHandler(svc *service.WalletService) *Handler {
	return &Handler{service: svc}
}

type createWalletRequest struct {
	OwnerID  string `json:"owner_id"`
	Currency string `json:"currency"`
}

type createWalletResponse struct {
	ID       string `json:"id"`
	OwnerID  string `json:"owner_id"`
	Currency string `json:"currency"`
}

type walletResponse struct {
	ID       string `json:"id"`
	OwnerID  string `json:"owner_id"`
	Currency string `json:"currency"`
}

type balanceResponse struct {
	WalletID         string `json:"wallet_id"`
	AvailableBalance int64  `json:"available_balance"`
	PendingBalance   int64  `json:"pending_balance"`
	Currency         string `json:"currency"`
}

type depositMoneyRequest struct {
	Amount int64 `json:"amount"`
}

type depositMoneyResponse struct {
	WalletID         string `json:"wallet_id"`
	AvailableBalance int64  `json:"available_balance"`
}

type withdrawMoneyRequest struct {
	Amount int64 `json:"amount"`
}

type withdrawMoneyResponse struct {
	WalletID         string `json:"wallet_id"`
	AvailableBalance int64  `json:"available_balance"`
}

type transferMoneyRequest struct {
	Source         string `json:"source_wallet_id"`
	Target         string `json:"target_wallet_id"`
	Amount int64 `json:"amount"`
}

type transferMoneyResponse struct {
	WalletBalance         []balanceResponse `json:"wallet_balance"`
}

func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req createWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		http.Error(w, "invalid owner_id", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		req.Currency = "INR"
	}

	wallet, err := h.service.CreateWallet(r.Context(), ownerID, req.Currency)
	if err != nil {
		http.Error(w, "failed to create wallet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createWalletResponse{
		ID:       wallet.ID.String(),
		OwnerID:  wallet.OwnerID.String(),
		Currency: wallet.Currency,
	})
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "walletID"))
	if err != nil {
		http.Error(w, "invalid wallet id", http.StatusBadRequest)
		return
	}

	wallet, err := h.service.GetWallet(r.Context(), walletID)
	if err != nil {
		http.Error(w, "wallet not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(walletResponse{
		ID:       wallet.ID.String(),
		OwnerID:  wallet.OwnerID.String(),
		Currency: wallet.Currency,
	})
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "walletID"))
	if err != nil {
		http.Error(w, "invalid wallet id", http.StatusBadRequest)
		return
	}

	balance, err := h.service.GetBalance(r.Context(), walletID)
	if err != nil {
		http.Error(w, "balance not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(balanceResponse{
		WalletID:         balance.WalletID.String(),
		AvailableBalance: balance.AvailableBalance,
		PendingBalance:   balance.PendingBalance,
		Currency:         balance.Currency,
	})
}

func (h *Handler) DepositMoney(w http.ResponseWriter, r *http.Request) {
	var req depositMoneyRequest
	walletID, err := uuid.Parse(chi.URLParam(r, "walletID"))
	if err != nil {
		http.Error(w, "invalid wallet id", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	updatedBalance, err := h.service.UpdateBalanceByWalletID(r.Context(), dbsqlc.ApplyNewBalanceParams{
		WalletID:         walletID,
		AvailableBalance: req.Amount,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(depositMoneyResponse{
		WalletID:         walletID.String(),
		AvailableBalance: updatedBalance.AvailableBalance,
	})
}

func (h *Handler) WithdrawMoney(w http.ResponseWriter, r *http.Request) {
	var req withdrawMoneyRequest
	walletID, err := uuid.Parse(chi.URLParam(r, "walletID"))
	if err != nil {
		http.Error(w, "invalid wallet id", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	updatedBalance, err := h.service.UpdateBalanceByWalletID(r.Context(), dbsqlc.ApplyNewBalanceParams{
		WalletID:         walletID,
		AvailableBalance: -1 * req.Amount,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(withdrawMoneyResponse{
		WalletID:         walletID.String(),
		AvailableBalance: updatedBalance.AvailableBalance,
	})
}

func (h *Handler) TransferMoney(w http.ResponseWriter, r *http.Request) {
	var req transferMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	source,err := uuid.Parse(req.Source)
	target,err := uuid.Parse(req.Target)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceBalance , targetBalance, err := h.service.TransferAmount(r.Context(),source,target,req.Amount)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var updatedBalances = []balanceResponse{{WalletID:         sourceBalance.WalletID.String(),
		AvailableBalance: sourceBalance.AvailableBalance,
		PendingBalance:   sourceBalance.PendingBalance,
		Currency:         sourceBalance.Currency,
		},{WalletID:         targetBalance.WalletID.String(),
		AvailableBalance: targetBalance.AvailableBalance,
		PendingBalance:   targetBalance.PendingBalance,
		Currency:         targetBalance.Currency,
		}}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(transferMoneyResponse{
		WalletBalance : updatedBalances,
	})
}