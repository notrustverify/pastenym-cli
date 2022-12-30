package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
)

func encrypt(plaintext *[]byte) (string, string, encParams) {
	passphrase, key, salt := genKey(32)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)

	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	ciphertextBytes := aesGCM.Seal(nonce, nonce, *plaintext, nil)

	cipherTextString := base64.StdEncoding.EncodeToString(ciphertextBytes[aesGCM.NonceSize():])
	//keyEncoded := base64.StdEncoding.EncodeToString(key)

	// parameters used by sjcl
	encParams := encParams{
		Iv:     base64.StdEncoding.EncodeToString(nonce),
		Salt:   base64.StdEncoding.EncodeToString(salt),
		Ks:     256,
		V:      1,
		Mode:   "gcm",
		Cipher: "aes",
		Iter:   10000,
		Ts:     128,
	}

	return passphrase, cipherTextString, encParams
}

func decrypt(passphrase string, encryptedString *string, encParams encParams) string {
	decodeSalt, _ := base64.StdEncoding.DecodeString(encParams.Salt)
	key, _ := deriveKey(passphrase, []byte(decodeSalt))

	encryptedText, err := base64.StdEncoding.DecodeString(*encryptedString)
	if err != nil {
		panic(err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)

	if err != nil {
		fmt.Printf("\n%sUnable to decrypt%s\n", Red, Reset)
		if *debug {
			panic(err.Error())
		}
	}

	//adataDecode, err := base64.StdEncoding.DecodeString(encParams.Adata)
	ivDecode, _ := base64.StdEncoding.DecodeString(encParams.Iv)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesGCM.Open(nil, ivDecode, encryptedText, nil)
	if err != nil {
		fmt.Printf("\n%sUnable to decrypt%s\n", Red, Reset)
		if *debug {
			panic(err.Error())
		}
	}

	return string(plaintext)

}

func genKey(size uint8) (string, []byte, []byte) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	passphrase := make([]byte, size)

	var i uint8
	for i = 0; i < size; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err.Error())
		}
		passphrase[i] = letters[num.Int64()]
	}

	key, salt := deriveKey(string(passphrase), nil)
	return string(passphrase), key, salt //encode key in bytes to string for saving

}

func deriveKey(passphrase string, salt []byte) ([]byte, []byte) {
	if salt == nil {
		salt = make([]byte, 8)
		// http://www.ietf.org/rfc/rfc2898.txt
		// Salt.
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(passphrase), salt, 10000, 32, sha256.New), salt
}
