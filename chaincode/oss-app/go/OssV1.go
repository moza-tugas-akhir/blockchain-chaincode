package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
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

func checkIfError(err error) error {
	if err != nil {
		return fmt.Errorf("Failed to read the data from world state: %v", err)
	}
	return nil
}

// sepertinya harus ada nested, dari user ID dulu terus Doc ID
// need to be discussed
// CreateDoc creates a new document in the world state
func (nibsc *NIBIssuanceSmartContract) CreateDoc(ctx contractapi.TransactionContextInterface, userID string, docID string, docName string, docType string, timestamp time.Time, ipfsHash string) error {
	docJSON, err := ctx.GetStub().GetState(userID) //read prop from world state using id userID //checks if the property already exists
	if err := checkIfError(err); err != nil {
		return err
	}

	if docJSON != nil {
		return fmt.Errorf("The document with userID %s already exists", userID)
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

	return ctx.GetStub().PutState(userID, docBytes)
}

// Query by ID [kalau UserID berarti semua dokumen punya user? terus kalo mau spesifik ke ID nya gimana?]
// Ini sama aja kayak query all?
// Ini harusnya query all docs, terus diubah jadi docbyId
// QueryDocByUserId queries a document by userID
func (nibsc *NIBIssuanceSmartContract) QueryDocByUserId(ctx contractapi.TransactionContextInterface, userID string) (*Doc, error) {
	docJSON, err := ctx.GetStub().GetState(userID) //read prop from world state using id //checks if the property already exists
	if err := checkIfError(err); err != nil {
		return nil, err
	}

	if docJSON == nil {
		return nil, fmt.Errorf("The document with userID %s does not exist", userID)
	}

	var doc Doc
	err = json.Unmarshal(docJSON, &doc)

	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// QueryAllDocs queries all documents
func (nibsc *NIBIssuanceSmartContract) QueryAllDocs(ctx contractapi.TransactionContextInterface) ([]*Doc, error) {
	docIterator, err := ctx.GetStub().GetStateByRange("", "") //getting all the values stored in the world state
	// cari getstate yang bisa get range by user id
	if err := checkIfError(err); err != nil {
		return nil, err
	}
	defer docIterator.Close()

	var docs []*Doc

	//for each loop
	for docIterator.HasNext() {
		response, err := docIterator.Next()
		if err != nil {
			return nil, err
		}

		var doc Doc
		err = json.Unmarshal(response.Value, &doc)
		if err != nil {
			return nil, err
		}
		docs = append(docs, &doc)
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
