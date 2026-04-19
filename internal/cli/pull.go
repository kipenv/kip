package cli

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/kipenv/kip/internal/client"
	"github.com/kipenv/kip/internal/config"
	"github.com/kipenv/kip/internal/crypto"
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	var (
		output   string
		stdout   bool
		password string
	)

	cmd := &cobra.Command{
		Use:   "pull <link>",
		Short: "Download and decrypt a shared .env file",
		Long:  "Retrieves an encrypted secret from the server and decrypts it using the key from the URL fragment.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			link := args[0]

			id, encodedKey, err := parseLink(link)
			if err != nil {
				return err
			}

			key, err := crypto.DecodeKey(encodedKey)
			if err != nil {
				return fmt.Errorf("invalid key in link: %w", err)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			c := client.New(cfg.ServerURL)
			resp, err := c.GetSecret(id)
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}

			ciphertext, err := base64.StdEncoding.DecodeString(resp.Ciphertext)
			if err != nil {
				return fmt.Errorf("decode ciphertext: %w", err)
			}

			nonce, err := base64.StdEncoding.DecodeString(resp.Nonce)
			if err != nil {
				return fmt.Errorf("decode nonce: %w", err)
			}

			// If password-protected, decrypt the outer (password) layer first
			if resp.PasswordProtected {
				pw := password
				if pw == "" {
					pw, err = promptPassword(cmd)
					if err != nil {
						return err
					}
				}

				salt, err := base64.StdEncoding.DecodeString(resp.Salt)
				if err != nil {
					return fmt.Errorf("decode salt: %w", err)
				}

				packed, err := crypto.DecryptWithPassword(ciphertext, nonce, salt, []byte(pw))
				if err != nil {
					return fmt.Errorf("wrong password or corrupted data")
				}

				// Unpack: nonce (12 bytes) + inner ciphertext
				if len(packed) < crypto.NonceSize {
					return fmt.Errorf("corrupted data after password decryption")
				}
				nonce = packed[:crypto.NonceSize]
				ciphertext = packed[crypto.NonceSize:]
			}

			plaintext, err := crypto.Decrypt(ciphertext, nonce, key)
			if err != nil {
				return fmt.Errorf("decrypt: %w", err)
			}

			if stdout {
				fmt.Fprint(cmd.OutOrStdout(), string(plaintext))
				return nil
			}

			filename := output
			if filename == "" {
				filename = resp.Filename
			}
			if filename == "" {
				filename = ".env"
			}

			if err := os.WriteFile(filename, plaintext, 0o600); err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Saved to %s\n", filename)
			if resp.ReadsLeft > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Reads remaining: %d\n", resp.ReadsLeft)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "This was the last read — secret has been deleted.\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "save with a different filename")
	cmd.Flags().BoolVar(&stdout, "stdout", false, "print to console instead of saving to file")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password for protected links")

	return cmd
}

// promptPassword asks the user for a password via stdin.
func promptPassword(cmd *cobra.Command) (string, error) {
	fmt.Fprint(cmd.OutOrStderr(), "This secret is password-protected.\nPassword: ")
	reader := bufio.NewReader(os.Stdin)
	pw, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(pw, "\r\n"), nil
}

// parseLink extracts the secret ID and key from an kip link.
// Expected format: <baseURL>/s/<id>#<key>
func parseLink(link string) (id, key string, err error) {
	// Sanitize: shells may escape # as \# when pasting URLs
	link = strings.ReplaceAll(link, `\#`, "#")

	parts := strings.SplitN(link, "#", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", "", fmt.Errorf("invalid link: missing key fragment (#)")
	}

	key = parts[1]
	urlPath := parts[0]

	// Extract ID from path: .../s/<id>
	idx := strings.LastIndex(urlPath, "/s/")
	if idx == -1 {
		return "", "", fmt.Errorf("invalid link: expected /s/<id>#<key> format")
	}

	id = urlPath[idx+3:]
	if id == "" {
		return "", "", fmt.Errorf("invalid link: missing secret ID")
	}

	return id, key, nil
}
