package main

import (
	"encoding/json"
)

type MessageType int
type WsStatus int

const (
	WS_STATUS_OK WsStatus = iota
	WS_STATUS_BAD_REQUEST
)
const (
	MESSAGE_TYPE_REGISTER_MOVE_2D MessageType = iota
	MESSAGE_TYPE_JOIN_MATCH_2D
	MESSAGE_TYPE_CREATE_MATCH_2D
	MESSAGE_TYPE_ABANDON_MATCH_2D
	MESSAGE_TYPE_ASK_DRAW_2D
)

type MatchOpts struct {
	W int `json:"w"`
	H int `json:"h"`
}

type WsRequest struct {
	Type MessageType `json:"type"`
	Body any         `json:"body"`
	ID   string      `json:"id"`
}

func createJson() []byte {
	type Puchi struct {
		W int `json:"w"`
		H int `json:"h"`
	}
	type JsType struct {
		Type MessageType `json:"type"`
		Body Puchi       `json:"body"`
		ID   string      `json:"id"`
	}
	someJSObject := JsType{
		Type: 1,
		Body: Puchi{
			W: 7,
			H: 6,
		},
		ID: "32",
	}
	ObjectJSON, err := json.Marshal(someJSObject)
	if err != nil {
		panic(err)
	}
	return ObjectJSON
}

func main() {
	/* jsJson := createJson()
	var Req WsRequest
	if err := json.Unmarshal(jsJson, &Req); err != nil {
		panic(err)
	}
	fmt.Printf("resp is: \n%+v\n", Req) */

	type Key2Type struct {
		Subkey1 int `json:"subKey1"`
	}
	type CorrespondingType struct {
		Key1 string   `json:"key1"`
		Key2 Key2Type `json:"key2"`
	}
	mapi := map[string]any{
		"key1": "value1",
		"key2": map[string]any{
			"subKey1": 1,
		},
	}

	var ct CorrespondingType
	bs, _ := json.Marshal(mapi)
	json.Unmarshal(bs, &ct)

}
