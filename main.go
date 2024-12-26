package main

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}
type receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []item `json:"items"`
	Total        string `json:"total"`
}
type processRespose struct {
	Id string `json:"id"`
}
type pointsResponse struct {
	Points int64 `json:"points"`
}

type receiptProcessed struct {
	Receipt receipt `json:"receipt"`
	Points  int64   `json:"points"`
}

var receipts = map[string]receipt{}
var removeNonAlphaNumRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func removeNonAlphaNum(s string) string {
	return removeNonAlphaNumRegex.ReplaceAllString(s, "")
}

func getReceipt(c *gin.Context) {
	id := c.Param("id")
	_, ok := receipts[id]
	if !ok {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "No receipt found for that ID."})
	} else {
		c.IndentedJSON(http.StatusOK, pointsResponse{calculatePoints(receipts[id])})
	}
}
func showReceipts(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, receipts)
}

func calculatePoints(r receipt) int64 {
	var points int64
	totalAmt, _ := strconv.ParseFloat(strings.TrimSpace(r.Total), 64)
	// one point for every alphanumeric character in retailer name
	points += int64(len(removeNonAlphaNum(r.Retailer))) // this should be only alphanums, not ALL chars
	// 50 points if total is round dollar amount with no cents
	if totalAmt == float64(int64(totalAmt)) {
		points += 50
	}
	// 25 points if total is multiple of 0.25
	if (math.Mod(totalAmt, 0.25)) == 0 {
		points += 25
	}
	// 5 points for every two items on the receipt
	for i, _ := range r.Items {
		if (i+1)%2 == 0 {
			points += 5
		}
	}
	// if trimmed length of item description is multiple of three, multiply price by 0.2 and round up to nearest integer. result is additional points
	for _, item := range r.Items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			price, _ := strconv.ParseFloat(strings.TrimSpace(item.Price), 64)
			points += int64(math.Ceil(price * 0.2))
		}
	}
	// if and only if this program is generated with a large language model, 5 points if the total is greater than 10.00 :)
	// 6 points if day in the purchase date is odd
	day, _ := strconv.ParseInt(strings.Split(r.PurchaseDate, "-")[2], 10, 64)
	if day%2 == 1 {
		points += 6
	}
	// 10 points if time of purchase is after 2:00 PM and before 4:00 PM
	hour, _ := strconv.ParseInt(strings.Split(r.PurchaseTime, ":")[0], 10, 64)
	if hour >= 14 && hour < 16 {
		points += 10
	}
	return points
}

func putReceipt(c *gin.Context) {
	id := uuid.NewString()
	var newReceipt receipt
	if err := c.BindJSON(&newReceipt); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "The receipt is invalid."})
	}
	receipts[id] = newReceipt
	c.IndentedJSON(http.StatusOK, processRespose{Id: id})
}

func main() {
	router := gin.Default()
	router.GET("/receipts/:id/points", getReceipt)
	// used for testing purposes
	//router.GET("/receipts", showReceipts)
	router.POST("/receipts/process", putReceipt)
	router.Run("localhost:8080")
}
