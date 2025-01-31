package getter

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wallet/storage/postgresql"
	"wallet/internal/lib/api/response"
	"fmt"
)

type MockGetterWallet struct {
	mock.Mock
}

func (m *MockGetterWallet) GetWallet(wallet_uuid uuid.UUID) (postgresql.Wallet, error) {
	args := m.Called(wallet_uuid)
	return args.Get(0).(postgresql.Wallet), args.Error(1)
}

func TestFetchWallet(t *testing.T) {
	tests := []struct {
		name             string
		walletUUID       string
		mockWalletUUID   uuid.UUID
		mockWallet       postgresql.Wallet
		mockErr          error
		expectedStatus   int
		expectedResponse response.Response
	}{
		{
			name:           "successful fetch wallet",
			walletUUID:     "f22bd5ed-9155-4ba0-90c4-4880912d7ad4",
			mockWalletUUID: uuid.MustParse("f22bd5ed-9155-4ba0-90c4-4880912d7ad4"),
			mockWallet: postgresql.Wallet{
				WalletID: uuid.MustParse("f22bd5ed-9155-4ba0-90c4-4880912d7ad4"),
				Balance:  100,
			},
			expectedStatus: http.StatusOK,
			expectedResponse: response.Response{
				Status: response.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock для GetterWallet
			mockGetterWallet := new(MockGetterWallet)
			if tt.mockErr == nil {
				mockGetterWallet.On("GetWallet", tt.mockWalletUUID).Return(tt.mockWallet, nil)
			} else {
				mockGetterWallet.On("GetWallet", tt.mockWalletUUID).Return(postgresql.Wallet{}, tt.mockErr)
			}

			// Создаем тестовый сервер с маршрутом
			r := chi.NewRouter()
			r.Get("/api/v1/wallets/{WALLET_UUID}", FetchWallet(nil, mockGetterWallet))

			for i := 0; i < 1000; i++ {  // Цикл для 1000 запросов
				req := httptest.NewRequest("GET", "/api/v1/wallets/"+tt.walletUUID, nil)
				rec := httptest.NewRecorder()

				// Логирование запроса
				fmt.Printf("Sending request %d: %s\n", i+1, req.URL)

				r.ServeHTTP(rec, req)

				// Проверяем статус
				assert.Equal(t, tt.expectedStatus, rec.Code)

				// Проверяем ответ
				var res response.Response
				if err := render.DecodeJSON(rec.Body, &res); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				assert.Equal(t, tt.expectedResponse, res)
			}

			// Проверяем, что mock был вызван корректно
			mockGetterWallet.AssertExpectations(t)
		})
	}
}

func TestFetchWalletOne(t *testing.T) {
	tests := []struct {
		name             string
		walletUUID       string
		mockWalletUUID   uuid.UUID
		mockWallet       postgresql.Wallet
		mockErr          error
		expectedStatus   int
		expectedResponse response.Response
	}{
		{
			name:           "successful fetch wallet 1",
			walletUUID:     "f22bd5ed-9155-4ba0-90c4-4880912d7ad4",
			mockWalletUUID: uuid.MustParse("f22bd5ed-9155-4ba0-90c4-4880912d7ad4"),
			mockWallet: postgresql.Wallet{
				WalletID: uuid.MustParse("f22bd5ed-9155-4ba0-90c4-4880912d7ad4"),
				Balance:  100,
			},
			expectedStatus: http.StatusOK,
			expectedResponse: response.Response{
				Status: response.StatusOK,
			},
		},
		{
			name:           "successful fetch wallet 2",
			walletUUID:     "a45c73fd-3e36-466a-8e57-15e1cf0f35d2",
			mockWalletUUID: uuid.MustParse("a45c73fd-3e36-466a-8e57-15e1cf0f35d2"),
			mockWallet: postgresql.Wallet{
				WalletID: uuid.MustParse("a45c73fd-3e36-466a-8e57-15e1cf0f35d2"),
				Balance:  250,
			},
			expectedStatus: http.StatusOK,
			expectedResponse: response.Response{
				Status: response.StatusOK,
			},
		},
		{
			name:           "wallet not found",
			walletUUID:     "b1234567-89ab-cdef-0123-456789abcdef",
			mockWalletUUID: uuid.MustParse("b1234567-89ab-cdef-0123-456789abcdef"),
			mockErr:        fmt.Errorf("wallet not found"),
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.Response{
				Status: response.StatusError,
				Error:  "failed to fetch wallet",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock для GetterWallet
			mockGetterWallet := new(MockGetterWallet)
			if tt.mockErr == nil {
				mockGetterWallet.On("GetWallet", tt.mockWalletUUID).Return(tt.mockWallet, nil)
			} else {
				mockGetterWallet.On("GetWallet", tt.mockWalletUUID).Return(postgresql.Wallet{}, tt.mockErr)
			}

			// Создаем тестовый сервер с маршрутом
			r := chi.NewRouter()
			r.Get("/api/v1/wallets/{WALLET_UUID}", FetchWallet(nil, mockGetterWallet))

			req := httptest.NewRequest("GET", "/api/v1/wallets/"+tt.walletUUID, nil)
			rec := httptest.NewRecorder()

			fmt.Printf("Sending request: %s\n", req.URL)

			r.ServeHTTP(rec, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Проверяем ответ
			var res response.Response
			if err := render.DecodeJSON(rec.Body, &res); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			assert.Equal(t, tt.expectedResponse, res)

			// Проверяем, что mock был вызван корректно
			mockGetterWallet.AssertExpectations(t)
		})
	}
}
