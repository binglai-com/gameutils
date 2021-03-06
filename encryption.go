package gameutils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

//sha1签名后进行base64编码
func Base64_Hamac_Sha1(val string, key string) string {
	return Base64String(Hmac_sha1(val, key))
}

//base64编码
func Base64String(str []byte) string {
	return base64.StdEncoding.EncodeToString(str)
}

//hamc sha1签名
func Hmac_sha1(val string, key string) []byte {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(val))
	return h.Sum(nil)
}

//md5签名
func Md5Str(val string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(val)))
}
