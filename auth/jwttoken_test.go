package auth

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func Test_Verify(t *testing.T) {
	var verifier, err = NewRsaAuthVerifier("./pub.key")
	if err != nil {
		t.Fatal(err.Error())
	}

	token, e := verifier.Verify("123123")
	if e == nil {
		t.Fatal("verify fatal with invalid token. 1")
	}

	if token != nil && token.Valid {
		t.Fatal("verify fatal with invalid token. 2")
	}

	priBytes, e1 := ioutil.ReadFile("./pri.key")
	if e1 != nil {
		t.Fatal("private read fail")
	}

	priKey, e2 := jwt.ParseRSAPrivateKeyFromPEM(priBytes)
	if e2 != nil {
		t.Fatal("private key invalid.")
	}

	var token2str, e3 = jwt.NewWithClaims(jwt.SigningMethodRS256, &jwt.StandardClaims{Subject: "testsubject", IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Unix() + 3600}).SignedString(priKey)
	if e3 != nil {
		t.Fatal("new key signed failed!")
	}

	var token2, e4 = verifier.Verify(token2str)
	if e4 != nil {
		t.Fatal(e4.Error())
	}

	if standardc, ok := token2.Claims.(*jwt.StandardClaims); !ok {
		t.Fatal("type error")
	} else {
		if standardc.Subject != "testsubject" {
			t.Fatal("parse fail.")
		}
	}
}
