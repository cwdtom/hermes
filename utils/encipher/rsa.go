// 加解密相关工具 author chenweidong

package encipher

import (
	"crypto/x509"
	"crypto/rsa"
	"encoding/pem"
	"crypto/rand"
	"bytes"
	"io/ioutil"
)

type RsaKey struct {
	PublicKey  string
	PrivateKey string
}

func RsaEncryptByPuk(data []byte, puk string) ([]byte, error) {
	block, _ := pem.Decode([]byte(puk))
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := publicKey.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

func RsaDecryptByPuk(data []byte, puk string) ([]byte, error) {
	block, _ := pem.Decode([]byte(puk))
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	result := bytes.NewBuffer(nil)
	err = pubKeyIO(publicKey.(*rsa.PublicKey), bytes.NewReader(data), result, false)
	if err != nil {
		return []byte(""), err
	}
	return ioutil.ReadAll(result)
}

func RsaEncryptByPrk(data []byte, prk string) ([]byte, error) {
	block, _ := pem.Decode([]byte(prk))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	output := bytes.NewBuffer(nil)
	err = priKeyIO(privateKey, bytes.NewReader(data), output, true)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(output)
}

func RsaDecryptByPrk(data []byte, prk string) ([]byte, error) {
	block, _ := pem.Decode([]byte(prk))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
}

func GenRsaKey(bits int) (*RsaKey, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	prkStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "privateKey",
		Bytes: prkStream,
	}
	prkBytes := pem.EncodeToMemory(block)
	// 生成公钥
	publicKey := &privateKey.PublicKey
	pukStream, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	block = &pem.Block{
		Type:  "publicKey",
		Bytes: pukStream,
	}
	pukBytes := pem.EncodeToMemory(block)
	return &RsaKey{PrivateKey: string(prkBytes), PublicKey: string(pukBytes)}, nil
}
