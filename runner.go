package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
	"io/ioutil"
)

const (
	//ideally we should pump this in as an env var or whatever but you know i am lazy for that crap
	sqsQueueURL    = "https://sqs.ap-southeast-2.amazonaws.com/989900959400/golangleetcode.fifo"
	existingGoCode = `
package main

import (
	"reflect"
	"testing"
)

func TestTwoSum(t *testing.T) {
	// Test cases and assertions
	nums := []int{2, 7, 11, 15}
	target := 9
	expected := []int{0, 1}
	result := twoSum(nums, target)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test failed, expected %v, got %v", expected, result)
	}

	// Additional test cases
	nums2 := []int{3, 2, 4}
	target2 := 6
	expected2 := []int{1, 2}
	result2 := twoSum(nums2, target2)
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Test failed, expected %v, got %v", expected2, result2)
	}

	nums3 := []int{3, 3}
	target3 := 6
	expected3 := []int{0, 1}
	result3 := twoSum(nums3, target3)
	if !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Test failed, expected %v, got %v", expected3, result3)
	}

	// Add more test cases as needed...
}


`
)

//
//// Define a struct to match the JSON structure
//type MessageBody struct {
//	CodeSnippet string `json:"codeSnippet"`
//}

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	}))
	sqsClient := sqs.New(sess)

	for {
		// Poll the SQS queue for messages
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsQueueURL),
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(20),
		})
		if err != nil {
			log.Printf("Error receiving message: %v\n", err)
			continue
		}

		if len(result.Messages) == 0 {
			continue
		}

		// Process each message
		for _, message := range result.Messages {
			fmt.Printf("Received message: %s\n", *message.Body)

			// Unmarshal the JSON data
			var msgBody MessageBody
			err := json.Unmarshal([]byte(*message.Body), &msgBody)
			if err != nil {
				log.Printf("Error parsing JSON message: %v\n", err)
				continue
			}

			// Combine user's code with existing code
			combinedCode := existingGoCode + msgBody.CodeSnippet

			// Create a temporary Go source file
			tempFile := fmt.Sprintf("./usercode_%s.go", uuid.New())
			err = ioutil.WriteFile(tempFile, []byte(combinedCode), 0644)
			if err != nil {
				log.Printf("Error creating temporary Go file: %v\n", err)
				continue
			}

			// Create and run a Docker container
			output, err := executeInDocker(tempFile)
			if err != nil {
				log.Printf("Error executing user code in Docker: %v\n", err)
				// Handle error...
			} else {
				fmt.Printf("User code output: %s\n", output)
				// Check if the test results indicate success
				if strings.Contains(output, "FAIL") {
					fmt.Println("User code test failed")
					// Handle test failure...
				} else {
					fmt.Println("User code test passed")
					// Send output to appropriate destination...
				}
			}

			// Delete the message from the queue
			_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(sqsQueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Error deleting message: %v\n", err)
				// Handle error...
			}

			// Clean up the temporary Go source file
			err = os.Remove(tempFile)
			if err != nil {
				log.Printf("Error deleting temporary Go file: %v\n", err)
				// Handle error...
			}
		}

		time.Sleep(1 * time.Second)
	}
}
