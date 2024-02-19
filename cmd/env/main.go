package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DimmyJing/personal/app"
)

const minDirLen = 5

func findFile(filename string) []string {
	candidates := []string{}
	//nolint:gomnd,dogsled
	_, b, _, _ := runtime.Caller(2)
	dir := filepath.Dir(b)

	for {
		if len(dir) < minDirLen {
			break
		} else if _, err := os.Stat(filepath.Join(dir, filename)); !errors.Is(err, os.ErrNotExist) {
			candidates = append(candidates, filepath.Join(dir, filename))
		}

		dir = filepath.Dir(dir)
	}

	return candidates
}

var errKeyNotFound = errors.New("key not found")

func getKey() ([]byte, error) {
	if key, found := os.LookupEnv("KEY"); found {
		res, err := hex.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: %w", err)
		}

		return res, nil
	}

	for _, candidate := range findFile(".key") {
		dat, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}

		res, err := hex.DecodeString(string(dat))
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: %w", err)
		}

		return res, nil
	}

	return nil, errKeyNotFound
}

func encrypt(key []byte, val string) (string, error) {
	aesIV := make([]byte, aes.BlockSize)

	if _, err := rand.Read(aesIV); err != nil {
		return "", fmt.Errorf("failed to read random: %w", err)
	}

	rawValue := []byte(val)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create new cipher: %w", err)
	}

	cfb := cipher.NewCFBEncrypter(block, aesIV)
	cipherText := make([]byte, len(rawValue))
	cfb.XORKeyStream(cipherText, rawValue)
	//nolint:makezero
	cipherText = append(aesIV, cipherText...)
	encryptedValue := base64.StdEncoding.EncodeToString(cipherText)

	return "enc:" + encryptedValue, nil
}

func decrypt(key []byte, val string) (string, error) {
	if !strings.HasPrefix(val, "enc:") {
		return val, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create new cipher: %w", err)
	}

	rawValue, err := base64.StdEncoding.DecodeString(val[4:])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	iv := rawValue[:aes.BlockSize]
	cipherText := rawValue[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)

	return string(plainText), nil
}

func setEnvJSON(envJSON []byte, key []byte, envKey string, envVal string, enc bool) ([]byte, error) {
	envVars := make(map[string]string)

	err := json.Unmarshal(envJSON, &envVars)
	if err != nil {
		return nil, fmt.Errorf("error umarshalling env json: %w", err)
	}

	if enc {
		encryptedValue, err := encrypt(key, envVal)
		if err != nil {
			return nil, fmt.Errorf("error encrypting value: %w", err)
		}

		envVars[envKey] = encryptedValue
	} else {
		envVars[envKey] = envVal
	}

	newEnvJSON, err := json.MarshalIndent(envVars, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling env json: %w", err)
	}

	return newEnvJSON, nil
}

func initEnv(envJSON []byte, secretKey []byte) error {
	envVars := make(map[string]string)

	err := json.Unmarshal(envJSON, &envVars)
	if err != nil {
		return fmt.Errorf("failed to unmarshal env vars: %w", err)
	}

	for key, value := range envVars {
		if _, found := os.LookupEnv(key); !found {
			decryptedValue, err := decrypt(secretKey, value)
			if err != nil {
				return fmt.Errorf("failed to decrypt value: %w", err)
			}

			err = os.Setenv(key, decryptedValue)
			if err != nil {
				return fmt.Errorf("failed to set env var: %w", err)
			}
		}
	}

	return nil
}

func main() {
	setOp := flag.Bool("set", false, "set an env var")
	plain := flag.Bool("plain", false, "set a plaintext env var")

	flag.Parse()

	key, err := getKey()
	if err != nil {
		log.Fatal("failed to get key:", err)
	}

	switch {
	case *setOp:
		//nolint:gomnd
		if flag.NArg() != 2 {
			log.Fatal("usage: env -set <key> <value>")
		}

		envKey := flag.Arg(0)
		envVal := flag.Arg(1)

		res, err := setEnvJSON(app.EnvJSON, key, envKey, envVal, !*plain)
		if err != nil {
			log.Fatal("failed to retrieve modified env json:", err)
		}

		err = app.SetEnvJSON(res)
		if err != nil {
			log.Fatal("failed to set env json:", err)
		}

		log.Print("successfully set env var", envKey)
	default:
		err := initEnv(app.EnvJSON, key)
		if err != nil {
			log.Fatal("failed to init env:", err)
		}

		//nolint:gomnd
		if len(os.Args) < 2 {
			log.Fatal("no command provided")
		}

		path := os.Args[1]
		args := os.Args[2:]
		cmd := exec.Command(path, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		//nolint:errorlint
		if err, ok := err.(*exec.ExitError); ok {
			os.Exit(err.ExitCode())
		}

		if err != nil {
			log.Fatal("failed to run command:", err)
		}
	}
}
