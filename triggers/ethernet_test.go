package triggers

import (
	"testing"

	"unknown.com/gokill/internal"
)

func TestEthernetDisconnetConfig(t *testing.T) {
	type EthernetTest struct {
		testConfig     internal.KillSwitchConfig
		expectedError  error
		expectedResult EthernetDisconnect
	}

	testConfigs := []EthernetTest{
		EthernetTest{
			testConfig: internal.KillSwitchConfig{
				Options: []byte("{}"),
			},

			expectedError:  internal.OptionMissingError{"interfaceName"},
			expectedResult: EthernetDisconnect{},
		},
		EthernetTest{
			testConfig: internal.KillSwitchConfig{
				Options: []byte(`{ 
					"waitTillConnected": false
				}`),
			},

			expectedError:  internal.OptionMissingError{"interfaceName"},
			expectedResult: EthernetDisconnect{},
		},
		EthernetTest{
			testConfig: internal.KillSwitchConfig{
				Options: []byte(`{ 
					"interfaceName": "eth0",
					"waitTillConnected": false
				}`),
			},

			expectedError:  nil,
			expectedResult: EthernetDisconnect{WaitTillConnected: false, InterfaceName: "eth0"},
		},
		EthernetTest{
			testConfig: internal.KillSwitchConfig{
				Options: []byte(`{ 
					"interfaceName": "eth0",
					"waitTillConnected": true
				}`),
			},

			expectedError:  nil,
			expectedResult: EthernetDisconnect{WaitTillConnected: true, InterfaceName: "eth0"},
		},
		EthernetTest{
			testConfig: internal.KillSwitchConfig{
				Options: []byte("{ \"interfaceName\": \"eth0\" }"),
			},

			expectedError:  nil,
			expectedResult: EthernetDisconnect{WaitTillConnected: true, InterfaceName: "eth0"},
		},
	}

	for _, testConfig := range testConfigs {
		result, err := NewEthernetDisconnect(testConfig.testConfig)

		if err != testConfig.expectedError {
			t.Errorf("Error was incorrect, got: %s, want: %s.", err, testConfig.expectedError)
		}

		if result.WaitTillConnected != testConfig.expectedResult.WaitTillConnected {
			t.Errorf("WaitTillConnected was incorrect, got: %v, want: %v.", result, testConfig.expectedResult)
		}

		if result.InterfaceName != testConfig.expectedResult.InterfaceName {
			t.Errorf("InterfaceName was incorrect, got: %v, want: %v.", result, testConfig.expectedResult)
		}
	}
}
