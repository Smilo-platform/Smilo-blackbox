// Copyright 2019 The Smilo-blackbox Authors
// This file is part of the Smilo-blackbox library.
//
// The Smilo-blackbox library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Smilo-blackbox library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Smilo-blackbox library. If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"net/http"
	"testing"

	"gopkg.in/urfave/cli.v1"

	"Smilo-blackbox/src/server/api"

	"encoding/base64"
	"encoding/json"

	"github.com/stretchr/testify/require"

	"os"
	"time"

	"Smilo-blackbox/src/crypt"
	"Smilo-blackbox/src/server/config"
	"Smilo-blackbox/src/server/syncpeer"
	"Smilo-blackbox/src/utils"
)

func TestMain(m *testing.M) {
	app := cli.NewApp()
	config.Init(app)
	err := config.LoadConfig("./server_test.conf")
	if err != nil {
		panic("Could not open config for server_hack_test")
	}

	go StartServer()

	config.WorkDir.Value = ""

	time.Sleep(2000000000)
	retcode := m.Run()
	os.Exit(retcode)
}

func TestGetPublicKeysFromOtherNode(t *testing.T) {
	keys, _, err := syncpeer.GetPublicKeysFromOtherNode("http://localhost:"+config.Port.Value, crypt.GetPublicKeys()[0])
	require.Nil(t, err, err)
	require.Equal(t, len(keys), 1)
	require.Equal(t, keys[0], crypt.GetPublicKeys()[0])
}

func TestUnixSend(t *testing.T) {
	to := make([]string, 1)
	to[0] = "OeVDzTdR95fhLKIgpBLxqdDNXYzgozgi7dnnS125A3w="
	sendRequest := api.SendRequest{Payload: base64.StdEncoding.EncodeToString([]byte("1234567890abcdefghijklmnopqrs")), From: "MD3fapkkHUn86h/W7AUhiD4NiDFkuIxtuRr0Nge27Bk=", To: to}
	req, err := json.Marshal(sendRequest)
	if err != nil {
		t.Fail()
	}
	response := DoUnixPostJSONRequest(t, "/send", string(req))
	var sendResponse api.KeyJSON
	err = json.Unmarshal([]byte(response), &sendResponse)
	require.NoError(t, err)

	receiveRequest := api.ReceiveRequest{Key: sendResponse.Key, To: sendRequest.To[0]}
	req2, err2 := json.Marshal(receiveRequest)
	require.Empty(t, err2)

	t.Log("Send Response: " + sendResponse.Key)

	response = DoUnixGetJSONRequest(t, "/receive", string(req2))
	var receiveResponse api.ReceiveResponse
	err = json.Unmarshal([]byte(response), &receiveResponse)
	require.NoError(t, err)

	t.Log("Receive Response: " + receiveResponse.Payload)
	require.Equal(t, sendRequest.Payload, receiveResponse.Payload)
}

func TestUnixSendRawTransactionGet(t *testing.T) {
	to := make([]string, 1)
	to[0] = "OeVDzTdR95fhLKIgpBLxqdDNXYzgozgi7dnnS125A3w="
	payload := "1234567890abcdefghijklmnopqrs"
	encPayload := base64.StdEncoding.EncodeToString([]byte(payload))
	from := "MD3fapkkHUn86h/W7AUhiD4NiDFkuIxtuRr0Nge27Bk="
	response := DoUnixPostRequest(t, "/sendraw", []byte(encPayload), http.Header{utils.HeaderFrom: []string{from}, utils.HeaderTo: to})

	key, err := base64.StdEncoding.DecodeString(response)
	if err != nil {
		t.Fail()
	}
	urlEncodedKey := base64.URLEncoding.EncodeToString(key)
	t.Log("Send Response: " + response)
	toBytes, err := base64.StdEncoding.DecodeString(to[0])
	if err != nil {
		t.Fail()
	}
	urlEncodedTo := base64.URLEncoding.EncodeToString(toBytes)
	response = DoUnixRequest(t, "/transaction/"+urlEncodedKey+"?to="+urlEncodedTo)
	var receiveResponse api.ReceiveResponse
	err = json.Unmarshal([]byte(response), &receiveResponse)
	require.NoError(t, err)

	retorno, _ := base64.StdEncoding.DecodeString(receiveResponse.Payload)
	t.Log("Receive Response: " + receiveResponse.Payload)
	if payload != string(retorno) {
		t.Fail()
	}
}
