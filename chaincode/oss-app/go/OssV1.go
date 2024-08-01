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

// QueryDocByDocId queries a document by docID
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

// QueryAllDocs queries all documents [not specified by user id]
// func (nibsc *NIBIssuanceSmartContract) QueryAllDocs(ctx contractapi.TransactionContextInterface) ([]*Doc, error) {
// 	// Getting all the values stored in the world state
// 	docIterator, err := ctx.GetStub().GetStateByRange("", "")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer docIterator.Close()

// 	var docs []*Doc

// 	// For each loop
// 	for docIterator.HasNext() {
// 		response, err := docIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}

// 		var doc Doc
// 		err = json.Unmarshal(response.Value, &doc)
// 		if err != nil {
// 			return nil, err
// 		}
// 		docs = append(docs, &doc)
// 	}

// 	if len(docs) == 0 {
// 		fmt.Println("No documents found")
// 	} else {
// 		fmt.Printf("Found %d documents\n", len(docs))
// 	}

// 	return docs, nil
// }

// QueryAllDocsJSON returns a readable results of the QueryAllDocs function
// func (nibsc *NIBIssuanceSmartContract) QueryAllDocsJSON(ctx contractapi.TransactionContextInterface) (string, error) {
// 	docs, err := nibsc.QueryAllDocs(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	if docs == nil {
// 		fmt.Println("QueryAllDocs returned nil")
// 		return "[]", nil
// 	}

// 	docsJSON, err := json.Marshal(docs)
// 	if err != nil {
// 		return "", err
// 	}

// 	fmt.Printf("Returning JSON: %s\n", string(docsJSON))
// 	return string(docsJSON), nil
// }

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
