package main

import (
	"encoding/json"
	"fmt"
	"time"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract implements a simple chaincode to manage an asset
type SmartContract struct {
  contractapi.Contract
}
//Asset 구조체->  DISPLAY – ID, DISID(디스플레이일련번호), RID, STATUS( requested, supplied, repair_requested, inrepairing)
type Display struct {
	ID	string
	RID	string
	Owner string
	Status  string // requested, supplied...
}
//history 결과저장구조체
type HistoryQueryResult struct {
	Record	*Display	  `json:"record"`
	TxId   string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete bool       `json:"isDelete"`
}
// Get 메서드 구현
func (s *SmartContract) GetDisplay(ctx contractapi.TransactionContextInterface, key string) (*Display, error) {
	DisplayAsBytes, err := ctx.GetStub().GetState(key)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if DisplayAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", key)
	}

	lcd := new(Display)
	_ = json.Unmarshal(DisplayAsBytes, lcd)

	return lcd, nil
}
// Set 메서드 구현
func (s *SmartContract) Apply(ctx contractapi.TransactionContextInterface, displayId string, requester string) error {
	//display  구조체생성
	lcd := Display{
	  ID :displayId,
	  RID :requester,
	  Status :"requested",
	}
	// marshal
	DisplayAsBytes, _ := json.Marshal(lcd)

	// putstate
	return ctx.GetStub().PutState(displayId, DisplayAsBytes)
}
func (s *SmartContract) Supply(ctx contractapi.TransactionContextInterface, displayNumber string, newOwner string) error {
	display, err := s.GetDisplay(ctx, displayNumber)

	if err != nil {
		return err
	}
	// RID == newOwner 검증
	display.Owner = newOwner

	displayAsBytes, _ := json.Marshal(display)

	return ctx.GetStub().PutState(displayNumber, displayAsBytes)
}

func (t *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", assetID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var car Display
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &car)
			if err != nil {
				return nil, err
			}
		} else {
			car = Display{
				ID : assetID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:	   response.TxId,
			Timestamp: timestamp,
			Record:    &car,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}
	
	return records, nil
}
/*
//History 메서드구현
func (t *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResult, error) {
	log.Printf("GetHistory: ID %v", key)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{
			  Key: key,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

//Transfer 메서드구현
func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, from string, to string, amount float64) error {
	fromAsset, err := s.Get(ctx, from)
	if err != nil {
		return err
	}
	toAsset, err := s.Get(ctx, to)
	if err != nil {
		return err
	}

	//검증
	if fromAsset.Value - amount < 0 {
	  return fmt.Errorf("not enough balance in from account: %s", from)
	}

	//전송
	  fromAsset.Value = fromAsset.Value - amount
	  toAsset.Value = toAsset.Value + amount

	// 
	fromAsBytes, _ := json.Marshal(fromAsset)
	ctx.GetStub().PutState(from, fromAsBytes)
	toAsbytes, _ := json.Marshal(toAsset)
	ctx.GetStub().PutState(to, toAsbytes)

	return nil
}
*/
// main function starts up the chaincode in the container during instantiate
func main() {
	
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error creats simpleasset chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting simpleasset chaincode: %s", err.Error())
	}
}
