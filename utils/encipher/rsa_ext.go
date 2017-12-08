// author gorsa github copy

package encipher

import (
	"math/big"
	"errors"
	"crypto/rsa"
	"io"
	"crypto/rand"
)

var (
	ErrDataToLarge = errors.New("message too long for RSA public key size")
	ErrDataLen     = errors.New("data length error")
	ErrDataBroken  = errors.New("data broken, first byte is not zero")
	ErrKeyPair     = errors.New("data is not encrypted by the private key")
	ErrDecryption  = errors.New("decryption error")
)

// 公钥加密或解密Reader
func pubKeyIO(pub *rsa.PublicKey, in io.Reader, out io.Writer, isEncrypt bool) (err error) {
	k := (pub.N.BitLen() + 7) / 8
	if isEncrypt {
		k = k - 11
	}
	buf := make([]byte, k)
	var b []byte
	size := 0
	for {
		size, err = in.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if size < k {
			b = buf[:size]
		} else {
			b = buf
		}
		if isEncrypt {
			b, err = rsa.EncryptPKCS1v15(rand.Reader, pub, b)
		} else {
			b, err = pubKeyDecrypt(pub, b)
		}
		if err != nil {
			return err
		}
		if _, err = out.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// 私钥加密或解密Reader
func priKeyIO(pri *rsa.PrivateKey, r io.Reader, w io.Writer, isEncrypt bool) (err error) {
	k := (pri.N.BitLen() + 7) / 8
	if isEncrypt {
		k = k - 11
	}
	buf := make([]byte, k)
	var b []byte
	size := 0
	for {
		size, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if size < k {
			b = buf[:size]
		} else {
			b = buf
		}
		if isEncrypt {
			b, err = priKeyEncrypt(rand.Reader, pri, b)
		} else {
			b, err = rsa.DecryptPKCS1v15(rand.Reader, pri, b)
		}
		if err != nil {
			return err
		}
		if _, err = w.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// 公钥解密
func pubKeyDecrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	k := (pub.N.BitLen() + 7) / 8
	if k != len(data) {
		return nil, ErrDataLen
	}
	m := new(big.Int).SetBytes(data)
	if m.Cmp(pub.N) > 0 {
		return nil, ErrDataToLarge
	}
	m.Exp(m, big.NewInt(int64(pub.E)), pub.N)
	d := leftPad(m.Bytes(), k)
	if d[0] != 0 {
		return nil, ErrDataBroken
	}
	if d[1] != 0 && d[1] != 1 {
		return nil, ErrKeyPair
	}
	var i = 2
	for ; i < len(d); i++ {
		if d[i] == 0 {
			break
		}
	}
	i++
	if i == len(d) {
		return nil, nil
	}
	return d[i:], nil
}

// 私钥加密
func priKeyEncrypt(rand io.Reader, prk *rsa.PrivateKey, hashed []byte) ([]byte, error) {
	tLen := len(hashed)
	k := (prk.N.BitLen() + 7) / 8
	if k < tLen+11 {
		return nil, ErrDataLen
	}
	em := make([]byte, k)
	em[1] = 1
	for i := 2; i < k-tLen-1; i++ {
		em[i] = 0xff
	}
	copy(em[k-tLen:k], hashed)
	m := new(big.Int).SetBytes(em)
	c, err := decrypt(rand, prk, m)
	if err != nil {
		return nil, err
	}
	copyWithLeftPad(em, c.Bytes())
	return em, nil
}

// 从crypto/rsa复制
var bigZero = big.NewInt(0)
var bigOne = big.NewInt(1)

// 从crypto/rsa复制
func decrypt(random io.Reader, prk *rsa.PrivateKey, c *big.Int) (m *big.Int, err error) {
	if c.Cmp(prk.N) > 0 {
		err = ErrDecryption
		return
	}
	var ir *big.Int
	if random != nil {
		var r *big.Int

		for {
			r, err = rand.Int(random, prk.N)
			if err != nil {
				return
			}
			if r.Cmp(bigZero) == 0 {
				r = bigOne
			}
			var ok bool
			ir, ok = modInverse(r, prk.N)
			if ok {
				break
			}
		}
		bigE := big.NewInt(int64(prk.E))
		cCopy := new(big.Int).Set(c)
		cCopy.Mul(cCopy, new(big.Int).Exp(r, bigE, prk.N))
		cCopy.Mod(cCopy, prk.N)
		c = cCopy
	}
	if prk.Precomputed.Dp == nil {
		m = new(big.Int).Exp(c, prk.D, prk.N)
	} else {
		m = new(big.Int).Exp(c, prk.Precomputed.Dp, prk.Primes[0])
		m2 := new(big.Int).Exp(c, prk.Precomputed.Dq, prk.Primes[1])
		m.Sub(m, m2)
		if m.Sign() < 0 {
			m.Add(m, prk.Primes[0])
		}
		m.Mul(m, prk.Precomputed.Qinv)
		m.Mod(m, prk.Primes[0])
		m.Mul(m, prk.Primes[1])
		m.Add(m, m2)

		for i, values := range prk.Precomputed.CRTValues {
			prime := prk.Primes[2+i]
			m2.Exp(c, values.Exp, prime)
			m2.Sub(m2, m)
			m2.Mul(m2, values.Coeff)
			m2.Mod(m2, prime)
			if m2.Sign() < 0 {
				m2.Add(m2, prime)
			}
			m2.Mul(m2, values.R)
			m.Add(m, m2)
		}
	}
	if ir != nil {
		m.Mul(m, ir)
		m.Mod(m, prk.N)
	}
	return
}

// 从crypto/rsa复制
func copyWithLeftPad(dest, src []byte) {
	numPaddingBytes := len(dest) - len(src)
	for i := 0; i < numPaddingBytes; i++ {
		dest[i] = 0
	}
	copy(dest[numPaddingBytes:], src)
}

// 从crypto/rsa复制
func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out)-n:], input)
	return
}

// 从crypto/rsa复制
func modInverse(a, n *big.Int) (ia *big.Int, ok bool) {
	g := new(big.Int)
	x := new(big.Int)
	y := new(big.Int)
	g.GCD(x, y, a, n)
	if g.Cmp(bigOne) != 0 {
		return
	}
	if x.Cmp(bigOne) < 0 {
		x.Add(x, n)
	}
	return x, true
}
