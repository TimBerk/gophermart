package client

import (
	model "TimBerk/gophermart/internal/app/models/order"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

const registerURI string = "/api/orders"
const checkOrderURI string = "/api/orders/"

type Client struct {
	url string
}

func NewClient(baseURL string) *Client {
	return &Client{url: baseURL}
}

func (c Client) getFullPath(path string) (string, error) {
	parsedURL, err := url.Parse(c.url)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": "C.NewClient", "error": err}).Error("failed to parse URL")
		return "", err
	}

	parsedURL.Path = path
	return parsedURL.String(), nil
}

func (c Client) Register(order string) (int, error) {
	action := "C.Register"

	fullPath, err := c.getFullPath(registerURI)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "order": order, "error": err}).Error("failed to build path")
		return http.StatusInternalServerError, err
	}

	requstData := model.OrderAccrualRegister{Number: order}
	payload, err := json.Marshal(requstData)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "order": order, "error": err}).Error("failed to parse request body")
		return http.StatusInternalServerError, err
	}

	buffer := bytes.NewBuffer(payload)

	resp, err := http.Post(fullPath, "application/json", buffer)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "order": order, "error": err}).Error("failed to send request")
		return http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	logrus.WithFields(logrus.Fields{"action": action, "order": order, "response": resp, "body": resp.Body}).Info("get info about order")
	return resp.StatusCode, nil
}

func (c Client) GetStatus(order string) (string, error) {
	action := "C.GetStatus"

	fullPath, err := c.getFullPath(checkOrderURI + order)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": "C.Register", "order": order, "error": err}).Error("failed to build path")
		return "", err
	}

	resp, err := http.Get(fullPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "order": order, "error": err}).Error("failed to send request")
		return "", err
	}
	defer resp.Body.Close()

	logrus.WithFields(logrus.Fields{"action": action, "order": order, "response": resp, "body": resp.Body}).Info("get info about order")

	if resp.StatusCode > 202 {
		return "", fmt.Errorf("order not ready")
	}

	decoder := json.NewDecoder(resp.Body)

	var orderAccrual model.OrderAccrual
	err = decoder.Decode(&orderAccrual)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "order": order, "error": err}).Error("failed to parse url")
		return "", err
	}

	return orderAccrual.Status, nil
}
