// encipher包测试 author chenweidong

package encipher

import (
	"fmt"
	"testing"
)

func TestRsa(t *testing.T) {
	rsaKey, _ := GenRsaKey(128)
	fmt.Println(rsaKey.PublicKey + "\n" + rsaKey.PrivateKey)

	text := []byte("test")
	data, _ := RsaEncryptByPuk(text, rsaKey.PublicKey)
	result, _ := RsaDecryptByPrk(data, rsaKey.PrivateKey)
	fmt.Println(string(result))

	text = []byte("test2")
	data, _ = RsaEncryptByPrk(text, rsaKey.PrivateKey)
	result, _ = RsaDecryptByPuk(data, rsaKey.PublicKey)
	fmt.Println(string(result))
}
