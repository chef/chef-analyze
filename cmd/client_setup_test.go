package cmd

import (
	"errors"
	"testing"

	"github.com/chef/chef-analyze/pkg/reporting"
	"github.com/chef/go-libs/credentials"
	"github.com/go-chef/chef"
)

func resetSetupDeps() {
	credentialsFromViper = credentials.FromViper
	createOutputDirs = createOutputDirectories
	newChefClient = reporting.NewChefClient
	infraFlags.noSSLverify = false
}

func TestSetupChefClientFromFlags_CredentialsError(t *testing.T) {
	defer resetSetupDeps()

	expectedErr := errors.New("credentials failure")
	credentialsFromViper = func(profile string, override ...credentials.OverrideFunc) (credentials.Credentials, error) {
		return credentials.Credentials{}, expectedErr
	}
	createOutputDirs = func() error {
		t.Fatal("createOutputDirs should not be called when credentials fail")
		return nil
	}
	newChefClient = func(cfg *reporting.Reporting) (*chef.Client, error) {
		t.Fatal("newChefClient should not be called when credentials fail")
		return nil, nil
	}

	client, err := setupChefClientFromFlags()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected credentials error %v, got %v", expectedErr, err)
	}
	if client != nil {
		t.Fatalf("expected nil client, got %#v", client)
	}
}

func TestSetupChefClientFromFlags_CreateOutputDirsError(t *testing.T) {
	defer resetSetupDeps()

	credentialsFromViper = func(profile string, override ...credentials.OverrideFunc) (credentials.Credentials, error) {
		return credentials.Credentials{}, nil
	}
	expectedErr := errors.New("mkdir failure")
	createOutputDirs = func() error {
		return expectedErr
	}
	newChefClient = func(cfg *reporting.Reporting) (*chef.Client, error) {
		t.Fatal("newChefClient should not be called when directory creation fails")
		return nil, nil
	}

	client, err := setupChefClientFromFlags()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected output directory error %v, got %v", expectedErr, err)
	}
	if client != nil {
		t.Fatalf("expected nil client, got %#v", client)
	}
}

func TestSetupChefClientFromFlags_PassesNoSSLVerifyFlag(t *testing.T) {
	defer resetSetupDeps()

	infraFlags.noSSLverify = true
	credentialsFromViper = func(profile string, override ...credentials.OverrideFunc) (credentials.Credentials, error) {
		return credentials.Credentials{}, nil
	}
	createOutputDirs = func() error {
		return nil
	}
	newChefClient = func(cfg *reporting.Reporting) (*chef.Client, error) {
		if cfg == nil {
			t.Fatal("expected non-nil reporting config")
		}
		if !cfg.NoSSLVerify {
			t.Fatal("expected NoSSLVerify=true when infra flag is enabled")
		}
		return &chef.Client{}, nil
	}

	client, err := setupChefClientFromFlags()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestSetupChefClientFromFlags_Success(t *testing.T) {
	defer resetSetupDeps()

	credentialsFromViper = func(profile string, override ...credentials.OverrideFunc) (credentials.Credentials, error) {
		return credentials.Credentials{}, nil
	}
	createOutputDirs = func() error {
		return nil
	}
	expectedClient := &chef.Client{}
	newChefClient = func(cfg *reporting.Reporting) (*chef.Client, error) {
		if cfg == nil {
			t.Fatal("expected non-nil reporting config")
		}
		if cfg.NoSSLVerify {
			t.Fatal("expected NoSSLVerify=false by default")
		}
		return expectedClient, nil
	}

	client, err := setupChefClientFromFlags()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client != expectedClient {
		t.Fatalf("expected %p, got %p", expectedClient, client)
	}
}
