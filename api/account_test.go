package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/longtk26/simple_bank/db/mock"
	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/token"
	"github.com/longtk26/simple_bank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name string
		accountID int64
		buildStubs func(store *mockdb.MockStore)
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.IMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	} {
		{
			name: "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.IMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "Not Found",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.IMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal Error",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.IMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid account id Error",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.IMaker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
		
			store := mockdb.NewMockStore(ctrl)
		
			// build stubs
			tc.buildStubs(store)

			// start test http server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()
		
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)	

			// check response
			tc.checkResponse(t, recorder)
		})
	}

}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	testCases := []struct {
		name string
		buildStubs func(store *mockdb.MockStore, account db.Account)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account)
	} {
		{
			name: "OK",
			buildStubs: func(store *mockdb.MockStore, account db.Account) {
				accountArgs := db.CreateAccountParams{
					Owner: account.Owner,
					Currency: account.Currency,
					Balance: 0,
				}
				
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(accountArgs)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "Internal Error",
			buildStubs: func(store *mockdb.MockStore, account db.Account) {
				accountArg := db.CreateAccountParams{
					Owner: account.Owner,
					Currency: account.Currency,
					Balance: 0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(accountArg)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Input",
			buildStubs: func(store *mockdb.MockStore, account db.Account) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i:= range testCases {
		tc := testCases[i]
		account := randomAccount(user.Username)

		if tc.name == "Invalid Input" {
			account.Owner = ""
		}

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			// build stubs
			tc.buildStubs(store, account)

			// start test http server
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			body, err := json.Marshal(account)
			require.NoError(t, err)

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder, account)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID: util.RandomInt(1, 1000),
		Owner: owner,
		Balance: util.RandomMoney(),	
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		Password: 		hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}