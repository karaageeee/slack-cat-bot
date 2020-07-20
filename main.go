package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Request is a struct of Slack request
type Request struct {
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
	Event     struct {
		Channel string `json:"channel"`
		Blocks  []struct {
			Elements []struct {
				Elements []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"elements"`
			} `json:"elements"`
		} `json:"blocks"`
	} `json:"event"`
}

// ChallengeResponse is a struct of challenge response
type ChallengeResponse struct {
	Challenge string `json:"challenge"`
}

// MessageRequest is a struct of Slack message request
type MessageRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// SlackMessageEndpoint is endpoint of slack API
const SlackMessageEndpoint = "https://slack.com/api/chat.postMessage"

// CatImgURL is to get cat images
const CatImgURL = "https://http.cat/"

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "にゃ〜ん")
	})

	router.POST("/bot", func(c *gin.Context) {

		var body Request
		c.BindJSON(&body)

		if body.Type == "url_verification" {
			res := ChallengeResponse{Challenge: body.Challenge}
			c.JSON(http.StatusOK, res)
		} else if body.Type == "event_callback" {

			if body.Event.Channel == "" {
				log.Fatal("Channel is empty")
				return
			}

			if len(body.Event.Blocks) == 0 ||
				len(body.Event.Blocks[0].Elements) == 0 ||
				len(body.Event.Blocks[0].Elements[0].Elements) == 0 {
				sendMessage(body.Event.Channel, "meow(SYSTEM ERROR)")
				return
			}

			var recievedMessage string
			for _, elem := range body.Event.Blocks[0].Elements[0].Elements {
				if elem.Type == "text" {
					recievedMessage = strings.TrimSpace(elem.Text)
					break
				}
			}

			if isValidHTTPStatusCode(recievedMessage) {
				sendMessage(body.Event.Channel, CatImgURL+recievedMessage)
			} else {
				sendMessage(body.Event.Channel, "にゃ〜ん")
			}
		}
	})

	router.Run(":" + port)
}

func sendMessage(channel, message string) error {

	messageReq := MessageRequest{
		Channel: channel,
		Text:    message,
	}
	messageReqJSON, _ := json.Marshal(messageReq)

	req, reqErr := http.NewRequest(
		http.MethodPost,
		SlackMessageEndpoint,
		bytes.NewBuffer(messageReqJSON),
	)
	if reqErr != nil {
		log.Fatal(reqErr)
		return reqErr
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SLACK_OAUTH_TOKEN"))

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Fatal(respErr)
		return respErr
	}

	defer resp.Body.Close()

	return nil
}

func containsInt(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func isValidHTTPStatusCode(code string) bool {

	// list of status on https://http.cat/
	httpStatusCodes := []int{
		100, 101,
		200, 201, 202, 204, 206, 207,
		300, 301, 302, 303, 304, 305, 307,
		400, 401, 402, 403, 404, 405, 406, 408, 409,
		410, 411, 412, 413, 414, 415, 416, 417, 418,
		420, 421, 422, 423, 424, 425, 426, 429, 431,
		444, 450, 451, 499,
		500, 501, 502, 503, 504, 505, 506, 507, 508, 509,
		510, 511, 599}

	httpStatusCode, strconvErr := strconv.Atoi(code)

	if strconvErr == nil && containsInt(httpStatusCode, httpStatusCodes) {
		return true
	}

	return false
}
