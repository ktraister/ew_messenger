package main  
              
import (      
        "fmt" 
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
        qPubKeyData, _ := qPubKey.MarshalBinary()
        qPrivKeyData, _ := qPrivKey.MarshalBinary()

	fmt.Println("Test Private Key --> ", qPrivKey.String())
	fmt.Println("Test Public Key --> ", qPubKey.String())
	fmt.Println("Test Public Key Data --> ", qPubKeyData)
        fmt.Println("Test Private Key Data --> ", qPrivKeyData)

	qPubKeyString := "[]byte{"
	for _, v :=  range qPubKeyData {
	       qPubKeyString = qPubKeyString + fmt.Sprintf("%d", v) + "," 
        }
	qPubKeyString = qPubKeyString + "}"

	qPrivKeyString := "[]byte{"
	for _, v := range qPrivKeyData {
	       qPrivKeyString = qPrivKeyString + fmt.Sprintf("%d", v) + "," 
        }
	qPrivKeyString = qPrivKeyString + "}"

	fmt.Println()
	fmt.Println("Easy Copy Pasta")
	fmt.Println("PubKey --> ", qPubKeyString)
	fmt.Println("PrivKey --> ", qPrivKeyString)


	/*
	var privateKeyString, publicKeyString string

        //this code is how we intake a privkey from a string
	privateKeyString = qPrivKey.String()
	// Convert the string to a byte slice
	privateKeyBytes := []byte(privateKeyString)
	// Create a new private key from the byte slice
	privateKey := suite.Scalar().SetBytes(privateKeyBytes)

	publicKeyString = fmt.Sprintf("%s", qPubKey)
	// Convert the string to a byte slice
	publicKeyBytes := []byte(publicKeyString)
	// Create a new point on the curve
	publicKeyPoint := suite.Point()
	// Set the public key point from the byte slice
	if err := publicKeyPoint.UnmarshalBinary(publicKeyBytes); err != nil {
		fmt.Println("Error setting public key:", err)
		return
	}

	fmt.Println()
	fmt.Println("Private Key Set:", privateKey)
	*/
}
