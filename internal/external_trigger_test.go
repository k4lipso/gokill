package internal

import (
	"fmt"
	"testing"
	"time"
)

type ExternalTriggerMapTest struct {
	secret                string
	testSecret            string
	receivedSecret        string
	expectedRegisterError error
	expectedError         error
	expectedTest          bool
}

func TestExternalTriggerMap(t *testing.T) {
	testRegisterTriggers := []ExternalTriggerMapTest{
		{
			secret:                "foo",
			testSecret:            "bar",
			receivedSecret:        "foo",
			expectedRegisterError: nil,
			expectedError:         nil,
			expectedTest:          false,
		},
		{
			secret:                "foo",
			testSecret:            "bar",
			receivedSecret:        "bar",
			expectedRegisterError: nil,
			expectedError:         nil,
			expectedTest:          true,
		},
		{
			secret:                "foo",
			testSecret:            "bar",
			receivedSecret:        "baz",
			expectedRegisterError: nil,
			expectedError:         fmt.Errorf("Some Error"),
		},
		{
			secret:                "1234",
			testSecret:            "bar",
			receivedSecret:        "1234",
			expectedRegisterError: nil,
			expectedError:         nil,
			expectedTest:          false,
		},
		{
			secret:                "",
			testSecret:            "",
			receivedSecret:        "",
			expectedRegisterError: fmt.Errorf("Some Error"),
			expectedError:         nil,
			expectedTest:          false,
		},
	}

	triggerMap := make(map[string]TriggerChannel)
	externalTriggerMap := ExternalTriggerMap{
		TriggerChannels: triggerMap,
	}
	//we expect test to finish within 10 seconds, otherwise deadlock expected
	timeout := time.After(10 * time.Second)
	done := make(chan bool)
	go func() {
		for _, testCase := range testRegisterTriggers {
			channel, err := externalTriggerMap.RegisterRemoteTrigger(testCase.secret, testCase.testSecret)

			if (err == nil) != (testCase.expectedRegisterError == nil) {
				t.Errorf("Expected RegisterError: %s, got: %s", testCase.expectedRegisterError, err)
				continue
			}

			if testCase.expectedRegisterError != nil {
				continue
			}

			channel2, _ := externalTriggerMap.RegisterRemoteTrigger(testCase.secret, testCase.testSecret)
			if channel != channel2 {
				t.Error("Registering a trigger twice returnes different channels. This should not happen!")
				continue
			}

			go func() {
				err = externalTriggerMap.ExecuteRemoteTrigger(TriggerEvent{Secret: testCase.receivedSecret})

				if (err == nil) != (testCase.expectedError == nil) {
					t.Errorf("Expected Error: %s, got: %s", testCase.expectedError, err)
				}
			}()

			if testCase.expectedError != nil {
				continue
			}

			receivedTrigger := <-channel

			if receivedTrigger.IsTest != testCase.expectedTest {
				t.Errorf("Expected Trigger: %t, got: %t", testCase.expectedTest, receivedTrigger.IsTest)
				continue
			}
		}
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("TestExternalTriggerMap timed out.")
	case <-done:
	}
}
