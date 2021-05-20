package auth

import (
	"crypto/rsa"
	"io/ioutil"

	"github.com/dgrijalva/jwt-go"
)

type RsaAuthVerifier struct {
	RSAPubKey *rsa.PublicKey
}

//使用RSA公匙文件初始化一个校验器
func NewRsaAuthVerifier(pubkeyfile string) (*RsaAuthVerifier, error) {
	pem, readerr := ioutil.ReadFile(pubkeyfile)
	if readerr != nil {
		return nil, readerr
	}

	pubkey, parserr := jwt.ParseRSAPublicKeyFromPEM(pem)
	if parserr != nil {
		return nil, parserr
	}

	var ret = new(RsaAuthVerifier)
	ret.RSAPubKey = pubkey
	return ret, nil
}

//验证token是否有效
func (this *RsaAuthVerifier) Verify(tokenstr string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenstr, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) { //使用公匙解密
		return this.RSAPubKey, nil
	})
}
