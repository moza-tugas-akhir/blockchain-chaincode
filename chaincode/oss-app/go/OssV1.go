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
		return fmt.Errorf("failed to read the data from world state: %v", err)
	}
	return nil
}

// CreateDoc creates a new document in the world state
func (nibsc *NIBIssuanceSmartContract) CreateDoc(ctx contractapi.TransactionContextInterface, userID string, docID string, docName string, docType string, timestamp time.Time, ipfsHash string) error {
	docJSON, err := ctx.GetStub().GetState(userID) //read prop from world state using id userID //checks if the property already exists
	if err := checkIfError(err); err != nil {
		return err
	}

	if docJSON != nil {
		return fmt.Errorf("the document with userID %s already exists", userID)
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

// helper function to reduce redundancy
func (nibsc *NIBIssuanceSmartContract) queryDoc(ctx contractapi.TransactionContextInterface, key string) (*Doc, error) {
	docJSON, err := ctx.GetStub().GetState(key) // read prop from world state using the key //checks if the property already exists
	if err := checkIfError(err); err != nil {
		return nil, err
	}

	if docJSON == nil {
		return nil, fmt.Errorf("the document with key %s does not exist", key)
	}

	var doc Doc
	err = json.Unmarshal(docJSON, &doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// QueryDocByUserId queries a document by userID
func (nibsc *NIBIssuanceSmartContract) QueryDocByUserId(ctx contractapi.TransactionContextInterface, userID string) (*Doc, error) {
	return nibsc.queryDoc(ctx, userID)
}

// QueryDocByDocId queries a document by docID
func (nibsc *NIBIssuanceSmartContract) QueryDocByDocId(ctx contractapi.TransactionContextInterface, docID string) (*Doc, error) {
	return nibsc.queryDoc(ctx, docID)
}

// QueryDocByName queries a document by docName
func (nibsc *NIBIssuanceSmartContract) QueryDocByName(ctx contractapi.TransactionContextInterface, docName string) (*Doc, error) {
	return nibsc.queryDoc(ctx, docName)
}

// QueryAllDocs queries all documents [not specified by user id]
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
