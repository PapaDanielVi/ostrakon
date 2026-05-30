package config

import (
	"errors"
	"testing"

	"github.com/PapaDanielVi/ostrakon/pkg/keyring"
	"github.com/PapaDanielVi/ostrakon/pkg/mocks"
	"go.uber.org/mock/gomock"
)

func TestStoreToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "ghp_testtoken123",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockKeyring := mocks.NewMockKeyring(ctrl)

			originalKeyring := keyringClient
			keyringClient = mockKeyring
			defer func() { keyringClient = originalKeyring }()

			if tt.token != "" {
				mockKeyring.EXPECT().Set(ServiceName, TokenKey, tt.token).Return(nil)
			}

			err := StoreToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetToken(t *testing.T) {
	tests := []struct {
		name       string
		tokenValue string
		keyringErr error
		wantErr    bool
		errMsg     string
	}{
		{
			name:      "existing token",
			tokenValue: "ghp_existingtoken",
			wantErr:   false,
		},
		{
			name:       "no token found",
			keyringErr: keyring.ErrNotFound,
			wantErr:    true,
			errMsg:     "no token found",
		},
		{
			name:       "keyring error",
			keyringErr: errors.New("keyring failure"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockKeyring := mocks.NewMockKeyring(ctrl)

			originalKeyring := keyringClient
			keyringClient = mockKeyring
			defer func() { keyringClient = originalKeyring }()

			mockKeyring.EXPECT().Get(ServiceName, TokenKey).Return(tt.tokenValue, tt.keyringErr)

			token, err := GetToken()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token != tt.tokenValue {
				t.Errorf("GetToken() = %v, want %v", token, tt.tokenValue)
			}
		})
	}
}

func TestStoreRepoInfo(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		owner    string
		repoName string
		wantErr  bool
	}{
		{
			name:     "valid repo info",
			repoURL:  "https://github.com/owner/repo",
			owner:    "owner",
			repoName: "repo",
			wantErr:  false,
		},
		{
			name:     "empty url",
			repoURL:  "",
			owner:    "owner",
			repoName: "repo",
			wantErr:  true,
		},
		{
			name:     "empty owner",
			repoURL:  "https://github.com/owner/repo",
			owner:    "",
			repoName: "repo",
			wantErr:  true,
		},
		{
			name:     "empty name",
			repoURL:  "https://github.com/owner/repo",
			owner:    "owner",
			repoName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockKeyring := mocks.NewMockKeyring(ctrl)

			originalKeyring := keyringClient
			keyringClient = mockKeyring
			defer func() { keyringClient = originalKeyring }()

			if tt.repoURL != "" && tt.owner != "" && tt.repoName != "" {
				mockKeyring.EXPECT().Set(ServiceName, RepoURLKey, tt.repoURL).Return(nil)
				mockKeyring.EXPECT().Set(ServiceName, RepoOwnerKey, tt.owner).Return(nil)
				mockKeyring.EXPECT().Set(ServiceName, RepoNameKey, tt.repoName).Return(nil)
			}

			err := StoreRepoInfo(tt.repoURL, tt.owner, tt.repoName)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreRepoInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorePasswordHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid hash",
			hash:    "a1b2c3d4e5f6",
			wantErr: false,
		},
		{
			name:    "empty hash",
			hash:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockKeyring := mocks.NewMockKeyring(ctrl)

			originalKeyring := keyringClient
			keyringClient = mockKeyring
			defer func() { keyringClient = originalKeyring }()

			if tt.hash != "" {
				mockKeyring.EXPECT().Set(ServiceName, PasswordHashKey, tt.hash).Return(nil)
			}

			err := StorePasswordHash(tt.hash)

			if (err != nil) != tt.wantErr {
				t.Errorf("StorePasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasGlobalMasterPassword(t *testing.T) {
	tests := []struct {
		name   string
		hasErr error
		want   bool
	}{
		{
			name:   "password exists",
			hasErr: nil,
			want:   true,
		},
		{
			name:   "password not found",
			hasErr: keyring.ErrNotFound,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockKeyring := mocks.NewMockKeyring(ctrl)

			originalKeyring := keyringClient
			keyringClient = mockKeyring
			defer func() { keyringClient = originalKeyring }()

			mockKeyring.EXPECT().Get(ServiceName, GlobalMasterPasswordKey).Return("password", tt.hasErr)

			got := HasGlobalMasterPassword()

			if got != tt.want {
				t.Errorf("HasGlobalMasterPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGlobalMasterPassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		keyringErr error
		wantErr    bool
	}{
		{
			name:     "existing password",
			password: "mypassword",
			wantErr:  false,
		},
		{
			name:       "password not found",
			password:   "",
			keyringErr: keyring.ErrNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockKeyring := mocks.NewMockKeyring(ctrl)

			originalKeyring := keyringClient
			keyringClient = mockKeyring
			defer func() { keyringClient = originalKeyring }()

			mockKeyring.EXPECT().Get(ServiceName, GlobalMasterPasswordKey).Return(tt.password, tt.keyringErr)

			pw, err := GetGlobalMasterPassword()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetGlobalMasterPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && pw != tt.password {
				t.Errorf("GetGlobalMasterPassword() = %v, want %v", pw, tt.password)
			}
		})
	}
}

func TestDeleteToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	keyringClient = mockKeyring
	defer func() { keyringClient = originalKeyring }()

	mockKeyring.EXPECT().Delete(ServiceName, TokenKey).Return(nil)

	err := DeleteToken()
	if err != nil {
		t.Errorf("DeleteToken() error = %v", err)
	}
}

func TestDeleteRepoInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	keyringClient = mockKeyring
	defer func() { keyringClient = originalKeyring }()

	mockKeyring.EXPECT().Delete(ServiceName, RepoURLKey).Return(nil)
	mockKeyring.EXPECT().Delete(ServiceName, RepoOwnerKey).Return(nil)
	mockKeyring.EXPECT().Delete(ServiceName, RepoNameKey).Return(nil)

	err := DeleteRepoInfo()
	if err != nil {
		t.Errorf("DeleteRepoInfo() error = %v", err)
	}
}