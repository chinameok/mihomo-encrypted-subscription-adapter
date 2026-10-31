// mihomo-subscription-adapter updates a local Mihomo provider from a
// FlClash-encrypted subscription. It never prints the URL or password.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chinameok/mihomo-encrypted-subscription-adapter/internal/controller"
	"github.com/chinameok/mihomo-encrypted-subscription-adapter/internal/credentials"
	"github.com/chinameok/mihomo-encrypted-subscription-adapter/internal/decrypt"
	"github.com/chinameok/mihomo-encrypted-subscription-adapter/internal/fetch"
	"github.com/chinameok/mihomo-encrypted-subscription-adapter/internal/provider"
	"github.com/chinameok/mihomo-encrypted-subscription-adapter/internal/runtime"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, redactError(err))
		os.Exit(1)
	}
}

func run(args []string) error {
	root := defaultRoot()
	if v := os.Getenv("MEOK_SUBSCRIPTION_ADAPTER_DIR"); v != "" {
		root = v
	}
	if len(args) == 0 {
		return errors.New("usage: mihomo-subscription-adapter <refresh|status|set-password|set-url|rollback|enable|disable>")
	}
	store := credentials.FileStore{Dir: root}
	mgr := runtime.Manager{Dir: root}
	switch args[0] {
	case "refresh":
		fs := flag.NewFlagSet("refresh", flag.ContinueOnError)
		dryRun := fs.Bool("dry-run", false, "validate without changing local provider state")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return refresh(store, mgr, *dryRun)
	case "status":
		status, err := mgr.Status()
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(status)
	case "set-password":
		value, err := readSecret("Password: ")
		if err != nil {
			return err
		}
		return store.Set("password", value)
	case "set-url":
		value, err := readLine("Subscription URL: ")
		if err != nil {
			return err
		}
		u, err := url.Parse(strings.TrimSpace(value))
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return errors.New("subscription URL must be an https URL")
		}
		return store.Set("url", value)
	case "rollback":
		return mgr.Rollback(context.Background())
	case "disable":
		return mgr.Disable(context.Background())
	case "enable":
		return mgr.Enable(context.Background())
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func refresh(store credentials.FileStore, mgr runtime.Manager, dryRun bool) error {
	urlValue, err := store.Get("url")
	if err != nil {
		return fail(mgr, "credentials_missing", fmt.Errorf("read subscription URL: %w", err))
	}
	password, err := store.Get("password")
	if err != nil {
		return fail(mgr, "credentials_missing", fmt.Errorf("read subscription password: %w", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client := fetch.Client{HTTPClient: &http.Client{Timeout: 45 * time.Second}}
	response, err := client.Do(ctx, urlValue)
	if err != nil {
		return fail(mgr, "download_failed", fmt.Errorf("download_failed: %w", err))
	}
	plaintext, err := decrypt.Decrypt(response.Body, password)
	if err != nil {
		return fail(mgr, "decrypt_failed", fmt.Errorf("decrypt_failed: %w", err))
	}
	result, err := provider.Normalize(plaintext)
	if err != nil {
		return fail(mgr, "invalid_profile", fmt.Errorf("invalid_profile: %w", err))
	}
	fmt.Fprintf(os.Stdout, "validated nodes=%d digest=%s\n", result.NodeCount, result.Digest)
	if dryRun {
		return nil
	}
	if err := mgr.Refresh(ctx, result.YAML, runtime.Metadata{NodeCount: result.NodeCount, Digest: result.Digest}); err != nil {
		return fail(mgr, "provider_write_failed", fmt.Errorf("provider_write_failed: %w", err))
	}
	if controllerURL := strings.TrimSpace(os.Getenv("MEOK_MIHOMO_CONTROLLER_URL")); controllerURL != "" {
		providerName := strings.TrimSpace(os.Getenv("MEOK_MIHOMO_PROVIDER_NAME"))
		if providerName == "" {
			providerName = "meok-subscription"
		}
		client := controller.Client{
			BaseURL: controllerURL,
			Secret:  os.Getenv("MEOK_MIHOMO_CONTROLLER_SECRET"),
		}
		if err := client.RefreshProvider(ctx, os.Getenv("MEOK_MIHOMO_PROVIDER_REFRESH_URL"), providerName); err != nil {
			rollbackErr := mgr.Rollback(ctx)
			if rollbackErr == nil {
				// Ask Mihomo to reload the restored last-known-good file. A
				// second failure is reported without exposing the controller URL.
				_ = client.RefreshProvider(ctx, os.Getenv("MEOK_MIHOMO_PROVIDER_REFRESH_URL"), providerName)
			}
			if rollbackErr != nil {
				return fail(mgr, "mihomo_refresh_failed", fmt.Errorf("mihomo_refresh_failed: %w; rollback failed: %v", err, rollbackErr))
			}
			return fail(mgr, "mihomo_refresh_failed", fmt.Errorf("mihomo_refresh_failed: %w", err))
		}
	}
	return nil
}

func fail(mgr runtime.Manager, class string, err error) error {
	_ = mgr.RecordFailure(context.Background(), class)
	return err
}

func readLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	s := bufio.NewScanner(os.Stdin)
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return "", err
		}
		return "", errors.New("no input")
	}
	return strings.TrimSpace(s.Text()), nil
}

func readSecret(prompt string) (string, error) {
	// The CLI deliberately avoids accepting secrets as arguments. When stdin is
	// a terminal, temporarily disable echo so the value is not displayed or
	// copied into terminal scrollback. launchd and automation use the protected
	// file/keychain store instead, so non-terminal stdin remains supported.
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprint(os.Stderr, prompt)
		if err := setTerminalEcho(false); err != nil {
			return "", fmt.Errorf("hide terminal input: %w", err)
		}
		defer func() {
			_ = setTerminalEcho(true)
			fmt.Fprintln(os.Stderr)
		}()
		return readLine("")
	}
	return readLine(prompt)
}

func setTerminalEcho(enabled bool) error {
	mode := "-echo"
	if enabled {
		mode = "echo"
	}
	cmd := exec.Command("stty", mode)
	cmd.Stdin = os.Stdin
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func defaultRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "subscription-adapter")
	}
	return filepath.Join(home, "Library", "Application Support", "meok-mihomo", "subscription-adapter")
}

func redactError(err error) string {
	// URL values are never included in errors by the core. Keep this boundary
	// for future transport errors that may echo request URLs.
	return strings.ReplaceAll(err.Error(), "password", "credential")
}
