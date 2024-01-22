package main  
              
import (      
        "fmt" 
	"encoding/base64"
        "go.dedis.ch/kyber/v3/group/edwards25519"
        "go.dedis.ch/kyber/v3/util/random"
)   

func main() {
        //suite used by the whole shebang
        suite := edwards25519.NewBlakeSHA256Ed25519()
	
	//this code used to set random priv/pub key, and display 
	//base64 encoded pubkey which we use in the 1 time KEX
        qPrivKey := suite.Scalar().Pick(random.New())
        qPubKey := suite.Point().Mul(qPrivKey, nil)
        qPubKeyData, err := qPubKey.MarshalBinary()
	qPubKeyDataBase64 := base64.StdEncoding.EncodeToString(qPubKeyData)
        if err != nil {
		fmt.Println("ERROR: ", err)
                return
        }

	fmt.Println("Test Private Key --> ", qPrivKey)
	fmt.Println("Test Public Key --> ", qPubKey)
	fmt.Println("Test Public Key Data --> ", qPubKeyData)
	fmt.Println("Test Public Key Data Base64 --> ", qPubKeyDataBase64)

        //this code is how we intake a privkey from a string
	var privateKeyString string
	privateKeyString = qPrivKey.String()
	// Convert the string to a byte slice
	privateKeyBytes := []byte(privateKeyString)
	// Create a new private key from the byte slice
	privateKey := suite.Scalar().SetBytes(privateKeyBytes)
	fmt.Println()
	fmt.Println("Private Key Set:", privateKey)
}
