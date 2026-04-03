package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"golang.org/x/crypto/ssh"
	"os"
)

func GenerateSshED25519() (string, string, error) {
	// 1. 生成 ED25519 密钥对
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		config.Logger.Error(err)
		return "", "", err
	}
	private, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		config.Logger.Error(err)
		return "", "", err
	}
	// 2. 将私钥转换为 PEM 格式（OpenSSH 兼容）
	privateKeyPEM := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: private.Bytes, // 密码可选，空字符串表示无密码
	}

	// 3. 使用 golang.org/x/crypto/ssh 生成 SSH 公钥格式
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		config.Logger.Error(err)
		return "", "", err
	}

	// 4. 格式化为标准 OpenSSH 公钥字符串（如：ssh-ed25519 AAAAC3... user@host）
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)
	// 5. 输出结果
	pem.Encode(os.Stdout, privateKeyPEM)

	return string(publicKeyBytes), string(privateKey.Seed()), err
}
func GenerateEd25519() (string, string, error) {
	var publicKey []byte
	var privateKey []byte
	var err error
	if publicKey, privateKey, err = ed25519.GenerateKey(nil); err != nil {
		err = fmt.Errorf("decode private key fail: %s", err.Error())
		return "", "", err
	}
	return string(publicKey), string(privateKey), nil
}
