package main

import (
	"encoding/json" //io/ou for the chaincode is json based
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//high-level API, always to be imported
)

// Defining smart contract
type PropertyTransferSmartContract struct {
	contractapi.Contract //interface
}

// property object, json based, use backtick
type Property struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Area      int    `json:"area"`
	OwnerName string `json:"ownerName"`
	Value     int    `json:"value"`
}

// adding property to ledger
func (pc *PropertyTransferSmartContract /*instance of smart contract, pc is short name*/) AddProperty(ctx contractapi.TransactionContextInterface, /*gives the required functions for worldstate (read/write from blockchain)  */
	id string, name string, area int, ownerName string, value int) error /*return type is error*/ {
	propertyJSON, err := ctx.GetStub().GetState(id) //read prop from world state using id //checks if the property already exists
	if err != nil {
		return fmt.Errorf("Failed to read the data from world state: %v", err)
	}

	if propertyJSON != nil {
		return fmt.Errorf("The property %s already exists", id)
	}
	//obj
	prop := Property{
		ID:        id,
		Name:      name,
		Area:      area,
		OwnerName: ownerName,
		Value:     value,
	}

	propertyBytes, err := json.Marshal(prop) //json message has to be marshalled to the required format so can be sent to the blockchain
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, propertyBytes)
	//save to the worldstate
}

// returning all existing properties
func (pc *PropertyTransferSmartContract /*instance of smart contract, pc is short name*/) QueryAllProperties(ctx contractapi.TransactionContextInterface /*gives the required functions for worldstate (read/write from blockchain)  */) ([]*Property, error) /* multiple return type*/ {
	propertyIterator, err := ctx.GetStub().GetStateByRange("", "") //getting all the values stored in the world state
	if err != nil {
		return nil, fmt.Errorf("Failed to read the data from world state: %v", err)
	}
	defer propertyIterator.Close()

	var properties []*Property

	//for each loop
	for propertyIterator.HasNext() {
		propertyResponse, err := propertyIterator.Next()
		if err != nil {
			return nil, err
		}

		var property *Property
		err = json.Unmarshal(propertyResponse.Value, &property)
		if err != nil {
			return nil, err
		}
		properties = append(properties, property)
	}

	return properties, nil
}

// Query by ID
func (pc *PropertyTransferSmartContract /*instance of smart contract, pc is short name*/) QueryPropertyByID(ctx contractapi.TransactionContextInterface /*gives the required functions for worldstate (read/write from blockchain) */, id string) (*Property, error) /* multiple return type*/ {
	propertyJSON, err := ctx.GetStub().GetState(id) //read prop from world state using id //checks if the property already exists
	if err != nil {
		return nil, fmt.Errorf("Failed to read the data from world state: %v", err)
	}

	if propertyJSON == nil {
		return nil, fmt.Errorf("The property %s does not exist", id)
	}

	var property *Property
	err = json.Unmarshal(propertyJSON, &property)

	if err != nil {
		return nil, err
	}
	return property, nil
}

// transfer ownership property
func (pc *PropertyTransferSmartContract /*instance of smart contract, pc is short name*/) TransferProperty(ctx contractapi.TransactionContextInterface, /*gives the required functions for worldstate (read/write from blockchain)  */
	id string, newOwner string) error /*return type is error*/ {
	property, err := pc.QueryPropertyByID(ctx, id)
	if err != nil {
		return err
	}

	property.OwnerName = newOwner
	propertyJSON, err := json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, propertyJSON)
	//save to the worldstate
}

// changing the property value
func (pc *PropertyTransferSmartContract /*instance of smart contract, pc is short name*/) ChangePropertyValue(ctx contractapi.TransactionContextInterface, /*gives the required functions for worldstate (read/write from blockchain)  */
	id string, newValue int) error /*return type is error*/ {
	property, err := pc.QueryPropertyByID(ctx, id)
	if err != nil {
		return err
	}

	property.Value = newValue
	propertyJSON, err := json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, propertyJSON)
	//save to the worldstate
}

func main() {
	//standard syntax
	fmt.Println("Starting Property Transfer Smart Contract")
	//creating an instance of smartcontract
	propTransferSmartContract := new(PropertyTransferSmartContract)

	//initiating new chaincode
	cc, err := contractapi.NewChaincode(propTransferSmartContract)

	//checking if chaincode is error
	if err != nil {
		fmt.Printf("Error creating new chaincode: %v\n", err)
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
		panic(err.Error())
	}

	fmt.Println("Chaincode started successfully")
}
