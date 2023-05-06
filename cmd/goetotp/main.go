package main

import (
	"crypto/aes"
	"crypto/sha1"
	"encoding/base32"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/magiconair/properties"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/pbkdf2"
)

// Config represents .et-top.properties fields.
type Config struct {
	Encoded string `properties:"encoded"`
	Salt    string `properties:"salt"`
	Name    string `properties:"name"`
}

// ReadSecret is restore otp secret from config
// https://github.com/ecki/et-otp/blob/master/src/main/java/net/eckenfels/etotp/GUI.java#L426
// 以下が等価であるため、Go実装もECBで実装する
// - Cipher c1 = Cipher.getInstance("AES");
// - Cipher c1 = Cipher.getInstance("AES/ECB/PKCS5Padding");
// https://docs.oracle.com/javase/jp/8/docs/technotes/guides/security/crypto/CryptoSpec.html#trans
func (c Config) ReadSecret(pass []byte) ([]byte, error) {
	base32Encoder := base32.StdEncoding.WithPadding(base32.NoPadding)

	salt, err := base32Encoder.DecodeString(c.Salt)
	if err != nil {
		return nil, fmt.Errorf("salt decode base32: %w", err)
	}

	ciphertext, err := base32Encoder.DecodeString(c.Encoded)
	if err != nil {
		return nil, fmt.Errorf("encoded decode base32: %w", err)
	}

	// PBKDF2 With HMAC-SHA1
	secret := pbkdf2.Key(pass, salt, 1000, 16, sha1.New) // AES-128

	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("secret new cipher: %w", err)
	}

	// PKCS5Padding
	padCipherText := ciphertext

	// ECB mode
	plaintext := make([]byte, len(padCipherText))
	for i, j := 0, 16; i < len(padCipherText); i, j = i+16, j+16 {
		block.Decrypt(plaintext[i:j], padCipherText[i:j])
	}

	trimming, err := PKCS5Trimming(plaintext)
	if err != nil {
		return nil, errors.New("bad password")
	}

	return trimming, nil
}

// PKCS5Trimming is trimming PKCS5Padding
// ref: https://gist.github.com/hothero/7d085573f5cb7cdb5801d7adcf66dcf3
func PKCS5Trimming(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]

	if len(encrypt)-int(padding) < 0 {
		return nil, errors.New("invalid pkcs5 padding layout")
	}

	return encrypt[:len(encrypt)-int(padding)], nil
}

func main() {

	app := &cli.App{
		Name:  "goetotp",
		Usage: "run with .et-top.properties in the same directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "unlockpassword",
				Usage:   "Unlock Password for et-OTP. You can also use ${ETOTP_PASSWORD} as unlock password.",
				Aliases: []string{"pass"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			p := properties.MustLoadFile(".et-otp.properties", properties.UTF8).
				FilterPrefix("key.1").
				FilterStripPrefix("key.1.")

			var cfg Config
			if err := p.Decode(&cfg); err != nil {
				log.Fatal(err)
			}

			unlockPassword := cCtx.String("unlockpassword")
			if unlockPassword == "" {
				// Overwrite environment variable
				unlockPassword = os.Getenv("ETOTP_PASSWORD")
			}
			if unlockPassword == "" {
				fmt.Print("Enter unlock password: ")
				stdInput, _ := terminal.ReadPassword(int(syscall.Stdin))
				unlockPassword = string(stdInput)
				fmt.Println()
			}

			secret, err := cfg.ReadSecret([]byte(unlockPassword))
			if err != nil {
				log.Fatal("read secret: ", err)
			}

			code, err := totp.GenerateCode(base32.StdEncoding.EncodeToString(secret), time.Now())
			if err != nil {
				log.Fatal("totp generate code: ", err)
			}
			fmt.Println(code)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
