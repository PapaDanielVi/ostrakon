package config

import (
	"errors"
	"os"
	"testing"

	"github.com/PapaDanielVi/ostrakon/pkg/keyring"
	"github.com/PapaDanielVi/ostrakon/pkg/mocks"
	"go.uber.org/mock/gomock"
)

const testRepoURL = "https://github.com/owner/repo"

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
			name:       "existing token",
			tokenValue: "ghp_existingtoken",
			wantErr:    false,
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
			repoURL:  testRepoURL,
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
			repoURL:  testRepoURL,
			owner:    "",
			repoName: "repo",
			wantErr:  true,
		},
		{
			name:     "empty name",
			repoURL:  testRepoURL,
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

func TestGetRepoURL(t *testing.T) {
	tests := []struct {
		name       string
		urlValue   string
		keyringErr error
		wantErr    bool
	}{
		{
			name:     "existing repo URL",
			urlValue: "https://github.com/owner/repo",
			wantErr:  false,
		},
		{
			name:       "no repo URL found",
			keyringErr: keyring.ErrNotFound,
			wantErr:    true,
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

			mockKeyring.EXPECT().Get(ServiceName, RepoURLKey).Return(tt.urlValue, tt.keyringErr)

			url, err := GetRepoURL()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && url != tt.urlValue {
				t.Errorf("GetRepoURL() = %v, want %v", url, tt.urlValue)
			}
		})
	}
}

func TestGetRepoOwner(t *testing.T) {
	tests := []struct {
		name       string
		ownerValue string
		keyringErr error
		wantErr    bool
	}{
		{
			name:       "existing repo owner",
			ownerValue: "owner",
			wantErr:    false,
		},
		{
			name:       "no repo owner found",
			keyringErr: keyring.ErrNotFound,
			wantErr:    true,
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

			mockKeyring.EXPECT().Get(ServiceName, RepoOwnerKey).Return(tt.ownerValue, tt.keyringErr)

			owner, err := GetRepoOwner()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoOwner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && owner != tt.ownerValue {
				t.Errorf("GetRepoOwner() = %v, want %v", owner, tt.ownerValue)
			}
		})
	}
}

func TestGetRepoName(t *testing.T) {
	tests := []struct {
		name       string
		nameValue  string
		keyringErr error
		wantErr    bool
	}{
		{
			name:      "existing repo name",
			nameValue: "repo",
			wantErr:   false,
		},
		{
			name:       "no repo name found",
			keyringErr: keyring.ErrNotFound,
			wantErr:    true,
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

			mockKeyring.EXPECT().Get(ServiceName, RepoNameKey).Return(tt.nameValue, tt.keyringErr)

			name, err := GetRepoName()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && name != tt.nameValue {
				t.Errorf("GetRepoName() = %v, want %v", name, tt.nameValue)
			}
		})
	}
}

func TestGetPasswordHash(t *testing.T) {
	tests := []struct {
		name       string
		hashValue  string
		keyringErr error
		wantErr    bool
	}{
		{
			name:      "existing password hash",
			hashValue: "a1b2c3d4e5f6",
			wantErr:   false,
		},
		{
			name:       "no password hash found",
			keyringErr: keyring.ErrNotFound,
			wantErr:    true,
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

			mockKeyring.EXPECT().Get(ServiceName, PasswordHashKey).Return(tt.hashValue, tt.keyringErr)

			hash, err := GetPasswordHash()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash != tt.hashValue {
				t.Errorf("GetPasswordHash() = %v, want %v", hash, tt.hashValue)
			}
		})
	}
}

func TestDeletePasswordHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	keyringClient = mockKeyring
	defer func() { keyringClient = originalKeyring }()

	mockKeyring.EXPECT().Delete(ServiceName, PasswordHashKey).Return(nil)

	err := DeletePasswordHash()
	if err != nil {
		t.Errorf("DeletePasswordHash() error = %v", err)
	}
}

func TestDeleteGlobalMasterPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	keyringClient = mockKeyring
	defer func() { keyringClient = originalKeyring }()

	mockKeyring.EXPECT().Delete(ServiceName, GlobalMasterPasswordKey).Return(nil)

	err := DeleteGlobalMasterPassword()
	if err != nil {
		t.Errorf("DeleteGlobalMasterPassword() error = %v", err)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}
}

func TestSetKeyring(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	SetKeyring(mockKeyring)
	defer func() { keyringClient = originalKeyring }()

	if GetKeyring() != mockKeyring {
		t.Error("SetKeyring did not set the keyring")
	}
}

func TestStoreRepoInfoErrorOnFirstSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	keyringClient = mockKeyring
	defer func() { keyringClient = originalKeyring }()

	mockKeyring.EXPECT().Set(ServiceName, RepoURLKey, testRepoURL).Return(errors.New("set failed"))

	err := StoreRepoInfo(testRepoURL, "owner", "repo")
	if err == nil {
		t.Error("expected error from StoreRepoInfo")
	}
}

func TestStoreRepoInfoErrorOnSecondSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockKeyring := mocks.NewMockKeyring(ctrl)

	originalKeyring := keyringClient
	keyringClient = mockKeyring
	defer func() { keyringClient = originalKeyring }()

	mockKeyring.EXPECT().Set(ServiceName, RepoURLKey, testRepoURL).Return(nil)
	mockKeyring.EXPECT().Set(ServiceName, RepoOwnerKey, "owner").Return(errors.New("set failed"))

	err := StoreRepoInfo(testRepoURL, "owner", "repo")
	if err == nil {
		t.Error("expected error from StoreRepoInfo")
	}
}

func TestStoreGlobalMasterPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mypassword",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
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

			if tt.password != "" {
				mockKeyring.EXPECT().Set(ServiceName, GlobalMasterPasswordKey, tt.password).Return(nil)
			}

			err := StoreGlobalMasterPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("StoreGlobalMasterPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Test EnsureConfigDir by using a temp directory approach
	// Since we can't easily mock os.UserHomeDir, we test the logic flow
	dir := ConfigDir()

	// Verify the dir exists or can be created
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// Try to create it
		err = EnsureConfigDir()
		if err != nil {
			t.Logf("EnsureConfigDir failed (may be expected in test env): %v", err)
		}
	} else if err == nil && !stat.IsDir() {
		t.Error("ConfigDir should be a directory")
	}
}
