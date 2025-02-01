package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"wallet/internal/lib/api/response"
	"wallet/storage"
	"wallet/storage/postgresql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOperation — мок реализация интерфейса Operation
type MockOperation struct {
	mock.Mock
}

func (m *MockOperation) DepositWallet(walletID uuid.UUID, amount int64) (postgresql.Wallet, error) {
	args := m.Called(walletID, amount)
	return args.Get(0).(postgresql.Wallet), args.Error(1)
}

func (m *MockOperation) WithdrawWallet(walletID uuid.UUID, amount int64) (postgresql.Wallet, error) {
	args := m.Called(walletID, amount)
	return args.Get(0).(postgresql.Wallet), args.Error(1)
}

// TestWalletOperationConcurrent — тест с 1000 запросами
func TestWalletOperationConcurrent(t *testing.T) {
	mockOp := new(MockOperation)

	// Создаем тестовый кошелек
	testWalletID := uuid.New()
	mockOp.On("DepositWallet", testWalletID, int64(100)).Return(postgresql.Wallet{WalletID: testWalletID, Balance: 100}, nil)

	logger := slog.Default()
	handler := WalletOperation(logger, mockOp)

	r := chi.NewRouter()
	r.Post("/api/v1/wallet/operation", handler)

	server := httptest.NewServer(r)
	defer server.Close()

	var wg sync.WaitGroup
	numRequests := 1000
	successCount := 0
	mu := sync.Mutex{}

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			reqBody, _ := json.Marshal(map[string]interface{}{
				"valletId":      testWalletID,
				"operationType": "DEPOSIT",
				"amount":        100,
			})

			req := httptest.NewRequest("POST", "/api/v1/wallet/operation", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			fmt.Printf("Sending request %d\n", i+1)

			r.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			} else {
				t.Errorf("Request %d failed with status: %d", i+1, rec.Code)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, numRequests, successCount, "Not all requests succeeded")
}

// TestWalletOperationCases — тестирование разных сценариев
func TestWalletOperationCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockWallet     postgresql.Wallet
		mockErr        error
		expectedStatus int
		expectedResp   response.Response
	}{
		{
			name: "successful deposit",
			requestBody: map[string]interface{}{
				"valletId":      "f22bd5ed-9155-4ba0-90c4-4880912d7ad4",
				"operationType": "DEPOSIT",
				"amount":        100,
			},
			mockWallet: postgresql.Wallet{
				WalletID: uuid.MustParse("f22bd5ed-9155-4ba0-90c4-4880912d7ad4"),
				Balance:  100,
			},
			expectedStatus: http.StatusOK,
			expectedResp:   response.Response{Status: response.StatusOK},
		},
		{
			name: "successful withdraw",
			requestBody: map[string]interface{}{
				"valletId":      "a45c73fd-3e36-466a-8e57-15e1cf0f35d2",
				"operationType": "WITHDRAW",
				"amount":        50,
			},
			mockWallet: postgresql.Wallet{
				WalletID: uuid.MustParse("a45c73fd-3e36-466a-8e57-15e1cf0f35d2"),
				Balance:  50,
			},
			expectedStatus: http.StatusOK,
			expectedResp:   response.Response{Status: response.StatusOK},
		},
		{
			name: "wallet not found",
			requestBody: map[string]interface{}{
				"valletId":      "b1234567-89ab-cdef-0123-456789abcdef",
				"operationType": "WITHDRAW",
				"amount":        100,
			},
			mockErr:        storage.ErrWalletNotFound,
			expectedStatus: http.StatusNotFound,
			expectedResp:   response.Response{Status: response.StatusError, Error: "walletId not found"},
		},
		{
			name: "insufficient funds",
			requestBody: map[string]interface{}{
				"valletId":      "c9876543-21ab-cdef-4567-89abcdef1234",
				"operationType": "WITHDRAW",
				"amount":        500,
			},
			mockErr:        storage.ErrInsufficientFunds,
			expectedStatus: http.StatusNotFound,
			expectedResp:   response.Response{Status: response.StatusError, Error: "insufficient funds"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOp := new(MockOperation)
			testWalletUUID := uuid.MustParse(tt.requestBody["valletId"].(string))
			amount := int64(tt.requestBody["amount"].(int))
	
			if tt.mockErr == nil {
				if tt.requestBody["operationType"] == "DEPOSIT" {
					mockOp.On("DepositWallet", testWalletUUID, amount).Return(tt.mockWallet, nil)
				} else {
					mockOp.On("WithdrawWallet", testWalletUUID, amount).Return(tt.mockWallet, nil)
				}
			} else {
				mockOp.On("WithdrawWallet", testWalletUUID, amount).Return(postgresql.Wallet{}, tt.mockErr)
			}
	
			logger := slog.Default()
			handler := WalletOperation(logger, mockOp)
	
			r := chi.NewRouter()
			r.Post("/api/v1/wallet/operation", handler)
	
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/wallet/operation", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
	
			fmt.Printf("Sending request: %s\n", req.URL)
	
			r.ServeHTTP(rec, req)
	
			assert.Equal(t, tt.expectedStatus, rec.Code)
	
			var res response.Response
			if err := render.DecodeJSON(rec.Body, &res); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
	
			assert.Equal(t, tt.expectedResp, res)
	
			mockOp.AssertExpectations(t)
		})
	}
}
