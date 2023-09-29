package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

/*
         // character to ASCII
         char := 'a' // rune, not string
         ascii := int(char)
         fmt.Println(string(char), " : ", ascii)

         // ASCII to character

         asciiNum := 65  // Uppercase A
         character := string(asciiNum)
         fmt.Println(asciiNum, " : ", character)

	 //outputs
	 a:97
	 65:A
*/

//both these functions need to pass around strings for ease of use

func toString(INPUT []string) string {
	st := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(INPUT)), ", "), "[]")
	return st
}

func fromString(INPUT string) []string {
	trim := strings.ReplaceAll(INPUT, "\n", "")
	s := strings.Split(strings.ReplaceAll(trim, " ", ""), ",")
	mySlice := []string{}
	for _, val := range s {
		mySlice = append(mySlice, val)
	}
	return mySlice
}

func pack_message(INPUT string) []rune {
	desiredLength := 4096
	fillCharacters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@$%^&*()-_=+[]{}|;:',.<>?~"

	rand.Seed(time.Now().UnixNano())

	result := INPUT + "###"

	for len(result) < desiredLength {
		randomIndex := rand.Intn(len(fillCharacters))
		result += string(fillCharacters[randomIndex])
	}

	// Trim the string if it's longer than desiredLength
	if len(result) > desiredLength {
		result = result[:desiredLength]
	}

	return []rune(result)
}

func unpack_message(INPUT string) string {
	result := strings.Split(INPUT, "###")
	return result[0]
}

func transform_pad(PAD string, PRIVKEY string) ([]string, error) {
	privKeyInt, _ := big.NewInt(0).SetString(PRIVKEY, 0)
	tmpBigInt := big.NewInt(0)
	otp := strings.Split(PAD, " ")
	final := []string{}

	for _, s := range otp {
		tmpBigInt, _ = tmpBigInt.SetString(s, 0)
		tmpBigInt.Mod(privKeyInt, tmpBigInt)
		final = append(final, tmpBigInt.String())
	}

	return final, nil
}

func pad_encrypt(MSG string, PAD string, PRIVKEY string) string {
	//implement pack_message here
	chars := pack_message(MSG)
	asc_chars := make([]int, 0)
	enc_msg := make([]string, 0)
	tmpBigInt := big.NewInt(1)

	//change chars to ascii_chars
	for i := 0; i < len(chars); i++ {
		asc_chars = append(asc_chars, int(chars[i]))
	}

	asc_pad, _ := transform_pad(PAD, PRIVKEY)

	//encrypt the message
	for i := 0; i < len(asc_chars); i++ {
		tmpBigInt.SetString(asc_pad[i], 0)
		//sticking with subtract logic for now
		tmpBigInt.Add(tmpBigInt, big.NewInt(int64(asc_chars[i])))
		enc_msg = append(enc_msg, tmpBigInt.String())
	}

	return toString(enc_msg)
}

func pad_decrypt(INPUT_MSG string, PAD string, PRIVKEY string) string {
	dec_msg := make([]string, 0)
	tmpBigInt := big.NewInt(1)
	bigIntToo := big.NewInt(1)

	//convert ENC_MSG string to []int
	enc_msg := fromString(INPUT_MSG)

	asc_pad, _ := transform_pad(PAD, PRIVKEY)

	//decrypt message
	for i := 0; i < len(enc_msg); i++ {
		tmpBigInt.SetString(enc_msg[i], 0)
		bigIntToo.SetString(asc_pad[i], 0)
		tmpBigInt.Sub(tmpBigInt, bigIntToo)
		dec_msg = append(dec_msg, string(rune(int(tmpBigInt.Int64()))))
	}

	//change ascii_chars to chars and stringify
	//easy way to do this in go
	//https://stackoverflow.com/questions/40310333/how-to-append-a-character-to-a-string-in-golang
	var sb strings.Builder
	for i := 0; i < len(dec_msg); i++ {
		sb.WriteString(dec_msg[i])
	}
	dec_string := sb.String()

	return unpack_message(dec_string)
}
