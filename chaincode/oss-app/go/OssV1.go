package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"math/rand"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// Defining smart contract
type NIBIssuanceSmartContract struct {
	contractapi.Contract
}

// property object, json based
type Doc struct {
	UserID    string    `json:"userid"`
	DocID     string    `json:"docid"`
	DocName   string    `json:"docname"`
	DocType   string    `json:"doctype"`
	Timestamp time.Time `json:"timestamp"`
	IPFSHash  string    `json:"ipfshash"`
}

type User struct {
	UserID string `json:"userid"`
	Email  string `json:"email"`
	Pwd    string `json:"pwd"` //omit when marshalling to JSON
}

func checkIfError(err error) error {
	if err != nil {
		return fmt.Errorf("failed to read the data from world state: %v", err)
	}
	return nil
}

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}

// Used to generate random user ID upon registering as a user
func generateUserID(ctx contractapi.TransactionContextInterface) (string, error) {
	for i := 0; i < 10; i++ {
		randomFloat := rand.Float64()

		// Convert to base 36 string (using digits and lowercase letters)
		randomString := strconv.FormatFloat(randomFloat, 'f', -1, 64)

		// Remove the "0." prefix
		randomString = randomString[2:]
		if len(randomString) > 7 {
			randomString = randomString[:7]
		}
		userID := randomString

		// Check if ID already exists
		userKey, err := ctx.GetStub().CreateCompositeKey("User", []string{userID})
		if err != nil {
			return "", err
		}

		existing, err := ctx.GetStub().GetState(userKey)
		if err != nil {
			return "", err
		}

		if existing == nil {
			return userID, nil // ID is unique, return it
		}
	}
	return "", fmt.Errorf("failed to generate a unique user ID after multiple attempts")
}

// func (nibsc *NIBIssuanceSmartContract) CreateUser(ctx contractapi.TransactionContextInterface, email string, pwd string) (string, error) {
func (nibsc *NIBIssuanceSmartContract) CreateUser(ctx contractapi.TransactionContextInterface, userID string, email string, pwd string) error {
	// Generate user ID upon registering as a user
	//userID, err := generateUserID(ctx)
	//if err != nil {
	//	return "", err
	//}

	// Create the composite key
	userKey, err := ctx.GetStub().CreateCompositeKey("User", []string{userID, email}) // []string = [userID, email]
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Retrieve the document
	userJSON, err := ctx.GetStub().GetState(userKey)
	// if err := checkIfError(err); err != nil {
	// 	return err
	// }
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	if userJSON != nil {
		return fmt.Errorf("the document with email %s already exists", email)
	}

	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to hash password: %v", err)
	// }

	user := User{
		UserID: userID,
		Email:  email,
		Pwd:    pwd,
		// Pwd:    string(hashedPassword),
	}

	// user := User{
	// 	UserID: userID,
	// 	Email:  email,
	// 	Pwd:    pwd,
	// }

	//json message has to be marshalled to the required format so can be sent to the blockchain
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Putting the document state to the ledger using the composite key
	err = ctx.GetStub().PutState(userKey, userBytes)
	if err != nil {
		return err
	}

	// Creating a secondary index for searching by document name
	emailKey, err := ctx.GetStub().CreateCompositeKey("Email", []string{email, userID})
	if err != nil {
		return fmt.Errorf("failed to create composite key for doc name: %v", err)
	}

	// return ctx.GetStub().PutState(emailKey, []byte(userKey))
	// Storing the document key in the secondary index
	err = ctx.GetStub().PutState(emailKey, []byte(userKey))
	if err != nil {
		return err
	}

	return nil
}

func (nibsc *NIBIssuanceSmartContract) QueryUserByEmail(ctx contractapi.TransactionContextInterface, email string) ([]*User, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Email", []string{email})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var users []*User
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		// Get the actual document key from the secondary index
		emailKey := queryResponse.Value

		// Retrieve the document using the email key
		userJSON, err := ctx.GetStub().GetState(string(emailKey))
		if err != nil {
			return nil, err
		}

		var user User // -> 0xAE492
		err = json.Unmarshal(userJSON, &user)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)

		// pointer/reference by address (e.g. *user) -> 0xAF928
		// reference by value (e.g. &user) -> { userID: "siapa", email: "email", ... }
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found with email %s", email)
	}

	return users, nil
}

