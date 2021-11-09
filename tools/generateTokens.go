package tools

import (
	"bufio"
	"encoding/base64"
	"log"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	Path = "init/apikeys.txt"
)

var (
	Tokens = SplitLines(ReadFile(Path))
)

func GenerateTokenCLI(username, password string) string {
	val := VerifyPassword(password)
	if !val {
		log.Fatal("The password does not meet the password policy requirements. Check the minimum password length(7), password complexity(at least 1 number, upper letter and symbol)")

	}
	b64Password := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	// fmt.Println(string(b64Password))

	hash, err := bcrypt.GenerateFromPassword([]byte(b64Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("Hash to store:", string(hash))
	WriteFile(Path, string(hash))
	return string(hash)
}

func ValidateToken(passwordStr string) bool {
	var (
		passwordByte = []byte(passwordStr)
		result       = false
	)

	for _, token := range Tokens {
		if token == "\n" {
			continue
		}
		val := bcrypt.CompareHashAndPassword([]byte(token), passwordByte)
		// fmt.Println([]byte(token))
		if val == nil {
			result = true
		}
	}
	return result
}

func VerifyPassword(s string) bool {
	var (
		letters     = 0
		number      = false
		upper       = false
		special     = false
		sevenOrMore = false
	)
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}

	sevenOrMore = letters >= 7
	// fmt.Printf("%t - number, %t - upper, %t - special,  %t - sevenorMore\n", number, upper, special, sevenOrMore)
	return number && upper && special && sevenOrMore
}

func SplitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}
