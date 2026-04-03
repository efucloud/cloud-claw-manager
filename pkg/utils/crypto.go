package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
)

type PaddingMode string

const PKCS5 PaddingMode = "PKCS5"
const PKCS7 PaddingMode = "PKCS7"
const ZEROS PaddingMode = "ZEROS"

func Padding(padding PaddingMode, src []byte, blockSize int) []byte {
	switch padding {
	case PKCS5:
		src = PKCS5Padding(src, blockSize)
	case PKCS7:
		src = PKCS7Padding(src, blockSize)
	case ZEROS:
		src = ZerosPadding(src, blockSize)
	}
	return src
}

func UnPadding(padding PaddingMode, src []byte) ([]byte, error) {
	switch padding {
	case PKCS5:
		return PKCS5UnPadding(src)
	case PKCS7:
		return PKCS7UnPadding(src)
	case ZEROS:
		return ZerosUnPadding(src)
	}
	return src, nil
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	return PKCS7Padding(src, blockSize)
}

func PKCS5UnPadding(src []byte) ([]byte, error) {
	return PKCS7UnPadding(src)
}

func PKCS7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS7UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return src, fmt.Errorf("src length is 0")
	}
	unpadding := int(src[length-1])
	if length < unpadding {
		return src, fmt.Errorf("src length is less than unpadding")
	}
	return src[:(length - unpadding)], nil
}

func ZerosPadding(src []byte, blockSize int) []byte {
	rem := len(src) % blockSize
	if rem == 0 {
		return src
	}
	return append(src, bytes.Repeat([]byte{0}, blockSize-rem)...)
}

func ZerosUnPadding(src []byte) ([]byte, error) {
	for i := len(src) - 1; ; i-- {
		if src[i] != 0 {
			return src[:i+1], nil
		}
	}
}

// AesSimpleEncrypt encrypts data with key using AES algorithm.
// In simple encryption mode, the user only needs to specify the key to complete the encryption.
// IV will be obtained by hashing the key. By default, PKCS7Padding and CBC modes are used.
// Return empty string if error occurs.
func AesSimpleEncrypt(data, key string) string {
	key = trimByMaxKeySize(key)
	keyBytes := ZerosPadding([]byte(key), aes.BlockSize)
	return AesCBCEncrypt(data, string(keyBytes), GenIVFromKey(key), PKCS7)
}

// AesCBCEncrypt encrypts data with key and iv using AES algorithm.
// You must make sure the length of key and iv is 16 bytes. This function does not perform any padding for key or iv.
func AesCBCEncrypt(data, key, iv string, paddingMode PaddingMode) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}

	src := Padding(paddingMode, []byte(data), block.BlockSize())
	encryptData := make([]byte, len(src))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(encryptData, src)
	return base64.StdEncoding.EncodeToString(encryptData)
}

// AesSimpleDecrypt decrypts data with key using AES algorithm.
// In simple decryption mode, the user only needs to specify the key to complete the decryption.
// This function will automatically obtain the IV by hashing the key.
func AesSimpleDecrypt(data, key string) string {
	key = trimByMaxKeySize(key)
	keyBytes := ZerosPadding([]byte(key), aes.BlockSize)
	return AesCBCDecrypt(data, string(keyBytes), GenIVFromKey(key), PKCS7)
}

// AesCBCDecrypt decrypts data with key and iv using AES algorithm.
func AesCBCDecrypt(data, key, iv string, paddingMode PaddingMode) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}

	decodeData, _ := base64.StdEncoding.DecodeString(data)
	decryptData := make([]byte, len(decodeData))
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(decryptData, decodeData)

	original, _ := UnPadding(paddingMode, decryptData)
	return string(original)
}

// GenIVFromKey generates IV from key.
func GenIVFromKey(key string) (iv string) {
	hashedKey := sha256.Sum256([]byte(key))
	return trimByBlockSize(hex.EncodeToString(hashedKey[:]))
}

func trimByBlockSize(key string) string {
	if len(key) > aes.BlockSize {
		return key[:aes.BlockSize]
	}
	return key
}

func trimByMaxKeySize(key string) string {
	if len(key) > 32 {
		return key[:32]
	}
	return key
}

const secretKey = "asdfS2324A><:"
const secretPrefix = "efuCloudSecretData:"

// ClusterCertDataDecrypt 集群数据解密
func ClusterCertDataDecrypt(data string) (result string, err error) {
	var deBytes []byte
	if strings.HasPrefix(data, secretPrefix) {
		data = strings.TrimPrefix(data, secretPrefix)
		result = AesSimpleDecrypt(data, secretKey)
	} else {
		// 尝试base64解码，失败则保持原数据
		deBytes, err = base64.StdEncoding.DecodeString(data)
		if err == nil {
			result = string(deBytes)
		}
	}
	if result == "" {
		result = data
	}
	return result, err
}

// ClusterCertDataEncrypt 集群数据加密
func ClusterCertDataEncrypt(data string) (result string, err error) {
	var deBytes []byte
	// 尝试base64解码，失败则保持原数据
	deBytes, err = base64.StdEncoding.DecodeString(data)
	if err == nil {
		data = string(deBytes)
	}
	result = AesSimpleEncrypt(data, secretKey)
	if result == "" {
		result = data
	}
	result = secretPrefix + result
	return result, err

}

type Ed25519KeyPair struct {
	PrivateKeyPEM    string // 私钥，PEM 格式（PKCS#8），可安全存入数据库（需加密）
	PublicKeyOpenSSH string // 公钥，OpenSSH 格式，如 "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."
}

func GenerateEd25519KeyPair() (private string, public string, err error) {
	// 1. 生成 Ed25519 密钥对
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	// 2. 将私钥编码为 PKCS#8 PEM 格式（通用且标准）
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY", // PKCS#8
		Bytes: privateKeyBytes,
	})

	// 3. 将公钥编码为 OpenSSH 格式（如 ssh-ed25519 AAAA...）
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create SSH public key: %w", err)
	}
	publicKeyOpenSSH := string(ssh.MarshalAuthorizedKey(sshPublicKey))

	return string(privatePEM), publicKeyOpenSSH, nil
}