func (nibsc *NIBIssuanceSmartContract) Login(ctx contractapi.TransactionContextInterface, email string, pwd string) (bool, error) {
	users, err := nibsc.QueryUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	if len(users) == 0 {
		return false, fmt.Errorf("no user found with email %s", email)
	}

	user := users[0]

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(pwd))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, fmt.Errorf("incorrect password")
		}
		return false, fmt.Errorf("error verifying password: %v", err)
	}

	return true, nil
}

// CreateDoc function to store documents with composite keys and secondary indexes
func (nibsc *NIBIssuanceSmartContract) CreateDoc(ctx contractapi.TransactionContextInterface, userID string, docID string, docName string, docType string, timestamp time.Time, ipfsHash string) error {
	// Create a composite key for the document using userID and docID
	docKey, err := ctx.GetStub().CreateCompositeKey("Doc", []string{userID, docID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Checking if the document already exists
	docJSON, err := ctx.GetStub().GetState(docKey)
	if err := checkIfError(err); err != nil {
		return err
	}

	if docJSON != nil {
		return fmt.Errorf("the document with userID %s and docID %s already exists", userID, docID)
	}

	doc := Doc{
		UserID:    userID,
		DocID:     docID,
		DocName:   docName,
		DocType:   docType,
		Timestamp: timestamp,
		IPFSHash:  ipfsHash,
	}

	//json message has to be marshalled to the required format so can be sent to the blockchain
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	// Putting the document state to the ledger using the composite key
	err = ctx.GetStub().PutState(docKey, docBytes)
	if err != nil {
		return err
	}

	// Creating a secondary index for searching by document name
	docNameKey, err := ctx.GetStub().CreateCompositeKey("DocName", []string{docName, userID, docID})
	if err != nil {
		return fmt.Errorf("failed to create composite key for doc name: %v", err)
	}

	// Storing the document key in the secondary index
	return ctx.GetStub().PutState(docNameKey, []byte(docKey))
}

// QueryDocByUserId queries a document by userID
func (nibsc *NIBIssuanceSmartContract) QueryDocByUserId(ctx contractapi.TransactionContextInterface, userID string) ([]*Doc, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Doc", []string{userID})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var docs []*Doc
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		var doc Doc
		err = json.Unmarshal(queryResponse.Value, &doc)
		if err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents found for user ID %s", userID)
	}

	return docs, nil
}

// QueryDocByDocId function to retrieve a document by docID
func (nibsc *NIBIssuanceSmartContract) QueryDocByDocId(ctx contractapi.TransactionContextInterface, userID string, docID string) (*Doc, error) {
	// Create the composite key
	docKey, err := ctx.GetStub().CreateCompositeKey("Doc", []string{userID, docID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Retrieve the document
	docJSON, err := ctx.GetStub().GetState(docKey)
	if err := checkIfError(err); err != nil {
		return nil, err
	}

	if docJSON == nil {
		return nil, fmt.Errorf("the document with userID %s and docID %s does not exist", userID, docID)
	}

	var doc Doc
	err = json.Unmarshal(docJSON, &doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (nibsc *NIBIssuanceSmartContract) QueryDocByName(ctx contractapi.TransactionContextInterface, docName string) ([]*Doc, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey("DocName", []string{docName})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var docs []*Doc
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		// Get the actual document key from the secondary index
		docKey := queryResponse.Value

		// Retrieve the document using the document key
		docJSON, err := ctx.GetStub().GetState(string(docKey))
		if err != nil {
			return nil, err
		}

		var doc Doc
		err = json.Unmarshal(docJSON, &doc)
		if err != nil {
			return nil, err
		}

		docs = append(docs, &doc)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents found with document name %s", docName)
	}

	return docs, nil
}

func main() {
	fmt.Println("Starting NIB Issuance Smart Contract")

	// creates an instance + initiating new chaincode
	chaincode, err := contractapi.NewChaincode(new(NIBIssuanceSmartContract))

	if err != nil {
		fmt.Printf("Error creating doc chaincode: %s", err)
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting doc chaincode: %s", err)
	}
}
