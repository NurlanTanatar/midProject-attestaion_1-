package tools

import (
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"midProject/internal/models"
	"net/url"

	dgoogauth "github.com/dgryski/dgoogauth"
	qr "rsc.io/qr"
)

var (
	qrFilename = "./tmp/qr.png"
)

type BasicInfo struct {
	Email string
	Name  string
}

func (info BasicInfo) GenOTPCLI() {
	secret := []byte(info.Name)

	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	URL, err := url.Parse("otpauth://totp")
	if err != nil {
		panic(err)
	}

	URL.Path += "/" + url.PathEscape(info.Name) + ":" + url.PathEscape(info.Email)

	params := url.Values{}
	params.Add("secret", secretBase32)
	params.Add("issuer", info.Name)

	URL.RawQuery = params.Encode()
	fmt.Printf("URL is %s\n", URL.String())

	code, err := qr.Encode(URL.String(), qr.Q)
	if err != nil {
		panic(err)
	}
	b := code.PNG()
	err = ioutil.WriteFile(qrFilename, b, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Printf("QR code is in %s. Please scan it into Google Authenticator app.\n", qrFilename)

	otpc := &dgoogauth.OTPConfig{
		Secret:      secretBase32,
		WindowSize:  3,
		HotpCounter: 0,
	}

	for {
		var token string
		fmt.Printf("Please enter the token value (or q or quit to quit): ")
		fmt.Scanln(&token)

		if token == "q" || token == "quit" {
			break
		}

		val, err := otpc.Authenticate(token)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !val {
			fmt.Println("Sorry, Not Authenticated")
			continue
		}

		fmt.Println("Authenticated!")
	}
}

func GenOTPREST(user *models.User) string {
	secret := []byte(user.Name)
	uniqueQRlink := "./tmp/qr" + "_" + fmt.Sprint(user.ID) + ".png"
	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	URL, err := url.Parse("otpauth://totp")
	if err != nil {
		panic(err)
	}

	URL.Path += "/" + url.PathEscape(user.Name) + ":" + url.PathEscape(user.Email)

	params := url.Values{}
	params.Add("secret", secretBase32)
	params.Add("issuer", user.Name)

	URL.RawQuery = params.Encode()
	fmt.Printf("URL is %s\n", URL.String())

	code, err := qr.Encode(URL.String(), qr.Q)
	if err != nil {
		panic(err)
	}
	b := code.PNG()
	err = ioutil.WriteFile(uniqueQRlink, b, 0600)
	if err != nil {
		panic(err)
	}
	return uniqueQRlink
}

func GivePerm(user *models.User, token string) string {
	secret := []byte(user.Name)
	secretBase32 := base32.StdEncoding.EncodeToString(secret)
	otpc := &dgoogauth.OTPConfig{
		Secret:      secretBase32,
		WindowSize:  3,
		HotpCounter: 0,
	}

	val, err := otpc.Authenticate(token)
	if err != nil {
		fmt.Printf("Unknown err: %v", err)
		return "error"
	}

	if !val {
		return "Sorry, Not Authenticated"
	}

	return "Authenticated!"

}
