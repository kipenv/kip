package cli

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/kipenv/kip/internal/client"
	"github.com/kipenv/kip/internal/config"
	"github.com/kipenv/kip/internal/crypto"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	var (
		maxReads   int
		ttlSeconds int
		password   string
	)

	cmd := &cobra.Command{
		Use:   "push <file>",
		Short: "Encrypt and share a .env file",
		Long:  "Encrypts a file with AES-256-GCM and uploads the ciphertext to the server. Returns a shareable link with the decryption key in the URL fragment.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			plaintext, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			if len(plaintext) == 0 {
				return fmt.Errorf("file is empty: %s", filePath)
			}

			// First layer: encrypt with random key (key goes in URL fragment)
			ciphertext, nonce, key, err := crypto.Encrypt(plaintext)
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}

			var salt []byte
			passwordProtected := password != ""

			if passwordProtected {
				// Pack nonce + ciphertext so we can recover nonce after password decryption
				packed := make([]byte, len(nonce)+len(ciphertext))
				copy(packed, nonce)
				copy(packed[len(nonce):], ciphertext)

				// Second layer: encrypt the packed data with password
				ciphertext, nonce, salt, err = crypto.EncryptWithPassword(packed, []byte(password))
				if err != nil {
					return fmt.Errorf("password encrypt: %w", err)
				}
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			createReq := client.CreateSecretRequest{
				Ciphertext:        base64.StdEncoding.EncodeToString(ciphertext),
				Nonce:             base64.StdEncoding.EncodeToString(nonce),
				Filename:          extractFilename(filePath),
				MaxReads:          maxReads,
				TTLSeconds:        ttlSeconds,
				PasswordProtected: passwordProtected,
			}
			if len(salt) > 0 {
				createReq.Salt = base64.StdEncoding.EncodeToString(salt)
			}

			c := client.New(cfg.ServerURL)
			resp, err := c.CreateSecret(cmd.Context(), createReq)
			if err != nil {
				return fmt.Errorf("upload: %w", err)
			}

			encodedKey := crypto.EncodeKey(key)
			link := fmt.Sprintf("%s/s/%s#%s", cfg.ServerURL, resp.ID, encodedKey)

			fmt.Fprintf(cmd.OutOrStdout(), "Secret shared successfully!\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Link: %s\n", link)
			fmt.Fprintf(cmd.OutOrStdout(), "Expires: %s (%s) | Reads: %d\n",
				resp.ExpiresAt.Local().Format("2006-01-02 15:04"),
				formatDuration(time.Duration(ttlSeconds)*time.Second),
				maxReads)
			if passwordProtected {
				fmt.Fprintf(cmd.OutOrStdout(), "Password: required (share it separately)\n")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nShare this link. The key never leaves your machine.\n")

			return nil
		},
	}

	cmd.Flags().IntVarP(&maxReads, "reads", "r", 1, "maximum number of reads before the secret is deleted")
	cmd.Flags().IntVarP(&ttlSeconds, "ttl", "t", 3600, "time-to-live in seconds (default 1 hour)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "add password protection (recipient must know the password)")

	return cmd
}

// formatDuration returns a human-readable duration like "1h 30m" or "2d 12h".
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("in %dd %dh", days, hours)
	case days > 0:
		return fmt.Sprintf("in %dd", days)
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("in %dh %dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("in %dh", hours)
	default:
		return fmt.Sprintf("in %dm", minutes)
	}
}

// extractFilename extracts the base filename from a path.
func extractFilename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
